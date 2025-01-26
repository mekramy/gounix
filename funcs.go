package gounix

import (
	"os"
)

// IsSudo checks if the program is running with sudo privileges.
func IsSudo() bool {
	return os.Geteuid() == 0
}
