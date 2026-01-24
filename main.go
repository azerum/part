package main

import (
	"fmt"
	"os"
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
