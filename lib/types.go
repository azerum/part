package lib

type Partition struct {
	DirPath string

	// nil if this partition has not been hashed yet, i.e. when it contains
	// no manifest file
	manifest *manifest
}

type manifest struct {
	// Maps file path relative to partition to info about the file
	//
	// Allowed to be either `{}` or `null` in JSON - both are interpreted
	// as "there were no files in the partition directory at the moment of hashing"
	Files map[string]*fileEntry `json:"files"`
}

type fileEntry struct {
	hash  string
	mtime int64
}
