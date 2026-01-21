package lib

import "errors"

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

func (partition *Partition) Check() []ManifestMismatch {
	panic(errors.ErrUnsupported)
}
