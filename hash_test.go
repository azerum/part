package main_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/azerum/partition/lib"
	. "github.com/onsi/gomega"
	gs "github.com/onsi/gomega/gstruct"
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

	partition, err := lib.LoadPartition(dirPath)

	if err != nil {
		panic(err)
	}

	return partition
}

func hashAndSave(partition *lib.Partition) {
	changes, err := partition.Hash()

	if err != nil {
		panic(err)
	}

	partition.ApplyChanges(changes)

	if err := partition.Save(); err != nil {
		panic(err)
	}
}

func addFileE(partition *lib.Partition) {
	if err := os.WriteFile(filepath.Join(partition.AbsoluteDirOsPath, "e"), ([]byte)("E"), 0o600); err != nil {
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

func Test_creates_entire_manifest_when_run_for_the_first_time(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	changes, err := p.Hash()

	if err != nil {
		panic(err)
	}

	g.Expect(changes).To(HaveExactElements(
		SatisfyAll(
			BeAssignableToTypeOf(lib.FileAdded{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("a"),
			}),
		),

		SatisfyAll(
			BeAssignableToTypeOf(lib.FileAdded{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("b"),
			}),
		),

		SatisfyAll(
			BeAssignableToTypeOf(lib.FileAdded{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("c/d"),
			}),
		),
	))
}

func Test_detects_added_files(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	addFileE(p)
	changes, err := p.Hash()

	if err != nil {
		panic(err)
	}

	g.Expect(changes).To(HaveExactElements(
		SatisfyAll(
			BeAssignableToTypeOf(lib.FileAdded{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("e"),
			}),
		),
	))
}

func Test_detects_removed_files(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	removeFileBAndDirectoryC(p)
	changes, err := p.Hash()

	if err != nil {
		panic(err)
	}

	g.Expect(changes).To(HaveExactElements(
		SatisfyAll(
			BeAssignableToTypeOf(lib.FileRemoved{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("b"),
			}),
		),

		SatisfyAll(
			BeAssignableToTypeOf(lib.FileRemoved{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("c/d"),
			}),
		),
	))
}

func Test_detects_modified_files(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	modifyFileA(p)
	changes, err := p.Hash()

	if err != nil {
		panic(err)
	}

	g.Expect(changes).To(HaveExactElements(
		SatisfyAll(
			BeAssignableToTypeOf(lib.FileModified{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("a"),
			}),
		),
	))
}

func Test_if_called_right_after_changes_are_applied_and_saved_returns_no_changes(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	addFileE(p)
	removeFileBAndDirectoryC(p)
	modifyFileA(p)

	changes, err := p.Hash()

	if err != nil {
		panic(err)
	}

	p.ApplyChanges(changes)

	if err := p.Save(); err != nil {
		panic(err)
	}

	changes2, err := p.Hash()

	if err != nil {
		panic(err)
	}

	g.Expect(changes2).To(BeEmpty())
}
