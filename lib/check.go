package lib

import "errors"

func (partition *Partition) Check() []ManifestChange {
	panic(errors.ErrUnsupported)
}
