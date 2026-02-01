package partition_lib

import (
	"errors"
	"os"
	"path/filepath"
)

const manifestFileName = ".manifest.json"
const manifestTmpFileName = manifestFileName + ".tmp"

func LoadPartition(dirPath string) (*Partition, error) {
	manifestBytes, err := os.ReadFile(
		filepath.Join(dirPath, manifestFileName),
	)

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		p := Partition{
			AbsoluteDirOsPath: dirPath,
			manifest:          nil,
		}

		return &p, nil
	}

	return DeserializePartition(dirPath, manifestBytes)
}

func (partition *Partition) Save() error {
	manifestBytes, err := partition.Serialize()

	if err != nil {
		return err
	}

	manifestPath := filepath.Join(partition.AbsoluteDirOsPath, manifestFileName)
	manifestTmpPath := filepath.Join(partition.AbsoluteDirOsPath, manifestTmpFileName)

	return overwrite(manifestPath, manifestTmpPath, manifestBytes)
}

// All paths must be absolute
//
// Given assumptions:
//
// A1.  For rename(A, B), if both A and B exist, it is guaranteed that at any
// point of time, either B has complete, uncorrupted contents it had
// before rename(), or it has complete, uncorrupted contents A had
//
// A2.  After fsync(f); close(f) succeed, all writes done by this process to f after
// last fsync() (or  after opening f if there was no fsync()) are persisted, durably
//
// A3. A2 applies persisting directories info when we rename files inside
//
// Guarantees two properties:
//
// P1:
//
// At any point of time, either `filePath` has complete, uncorrupted
// contents as before the call to overwrite(), or has complete, uncorrupted
// contents with `data` - but nothing in-between, partial, corrupted
//
// P2:
//
// Once it returns, `filePath` surely contains complete, uncorrupted `data`,
// persisted on disk
func overwrite(filePath string, tmpPath string, data []byte) error {
	// Proof of P1:
	//
	// We first write to tmpPath, then rename(tmpPath, filePath). This can
	// result in filePath being corrupted only if:
	//
	// 1. We did rename() before tmpPath was fully written
	// 2. rename() may leave filePath corrupted even if tmpPath is not
	//
	// For 1: we do fsync(); close() on tmpPath - prevented by A2
	// For 2: prevented by A1

	// Proof of P2:
	//
	// We may return without error even though filePath is not completely
	// written if:
	//
	// 1. We proceed before tmpPath is complete (prevented by A2, se above)
	// 2. We proceed before effect rename() is persisted
	//
	// For 1: prevented by A2, see above
	// For 2: we do fsync(); close() on the directory - prevented by A3

	tmpFile, err := os.Create(tmpPath)

	if err != nil {
		return err
	}

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)

		return err
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)

		return err
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, filePath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	dirPath := filepath.Dir(filePath)
	dir, err := os.Open(dirPath)

	if err != nil {
		return err
	}

	if err := dir.Sync(); err != nil {
		_ = dir.Close()
		return err
	}

	return dir.Close()
}
