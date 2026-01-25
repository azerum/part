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

	// Do not apply changes immediately, as Hash() reads from partition.manifest.Files,
	// and concurrent read+write is not safe
	//
	// TODO: use sync.Map? Concurrent parts of code never act on the same map
	// key
	changesList := make([]lib.ManifestChange, 0)

	for c := range changes.Channel {
		line := sprintManifestChange(c)
		fmt.Println(line)

		changesList = append(changesList, c)
	}

	if changes.Err != nil {
		return changes.Err
	}

	for _, c := range changesList {
		partition.ApplyChange(c)
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
