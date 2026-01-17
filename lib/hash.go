package lib

import (
	"errors"
	"fmt"
)

type ManifestChange interface {
	apply(manifest *manifest) error
}

type FileAdded struct {
	Path  string
	hash  string
	mtime int64
}

func (c FileAdded) apply(manifest *manifest) error {
	_, exists := manifest.Files[c.Path]

	if exists {
		return fmt.Errorf(
			"cannot apply FileAdded: file %s already exists in manifest",
			c.Path,
		)
	}

	manifest.Files[c.Path] = &fileEntry{
		hash:  c.hash,
		mtime: c.mtime,
	}

	return nil
}

type FileModified struct {
	Path  string
	hash  string
	mtime int64
}

func (c FileModified) apply(manifest *manifest) error {
	entry, exists := manifest.Files[c.Path]

	if !exists {
		return fmt.Errorf(
			"cannot apply FileModified: file %s does not exist in manifest",
			c.Path,
		)
	}

	entry.hash = c.hash
	entry.mtime = c.mtime

	return nil
}

type FileRemoved struct {
	Path string
}

func (c FileRemoved) apply(manifest *manifest) error {
	_, exists := manifest.Files[c.Path]

	if !exists {
		return fmt.Errorf(
			"cannot apply FileRemoved: file %s does not exist in manifest",
			c.Path,
		)
	}

	delete(manifest.Files, c.Path)
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

type ManifestMismatch interface {
	isManifestMismatch()
}

type FileMissing struct {
	Path string
}

func (m FileMissing) isManifestMismatch() {}

type FileNotHashed struct {
	Path string
}

func (m FileNotHashed) isManifestMismatch() {}

type HashDoesNotMatch struct {
	Path         string
	ActualHash   string
	ExpectedHash string
}

func (m HashDoesNotMatch) isManifestMismatch() {}

func (partition *Partition) Hash() ([]ManifestChange, error) {

}
