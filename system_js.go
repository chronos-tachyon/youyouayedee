//go:build js
// +build js

package youyouayedee

import (
	"os"
)

const lockFileSupported = false

func lockFile(file *os.File) error {
	return &os.SyscallError{
		Syscall: "Flock",
		Err:     ErrLockNotSupported{},
	}
}

func listHardwareAddresses() ([]Node, error) {
	return nil, nil
}
