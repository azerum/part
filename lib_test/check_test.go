package lib_test

import (
	"context"
	"testing"

	"github.com/azerum/partition/lib"
	. "github.com/onsi/gomega"
	gs "github.com/onsi/gomega/gstruct"
)

func Test_Check_detects_added_but_not_hashed_files(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	addFileF(p)
	mismatches, err := p.Check(context.Background())

	if err != nil {
		panic(err)
	}

	g.Expect(mismatches).To(HaveExactElements(
		SatisfyAll(
			BeAssignableToTypeOf(lib.FileNotHashed{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("f"),
			}),
		),
	))
}

func Test_Check_detects_missing_files(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	removeFileBAndDirectoryC(p)
	mismatches, err := p.Check(context.Background())

	if err != nil {
		panic(err)
	}

	g.Expect(mismatches).To(HaveExactElements(
		SatisfyAll(
			BeAssignableToTypeOf(lib.FileMissing{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("b"),
			}),
		),

		SatisfyAll(
			BeAssignableToTypeOf(lib.FileMissing{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("c/d"),
			}),
		),
	))
}

func Test_Check_detects_when_file_contents_change_after_hashing(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	modifyFileA(p)
	mismatches, err := p.Check(context.Background())

	if err != nil {
		panic(err)
	}

	g.Expect(mismatches).To(HaveExactElements(
		SatisfyAll(
			BeAssignableToTypeOf(lib.HashDoesNotMatch{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("a"),
			}),
		),
	))
}

func Test_Check_does_not_consider_file_modified_if_it_changes_mtime_but_not_contents(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	modifyFileEMtime(p)
	mismatches, err := p.Check(context.Background())

	if err != nil {
		panic(err)
	}

	g.Expect(mismatches).To(BeEmpty())
}
