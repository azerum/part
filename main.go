package main

import (
	"fmt"
	"os"

	"github.com/azerum/partition/lib"
)

func main() {
	if len(os.Args) < 2 {
		printUsageAndExit("Missing subcommand")
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "check":
		partitionDirs := os.Args[2:]
		err := checkCommand(partitionDirs)

		if err != nil {
			panic(err)
		}

		return

	case "hash":
		if len(os.Args) != 3 {
			printUsageAndExit("hash requires exactly 1 arg")
		}

		partitionDir := os.Args[2]
		err := hashCommand(partitionDir)

		if err != nil {
			panic(err)
		}

		return

	default:
		printUsageAndExit(fmt.Sprintf("Unknown subcommand %s", subcommand))
	}
}

func printUsageAndExit(specificMessage string) {
	fmt.Printf("error: %s\n\n", specificMessage)

	fmt.Print(
		"Usage:\n\n" +
			"- check <partition_dirs>... - check hashes of given partition directories\n" +
			"- hash <partition_dir> - (re)hash given partition directory. Incremental\n\n",
	)

	os.Exit(1)
}

func hashCommand(partitionDir string) error {
	partition, err := lib.LoadPartition(partitionDir)

	if err != nil {
		return nil
	}

	changes, err := partition.Hash()

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

func checkCommand(partitionDirs []string) error {
	return nil
}
