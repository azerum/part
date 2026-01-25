package partition_lib

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/azerum/part/utils"
)

func (partition *Partition) Hash(ctx context.Context) *utils.ChanWithError[ManifestChange] {
	out := utils.NewChanWithError[ManifestChange](1)
	go hashWorker(partition, out, ctx)

	return out
}

func hashWorker(
	partition *Partition,
	out *utils.ChanWithError[ManifestChange],
	ctx context.Context,
) {
	seenInPartition := make(map[string]struct{})

	if partition.manifest == nil {
		partition.manifest = &manifest{
			Files: make(map[string]*fileEntry),
		}
	}

	walk := func(absoluteOsPath string, manifestPath string, entry fs.DirEntry) error {
		seenInPartition[manifestPath] = struct{}{}

		info, err := entry.Info()

		if err != nil {
			return err
		}

		mtime := info.ModTime().Unix()

		// If we have no manifest yet, everything is added
		if partition.manifest == nil {
			hash, err := HashFile(absoluteOsPath)

			if err != nil {
				return err
			}

			out.Channel <- FileAdded{
				ManifestPath: manifestPath,
				hash:         hash,
				mtime:        mtime,
			}

			return nil
		}

		manifestEntry := partition.manifest.Files[manifestPath]

		if manifestEntry == nil {
			hash, err := HashFile(absoluteOsPath)

			if err != nil {
				return err
			}

			out.Channel <- FileAdded{
				ManifestPath: manifestPath,
				hash:         hash,
				mtime:        mtime,
			}

			return nil
		}

		// If file's mtime is the same as in the manifest, assume it has
		// not changed. Avoid hashing file until this check, as reading files
		// is slow
		//
		// If mtime did change, verify if the contents has changed using hashes
		//
		// It is possible that mtime changed and hash didn't - we should update
		// mtime in the manifest in such case, to avoid hashing this file
		// next time

		if manifestEntry.Mtime == mtime {
			return nil
		}

		hash, err := HashFile(absoluteOsPath)

		if err != nil {
			return err
		}

		if hash == manifestEntry.Hash {
			out.Channel <- SpuriousMtimeChange{
				ManifestPath: manifestPath,
				mtime:        mtime,
			}

			return nil
		}

		out.Channel <- FileModified{
			ManifestPath: manifestPath,
			hash:         hash,
			mtime:        mtime,
		}

		return nil
	}

	err := partition.Walk(walk, ctx)

	if err != nil {
		out.CloseWithError(err)
		return
	}

	// Files that were not seen in the partition but are in the manifest
	// are the files that were deleted
	for p := range partition.manifest.Files {
		_, seen := seenInPartition[p]

		if !seen {
			out.Channel <- FileDeleted{ManifestPath: p}
		}
	}

	out.CloseOk()

}

type ManifestChange interface {
	apply(manifest *manifest) error
}

type FileAdded struct {
	ManifestPath string
	hash         string
	mtime        int64
}

func (c FileAdded) apply(manifest *manifest) error {
	_, exists := manifest.Files[c.ManifestPath]

	if exists {
		return fmt.Errorf(
			"cannot apply FileAdded: file %s already exists in manifest",
			c.ManifestPath,
		)
	}

	manifest.Files[c.ManifestPath] = &fileEntry{
		Hash:  c.hash,
		Mtime: c.mtime,
	}

	return nil
}

type FileModified struct {
	ManifestPath string
	hash         string
	mtime        int64
}

func (c FileModified) apply(manifest *manifest) error {
	entry, exists := manifest.Files[c.ManifestPath]

	if !exists {
		return fmt.Errorf(
			"cannot apply FileModified: file %s does not exist in manifest",
			c.ManifestPath,
		)
	}

	entry.Hash = c.hash
	entry.Mtime = c.mtime

	return nil
}

type FileDeleted struct {
	ManifestPath string
}

func (c FileDeleted) apply(manifest *manifest) error {
	_, exists := manifest.Files[c.ManifestPath]

	if !exists {
		return fmt.Errorf(
			"cannot apply FileDeleted: file %s does not exist in manifest",
			c.ManifestPath,
		)
	}

	delete(manifest.Files, c.ManifestPath)
	return nil
}

type SpuriousMtimeChange struct {
	ManifestPath string
	mtime        int64
}

func (c SpuriousMtimeChange) apply(manifest *manifest) error {
	entry, exists := manifest.Files[c.ManifestPath]

	if !exists {
		return fmt.Errorf(
			"cannot apply SpuriousMtimeChange: file %s does not exist in manifest",
			c.ManifestPath,
		)
	}

	entry.Mtime = c.mtime
	return nil
}

// Passed changes are expected to make sense, i.e. to hold invariants:
//
// - FileAdded never asks to add already added file
// - FileModified & FileDeleted never refer to not-added file
//
// **panics** if invariants are violated
func (partition *Partition) ApplyChange(change ManifestChange) {
	err := change.apply(partition.manifest)

	if err == nil {
		return
	}

	fullErr := errors.Join(
		fmt.Errorf("invariant violated while applying the change"),
		err,
	)

	panic(fullErr)
}
