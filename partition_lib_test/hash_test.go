package partition_lib_test

import (
	"context"
	"testing"

	"github.com/azerum/part/partition_lib"
	. "github.com/onsi/gomega"
	gs "github.com/onsi/gomega/gstruct"
)

func Test_Hash_creates_entire_manifest_when_run_for_the_first_time(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	changes, err := p.Hash(context.Background()).Drain()

	if err != nil {
		panic(err)
	}

	g.Expect(changes).To(ConsistOf(
		SatisfyAll(
			BeAssignableToTypeOf(partition_lib.FileAdded{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("a"),
			}),
		),

		SatisfyAll(
			BeAssignableToTypeOf(partition_lib.FileAdded{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("b"),
			}),
		),

		SatisfyAll(
			BeAssignableToTypeOf(partition_lib.FileAdded{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("c/d"),
			}),
		),

		SatisfyAll(
			BeAssignableToTypeOf(partition_lib.FileAdded{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("e"),
			}),
		),
	))
}

func Test_Hash_detects_added_files(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	addFileF(p)
	changes, err := p.Hash(context.Background()).Drain()

	if err != nil {
		panic(err)
	}

	g.Expect(changes).To(ConsistOf(
		SatisfyAll(
			BeAssignableToTypeOf(partition_lib.FileAdded{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("f"),
			}),
		),
	))
}

func Test_Hash_detects_deleted_files(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	removeFileBAndDirectoryC(p)
	changes, err := p.Hash(context.Background()).Drain()

	if err != nil {
		panic(err)
	}

	g.Expect(changes).To(ConsistOf(
		SatisfyAll(
			BeAssignableToTypeOf(partition_lib.FileDeleted{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("b"),
			}),
		),

		SatisfyAll(
			BeAssignableToTypeOf(partition_lib.FileDeleted{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("c/d"),
			}),
		),
	))
}

func Test_Hash_detects_modified_files(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	modifyFileA(p)
	changes, err := p.Hash(context.Background()).Drain()

	if err != nil {
		panic(err)
	}

	g.Expect(changes).To(ConsistOf(
		SatisfyAll(
			BeAssignableToTypeOf(partition_lib.FileModified{}),

			gs.MatchFields(gs.IgnoreExtras, gs.Fields{
				"ManifestPath": Equal("a"),
			}),
		),
	))
}

func Test_Hash_if_called_right_after_changes_are_applied_and_saved_returns_no_changes(t *testing.T) {
	g := NewGomegaWithT(t)

	p := setupTestPartition(t)
	hashAndSave(p)

	addFileF(p)
	removeFileBAndDirectoryC(p)
	modifyFileA(p)
	modifyFileEMtime(p)

	changes := p.Hash(context.Background())

	for c := range changes.Channel {
		p.ApplyChange(c)
	}

	if changes.Err != nil {
		panic(changes.Err)
	}

	if err := p.Save(); err != nil {
		panic(err)
	}

	changes2, err := p.Hash(context.Background()).Drain()

	if err != nil {
		panic(err)
	}

	g.Expect(changes2).To(BeEmpty())
}
