package partition_lib

import (
	"context"
	"io/fs"
	"path/filepath"
)

type WalkPartitionCallback func(absoluteOsPath string, manifestPath string, entry fs.DirEntry) error

func (partition *Partition) Walk(callback WalkPartitionCallback, ctx context.Context) error {
	shouldIgnore := func(path string) bool {
		fileName := filepath.Base(path)

		_, exists := fileNamesToIgnore[fileName]
		return exists
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

var fileNamesToIgnore = map[string]struct{}{
	manifestFileName:    {},
	manifestTmpFileName: {},

	// macOS
	// Source: https://github.com/github/gitignore/blob/main/Global/macOS.gitignore
	//
	// Tweaks: removed `._*` and `Icon` as they seem way too generic (may affect
	// non-OS-specific files)

	".DS_Store":                           {},
	"._.DS_Store":                         {},
	"__MACOSX/":                           {},
	".AppleDouble":                        {},
	".LSOverride":                         {},
	".DocumentRevisions-V100":             {},
	".fseventsd":                          {},
	".Spotlight-V100":                     {},
	".TemporaryItems":                     {},
	".Trashes":                            {},
	".VolumeIcon.icns":                    {},
	".com.apple.timemachine.donotpresent": {},
	".AppleDB":                            {},
	".AppleDesktop":                       {},
	"Network Trash Folder":                {},
	"Temporary Items":                     {},
	".apdisk":                             {},
}
