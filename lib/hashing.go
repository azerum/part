package lib

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

func HashString(s string) string {
	bytes := sha1.Sum(([]byte)(s))
	return fmt.Sprintf("%x", bytes)
}

func HashFile(absoluteOsPath string) (string, error) {
	file, err := os.Open(absoluteOsPath)

	if err != nil {
		return "", err
	}

	hasher := sha1.New()

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	bytes := hasher.Sum(nil)
	return fmt.Sprintf("%x", bytes), nil
}
