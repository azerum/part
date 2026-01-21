package main_test

import (
	"testing"

	"github.com/azerum/partition/lib"
	. "github.com/onsi/gomega"
	gs "github.com/onsi/gomega/gstruct"
)

func Test_Check_detects_unhashed_file(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	addFileE(p)
	mismatches := p.Check()

	g.Expect(mismatches).To(HaveExactElements(
		SatisfyAll(
			BeAssignableToTypeOf(lib.FileNotHashed{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("e"),
			}),
		),
	))
}

func Test_Check_detects_deleted_file(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	removeFileBAndDirectoryC(p)
	mismatches := p.Check()

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

func Test_Check_detects_changed_file(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	modifyFileA(p)
	mismatches := p.Check()

	g.Expect(mismatches).To(HaveExactElements(
		SatisfyAll(
			BeAssignableToTypeOf(lib.HashDoesNotMatch{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("a"),
			}),
		),
	))
}

func Test_Check_ignores_mtime_only_changes(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	modifyFileAMtime(p)
	mismatches := p.Check()

	g.Expect(mismatches).To(BeEmpty())
}
