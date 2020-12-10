package scp

import (
	"fmt"
	"os"
)

func ConvertFileModeToPermString(mode os.FileMode) string {
	perm := mode.Perm()

	return fmt.Sprintf("0%.3o", perm)
}
