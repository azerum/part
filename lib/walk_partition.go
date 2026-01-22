package lib

import (
	"context"
	"io/fs"
	"path/filepath"
)

type WalkPartitionCallback func(absoluteOsPath string, manifestPath string, entry fs.DirEntry) error

func (partition *Partition) Walk(callback WalkPartitionCallback, ctx context.Context) error {
	pathsToIgnore := make(map[string]struct{})

	pathsToIgnore[filepath.Join(partition.AbsoluteDirOsPath, manifestFileName)] = struct{}{}
	pathsToIgnore[filepath.Join(partition.AbsoluteDirOsPath, manifestTmpFileName)] = struct{}{}

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

		if _, exists := pathsToIgnore[path]; exists {
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
