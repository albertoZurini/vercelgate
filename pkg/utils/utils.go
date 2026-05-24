package utils

import (
	"fmt"
	"os"
)

func OpenFile(filepath string) ([]byte, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory", filepath)
	}

	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}
