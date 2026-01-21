package lib

import "path/filepath"

type Partition struct {
	AbsoluteDirOsPath string

	// nil if this partition has not been hashed yet, i.e. when it contains
	// no manifest file
	manifest *manifest
}

type manifest struct {
	// Maps "manifest path" of file to its info. "Manifest path" is created
	// by toManifestPath() function, see the comments on it
	//
	// Allowed to be either `{}` or `null` in JSON - both are interpreted
	// as "there were no files in the partition directory at the moment of hashing"
	Files map[string]*fileEntry `json:"files"`
}

type fileEntry struct {
	hash  string
	mtime int64
}

func toManifestPath(partitionDirAbsoluteOsPath string, absolutePath string) (string, error) {
	p, err := filepath.Rel(partitionDirAbsoluteOsPath, absolutePath)

	if err != nil {
		return "", err
	}

	return filepath.ToSlash(p), nil
}
