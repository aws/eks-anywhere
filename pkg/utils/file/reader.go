package file

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func ReadFile(fileName string) (io.Reader, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}

	return bytes.NewReader(content), nil
}
