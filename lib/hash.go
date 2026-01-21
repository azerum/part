package lib

import (
	"errors"
	"fmt"
	"io/fs"
)

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
		hash:  c.hash,
		mtime: c.mtime,
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

	entry.hash = c.hash
	entry.mtime = c.mtime

	return nil
}

type FileRemoved struct {
	ManifestPath string
}

func (c FileRemoved) apply(manifest *manifest) error {
	_, exists := manifest.Files[c.ManifestPath]

	if !exists {
		return fmt.Errorf(
			"cannot apply FileRemoved: file %s does not exist in manifest",
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

	entry.mtime = c.mtime
	return nil
}

// Passed changes are expected to make sense, i.e. to hold invariants:
//
// - FileAdded never asks to add already added file
// - FileModified & FileRemoved never refer to not-added file
//
// **panics** if invariants are violated
func (partition *Partition) ApplyChanges(changes []ManifestChange) {
	for index, change := range changes {
		err := change.apply(partition.manifest)

		if err == nil {
			continue
		}

		fullErr := errors.Join(
			fmt.Errorf("invariant violated while applying change index=%d", index),
			err,
		)

		panic(fullErr)
	}
}

func (partition *Partition) Hash() ([]ManifestChange, error) {
	changes := make([]ManifestChange, 0)
	seenInPartition := make(map[string]struct{})

	if partition.manifest == nil {
		partition.manifest = &manifest{
			Files: make(map[string]*fileEntry),
		}
	}

	err := partition.Walk(func(absoluteOsPath string, manifestPath string, entry fs.DirEntry) error {
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

			changes = append(changes, FileAdded{
				ManifestPath: manifestPath,
				hash:         hash,
				mtime:        mtime,
			})

			return nil
		}

		manifestEntry := partition.manifest.Files[manifestPath]

		if manifestEntry == nil {
			hash, err := HashFile(absoluteOsPath)

			if err != nil {
				return err
			}

			changes = append(changes, FileAdded{
				ManifestPath: manifestPath,
				hash:         hash,
				mtime:        mtime,
			})

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

		if manifestEntry.mtime == mtime {
			return nil
		}

		hash, err := HashFile(absoluteOsPath)

		if err != nil {
			return err
		}

		if hash == manifestEntry.hash {
			changes = append(changes, SpuriousMtimeChange{
				ManifestPath: manifestPath,
				mtime:        mtime,
			})

			return nil
		}

		changes = append(changes, FileModified{
			ManifestPath: manifestPath,
			hash:         hash,
			mtime:        mtime,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Files that were not seen in the partition but are in the manifest
	// are the files that were removed
	for p := range partition.manifest.Files {
		_, seen := seenInPartition[p]

		if !seen {
			changes = append(changes, FileRemoved{ManifestPath: p})
		}
	}

	return changes, nil
}
