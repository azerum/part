package main

import (
	"context"
	"fmt"

	"github.com/azerum/part/lib"
)

func hashCommand(partitionDir string) error {
	partition, err := lib.LoadPartition(partitionDir)

	if err != nil {
		return nil
	}

	changes := partition.Hash(context.Background())

	for c := range changes.Channel {
		line := sprintManifestChange(c)
		fmt.Println(line)

		partition.ApplyChange(c)
	}

	if changes.Err != nil {
		return changes.Err
	}

	return partition.Save()
}

func sprintManifestChange(change lib.ManifestChange) string {
	switch c := change.(type) {
	case lib.FileAdded:
		return fmt.Sprintf("+ %s", c.ManifestPath)

	case lib.FileModified:
		return fmt.Sprintf("* %s", c.ManifestPath)

	case lib.FileDeleted:
		return fmt.Sprintf("- %s", c.ManifestPath)

	default:
		panic(fmt.Sprintf("Unknown ManifestChange: %+v", change))
	}
}
