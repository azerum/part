package partition_lib

import (
	"context"
	"io/fs"
	"path/filepath"
)

type WalkPartitionCallback func(absoluteOsPath string, manifestPath string, entry fs.DirEntry) error

func (partition *Partition) Walk(callback WalkPartitionCallback, ctx context.Context) error {
	topLevelManifestPath := filepath.Join(partition.AbsoluteDirOsPath, manifestFileName)
	topLevelTmpManifestPath := filepath.Join(partition.AbsoluteDirOsPath, manifestTmpFileName)

	shouldIgnore := func(path string) bool {
		if path == topLevelManifestPath || path == topLevelTmpManifestPath {
			return true
		}

		fileName := filepath.Base(path)
		return fileName == ".DS_Store"
	}

	return filepath.WalkDir(partition.AbsoluteDirOsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if shouldIgnore(path) {
			return nil
		}

		manifestPath, err := toManifestPath(partition.AbsoluteDirOsPath, path)

		if err != nil {
			return err
		}

		if err := callback(path, manifestPath, d); err != nil {
			return err
		}

		return nil
	})
}
