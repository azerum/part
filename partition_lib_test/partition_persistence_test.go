package partition_lib_test

import (
	"fmt"
	"testing"

	"github.com/azerum/data-storage-suite/partition_lib"
	. "github.com/onsi/gomega"
)

func Test_LoadPartition_returns_partition_without_manifest_if_directory_does_not_contain_manifest_file(t *testing.T) {
	g := NewGomegaWithT(t)

	dirPath := t.TempDir()
	p, err := partition_lib.LoadPartition(dirPath)

	g.Expect(err).To(BeNil())
	g.Expect(p).ToNot(BeNil())
}

func Test_LoadPartition_returns_error_if_manifest_file_is_malformed(t *testing.T) {
	g := NewGomegaWithT(t)

	// Here, .dataJson missing is missing
	manifestJson := `{"dataHash": "abc"}`

	_, err := partition_lib.DeserializePartition("dir", []byte(manifestJson))
	g.Expect(err).To(MatchError(ContainSubstring("while loading serialized manifest")))
}

func Test_LoadPartition_returns_error_if_manifest_hash_does_not_match_the_contents(t *testing.T) {
	g := NewGomegaWithT(t)

	dataJson := "{}"

	manifestJson := `{
		"dataHash": "wrong",
		"dataJson": "{}"
	}`

	actualHash := partition_lib.HashString(dataJson)

	_, err := partition_lib.DeserializePartition("dir", []byte(manifestJson))

	g.Expect(err).To(MatchError(SatisfyAll(
		ContainSubstring("manifest hash mismatch"),

		// Should mention expected hash
		ContainSubstring("wrong"),

		// Should mention actual hash
		ContainSubstring(actualHash),
	)))
}

func Test_LoadPartition_returns_error_if_manifest_is_malformed(t *testing.T) {
	g := NewGomegaWithT(t)

	// Here, the file entry is missing .hash

	dataJson := `{"files":{"path/to/file.txt":{"mtime":42}}}`
	dataHash := partition_lib.HashString(dataJson)

	manifestJson := fmt.Sprintf(`{
		"dataHash": "%s",
		"dataJson": "{\"files\":{\"path/to/file.txt\":{\"mtime\":42}}}"
	}`, dataHash)

	_, err := partition_lib.DeserializePartition("dir", []byte(manifestJson))

	g.Expect(err).To(MatchError(SatisfyAll(
		// Should mention that the error is in the .dataJson, not the outer
		// JSON of the manifest
		ContainSubstring(".dataJson"),

		// Should mention which entry is malformed
		ContainSubstring("path/to/file.txt"),

		// Should mention what is wrong about the entry - .hash is missing
		ContainSubstring(".hash"),
	)))
}
