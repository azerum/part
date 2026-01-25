package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/azerum/data-storage-suite/pkg/utils"
)

func main() {
	concurrency := runtime.NumCPU()
	fileNames := fanOutArgs(concurrency)

	renames := utils.MapConcurrently(
		fileNames,
		calculateRenames,
		concurrency,
	)

	for r := range renames.Channel {
		err := os.Rename(r.oldPath, r.newPath)

		if err != nil {
			panic(err)
		}
	}

	if renames.Err != nil {
		panic(renames.Err)
	}
}

func fanOutArgs(bufferSize int) <-chan string {
	out := make(chan string, bufferSize)

	go func() {
		for _, filePath := range os.Args[1:] {
			out <- filePath
		}

		close(out)
	}()

	return out
}

type fileRename struct {
	oldPath string
	newPath string
}

func calculateRenames(filePath string) *utils.ChanWithError[fileRename] {
	out := utils.NewChanWithError[fileRename](0)

	go func() {
		info, err := os.Stat(filePath)

		if err != nil {
			out.CloseWithError(err)
			return
		}

		modTime := info.ModTime().UTC().Format("2006-01-02T15-04-05Z")
		_, extension := splitFileNameAndExtension(filePath)

		newPath := filepath.Join(
			filepath.Dir(filePath),
			fmt.Sprintf("%s.%s", modTime, extension),
		)

		out.Channel <- fileRename{
			oldPath: filePath,
			newPath: newPath,
		}

		out.CloseOk()
	}()

	return out
}

func splitFileNameAndExtension(filePath string) (string, string) {
	parts := strings.SplitAfterN(filePath, ".", 2)

	if len(parts) < 1 {
		panic(fmt.Sprintf("SplitAfterN() returned 0 strings. File path: %s", filePath))
	}

	if len(parts) == 1 {
		return parts[0], ""
	}

	return parts[0], parts[1]
}
