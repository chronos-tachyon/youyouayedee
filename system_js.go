//go:build js
// +build js

package youyouayedee

import (
	"fmt"
	"os"
)

const lockFileSupported = false

func lockFile(file *os.File) error {
	return &os.SyscallError{
		Syscall: "Flock",
		Err:     fmt.Errorf("not available for WASM"),
	}
}

func listHardwareAddresses() ([]Node, error) {
	return nil, nil
}
