package lib

import (
	"context"
	"errors"
	"io/fs"

	"github.com/azerum/part/utils"
)

func (partition *Partition) Check(ctx context.Context) *utils.ChanWithError[ManifestMismatch] {
	out := utils.NewChanWithError[ManifestMismatch](1)
	go checkWorker(partition, out, ctx)

	return out
}

func checkWorker(
	partition *Partition,
	out *utils.ChanWithError[ManifestMismatch],
	ctx context.Context,
) {
	if partition.manifest == nil {
		out.CloseWithError(errors.New("partition has no manifest"))
		return
	}

	seenInPartition := make(map[string]struct{})

	walk := func(absoluteOsPath string, manifestPath string, _entry fs.DirEntry) error {
		err := ctx.Err()

		if err != nil {
			return err
		}

		seenInPartition[manifestPath] = struct{}{}

		manifestEntry := partition.manifest.Files[manifestPath]

		if manifestEntry == nil {
			out.Channel <- FileNotHashed{ManifestPath: manifestPath}
			return nil
		}

		hash, err := HashFile(absoluteOsPath)

		if err != nil {
			return err
		}

		if hash != manifestEntry.Hash {
			out.Channel <- HashDoesNotMatch{
				ManifestPath: manifestPath,
				ActualHash:   hash,
				ExpectedHash: manifestEntry.Hash,
			}
		}

		return nil
	}

	err := partition.Walk(walk, ctx)

	if err != nil {
		out.CloseWithError(err)
		return
	}

	for p := range partition.manifest.Files {
		_, seen := seenInPartition[p]

		if !seen {
			out.Channel <- FileMissing{ManifestPath: p}
		}
	}

	out.CloseOk()
}

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
