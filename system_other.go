//go:build !js && !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris
// +build !js,!aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package youyouayedee

import (
	"fmt"
	"net"
	"os"
)

func lockFile(file *os.File) error {
	return &os.SyscallError{
		Syscall: "Flock",
		Err:     fmt.Errorf("not available for this OS"),
	}
}

func listHardwareAddresses() ([]Node, error) {
	list, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	candidates := make(hwaddrCandidates, 0, len(list))
	for _, iface := range list {
		if (iface.Flags & net.FlagLoopback) != 0 {
			continue
		}

		if node, ok := convertHardwareAddrToNode([]byte(iface.HardwareAddr)); ok {
			candidates = append(candidates, hwaddrCandidate{
				Node:      node,
				Index:     iface.Index,
				IsGlobal:  node.IsGlobal(),
				IsUnicast: node.IsUnicast(),
			})
		}
	}

	candidates.Sort()

	candidatesLen := uint(len(candidates))
	nodes := make([]Node, candidatesLen)
	for index := uint(0); index < candidatesLen; index++ {
		nodes[index] = candidates[index].Node
	}
	return nodes, nil
}
