package lib

import "errors"

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

func (partition *Partition) Check() []ManifestMismatch {
	panic(errors.ErrUnsupported)
}
