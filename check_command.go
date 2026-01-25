package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/azerum/part/partition_lib"
	"github.com/azerum/part/utils"
)

func checkCommand(partitionDirs []string) (int, error) {
	input := fanOutPartitionDirs(partitionDirs)

	lines := utils.MapConcurrently(
		input,
		checkPartitionAndGetStdoutLines,
		runtime.NumCPU(),
	)

	hadAtLeastOneMismatch := false

	for l := range lines.Channel {
		fmt.Println(l)
		hadAtLeastOneMismatch = true
	}

	if lines.Err != nil {
		return 1, lines.Err
	}

	if hadAtLeastOneMismatch {
		return 1, nil
	}

	return 0, nil
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

func checkPartitionAndGetStdoutLines(
	partitionDir string,
) *utils.ChanWithError[string] {
	lines := utils.NewChanWithError[string](1)

	go func() {
		partition, err := partition_lib.LoadPartition(partitionDir)

		if err != nil {
			lines.CloseWithError(err)
			return
		}

		mismatches := partition.Check(context.Background())

		for m := range mismatches.Channel {
			l := sprintManifestMismatch(partitionDir, m)
			lines.Channel <- l
		}

		if mismatches.Err != nil {
			lines.CloseWithError(mismatches.Err)
		} else {
			lines.CloseOk()
		}
	}()

	return lines
}

func sprintManifestMismatch(partitionDir string, mismatch partition_lib.ManifestMismatch) string {
	switch c := mismatch.(type) {
	case partition_lib.FileNotHashed:
		return fmt.Sprintf("?+ %s %s", partitionDir, c.ManifestPath)

	case partition_lib.FileMissing:
		return fmt.Sprintf("?- %s %s", partitionDir, c.ManifestPath)

	case partition_lib.HashDoesNotMatch:
		return fmt.Sprintf("?* %s %s actual=%s expected=%s", partitionDir, c.ManifestPath, c.ActualHash, c.ExpectedHash)

	default:
		panic(fmt.Sprintf("Unknown ManifestMismatch: %+v", mismatch))
	}
}
