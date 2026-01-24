package lib_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/azerum/part/lib"
)

func setupTestPartition(t *testing.T) *lib.Partition {
	dirPath := t.TempDir()

	// Create directory
	//
	// dir/
	//  a (contains A)
	//  b (contains B)
	//  c/
	//    d (contains D)
	//  e (contains E)

	if err := os.WriteFile(filepath.Join(dirPath, "a"), ([]byte)("A"), 0o600); err != nil {
		panic(err)
	}

	if err := os.WriteFile(filepath.Join(dirPath, "b"), ([]byte)("B"), 0o600); err != nil {
		panic(err)
	}

	if err := os.Mkdir(filepath.Join(dirPath, "c"), 0o700); err != nil {
		panic(err)
	}

	if err := os.WriteFile(filepath.Join(dirPath, "c", "d"), ([]byte)("D"), 0o600); err != nil {
		panic(err)
	}

	if err := os.WriteFile(filepath.Join(dirPath, "e"), ([]byte)("E"), 0o600); err != nil {
		panic(err)
	}

	partition, err := lib.LoadPartition(dirPath)

	if err != nil {
		panic(err)
	}

	return partition
}

func hashAndSave(partition *lib.Partition) {
	changes := partition.Hash(context.Background())

	for c := range changes.Channel {
		partition.ApplyChange(c)
	}

	if changes.Err != nil {
		panic(changes.Err)
	}

	if err := partition.Save(); err != nil {
		panic(err)
	}
}

func addFileF(partition *lib.Partition) {
	if err := os.WriteFile(filepath.Join(partition.AbsoluteDirOsPath, "f"), ([]byte)("F"), 0o600); err != nil {
		panic(err)
	}
}

func removeFileBAndDirectoryC(partition *lib.Partition) {
	if err := os.Remove(filepath.Join(partition.AbsoluteDirOsPath, "b")); err != nil {
		panic(err)
	}

	if err := os.RemoveAll(filepath.Join(partition.AbsoluteDirOsPath, "c")); err != nil {
		panic(err)
	}
}

func modifyFileA(partition *lib.Partition) {
	// Wait at least one second so mtime will be different even if
	// this FS has 1s resolution
	time.Sleep(time.Second)

	if err := os.WriteFile(filepath.Join(partition.AbsoluteDirOsPath, "a"), ([]byte)("A2"), 0o600); err != nil {
		panic(err)
	}
}

func modifyFileEMtime(partition *lib.Partition) {
	path := filepath.Join(partition.AbsoluteDirOsPath, "e")
	t := time.Now().Add(10 * time.Second)

	if err := os.Chtimes(path, t, t); err != nil {
		panic(err)
	}
}
