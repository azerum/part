package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/azerum/partition/lib"
)

func main() {
	if len(os.Args) < 2 {
		printUsageAndExit("Missing subcommand")
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "hash":
		if len(os.Args) != 3 {
			printUsageAndExit("hash requires exactly 1 arg")
		}

		partitionDir := os.Args[2]
		err := hashCommand(partitionDir)

		if err != nil {
			panic(err)
		}

	case "check":
		partitionDirs := os.Args[2:]
		exitCode, err := checkCommand(partitionDirs)

		if err != nil {
			panic(err)
		}

		if exitCode != 0 {
			os.Exit(exitCode)
		}

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

func checkCommand(partitionDirs []string) (int, error) {
	input := fanOutPartitionDirs(partitionDirs)
	ctx, cancel := context.WithCancelCause(context.Background())

	lines := mapConcurrently(input, doCheck, runtime.NumCPU(), ctx, cancel)
	hadAtLeastOneMismatch := false

	for {
		select {
		case l, ok := <-lines:
			if !ok {
				if hadAtLeastOneMismatch {
					return 1, nil
				}

				return 0, nil
			}

			hadAtLeastOneMismatch = true
			fmt.Println(l)

		case <-ctx.Done():
			return 1, ctx.Err()
		}
	}
}

func fanOutPartitionDirs(partitionDirs []string) <-chan string {
	out := make(chan string)

	go func() {
		for _, dir := range partitionDirs {
			out <- dir
		}

		close(out)
	}()

	return out
}

// Returns list of lines of mismatches to output to stdout
func doCheck(partitionDir string, ctx context.Context) ([]string, error) {
	partition, err := lib.LoadPartition(partitionDir)

	if err != nil {
		return nil, err
	}

	mismatches, err := partition.Check(ctx)

	if err != nil {
		return nil, err
	}

	lines := make([]string, 0)

	for _, m := range mismatches {
		l := sprintManifestMismatch(partitionDir, m)
		lines = append(lines, l)
	}

	return lines, nil
}

func sprintManifestMismatch(partitionDir string, mismatch lib.ManifestMismatch) string {
	switch c := mismatch.(type) {
	case lib.FileNotHashed:
		return fmt.Sprintf("?+ %s %s", partitionDir, c.ManifestPath)

	case lib.FileMissing:
		return fmt.Sprintf("?- %s %s", partitionDir, c.ManifestPath)

	case lib.HashDoesNotMatch:
		return fmt.Sprintf("?* %s %s actual=%s expected=%s", partitionDir, c.ManifestPath, c.ActualHash, c.ExpectedHash)

	default:
		panic(fmt.Sprintf("Unknown ManifestMismatch: %+v", mismatch))
	}
}
