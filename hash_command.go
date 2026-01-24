package main

import (
	"context"
	"fmt"

	"github.com/azerum/partition/lib"
)

func hashCommand(partitionDir string) error {
	partition, err := lib.LoadPartition(partitionDir)

	if err != nil {
		return nil
	}

	changes, err := partition.Hash(context.Background()).Drain()

	if err != nil {
		return err
	}

	for _, change := range changes {
		line := sprintManifestChange(change)
		fmt.Println(line)
	}

	partition.ApplyChanges(changes)
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
