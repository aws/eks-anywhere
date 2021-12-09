package filewriter

import (
	"os"
)

const DefaultTmpFolder = "generated"

func defaultFileOptions() *FileOptions {
	return &FileOptions{true, os.ModePerm}
}

func Permission0600(op *FileOptions) {
	op.Permissions = 0o600
}

func PersistentFile(op *FileOptions) {
	op.IsTemp = false
}
