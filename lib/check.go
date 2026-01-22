package lib

import (
	"errors"
	"io/fs"
)

type ManifestMismatch interface {
	isManifestMismatch()
}

type FileMissing struct {
	ManifestPath string
}

func (m FileMissing) isManifestMismatch() {}

type FileNotHashed struct {
	ManifestPath string
}

func (m FileNotHashed) isManifestMismatch() {}

type HashDoesNotMatch struct {
	ManifestPath string
	ActualHash   string
	ExpectedHash string
}

func (m HashDoesNotMatch) isManifestMismatch() {}

func (partition *Partition) Check() ([]ManifestMismatch, error) {
	if partition.manifest == nil {
		return nil, errors.New("partition has no manifest")
	}

	mismatches := make([]ManifestMismatch, 0)
	seenInPartition := make(map[string]struct{})

	err := partition.Walk(func(absoluteOsPath string, manifestPath string, _entry fs.DirEntry) error {
		seenInPartition[manifestPath] = struct{}{}

		manifestEntry := partition.manifest.Files[manifestPath]

		if manifestEntry == nil {
			mismatches = append(mismatches, FileNotHashed{ManifestPath: manifestPath})
			return nil
		}

		hash, err := HashFile(absoluteOsPath)

		if err != nil {
			return err
		}

		if hash != manifestEntry.Hash {
			mismatches = append(mismatches, HashDoesNotMatch{
				ManifestPath: manifestPath,
				ActualHash:   hash,
				ExpectedHash: manifestEntry.Hash,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for p := range partition.manifest.Files {
		_, seen := seenInPartition[p]

		if !seen {
			mismatches = append(mismatches, FileMissing{ManifestPath: p})
		}
	}

	return mismatches, nil
}
