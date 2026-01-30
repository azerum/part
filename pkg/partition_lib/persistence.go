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
// Once it returns, `filePath surely contains complete, uncorrupted `data`,
// persisted on disk
//
// The properties hold if anything other than the disk (this program, OS, computer)
// crashes
//
// TODO: check if this works for Windows!!!
func overwrite(filePath string, tmpPath string, data []byte) error {
	// General idea:
	//
	// To achieve P1: write to temporary file, then rename it to the original
	// file (replacing the original file)
	//
	// This relies on sub-property SP1:
	//
	// For rename(A, B), if both A and B exist, it is guaranteed that at any
	// point of time, either B has complete, uncorrupted contents it had
	// before rename(), or it has complete, uncorrupted contents A had
	//
	// To achieve P2: fsync() and close() all files we write to, fsync()
	// and close() directory(-ies) where rename() takes place
	//
	// This relies on sub-properties SP2 and SP3:
	//
	// SP2: After fsync(f); close(f) succeed, all writes done to f before that
	// are persisted on durable disk
	//
	// SP3: SP2 applies to directories when we rename files inside

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
