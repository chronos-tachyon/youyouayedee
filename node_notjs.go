//go:build !js
// +build !js

package uuid

import (
	"net"
	"sort"
)

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

		if node, ok := convertHardwareAddrToNode(iface.HardwareAddr); ok {
			candidates = append(candidates, hwaddrCandidate{
				Node:      node,
				Index:     iface.Index,
				IsGlobal:  node.IsGlobal(),
				IsUnicast: node.IsUnicast(),
			})
		}
	}

	sort.Sort(candidates)

	candidatesLen := uint(len(candidates))
	nodes := make([]Node, candidatesLen)
	for index := uint(0); index < candidatesLen; index++ {
		nodes[index] = candidates[index].Node
	}
	return nodes, nil
}

func convertHardwareAddrToNode(hwaddr net.HardwareAddr) (Node, bool) {
	hwaddrLen := uint(len(hwaddr))

	// EUI-48
	if hwaddrLen == 6 {
		var node Node
		copy(node[:], hwaddr)
		if !node.IsZero() {
			return node, true
		}
	}

	// EUI-64
	if hwaddrLen == 8 && hwaddr[3] == 0xff && hwaddr[4] == 0xfe {
		var node Node
		node[0] = hwaddr[0] ^ 0x02
		node[1] = hwaddr[1]
		node[2] = hwaddr[2]
		node[3] = hwaddr[5]
		node[4] = hwaddr[6]
		node[5] = hwaddr[7]
		if !node.IsZero() {
			return node, true
		}
	}

	return NilNode, false
}

type hwaddrCandidate struct {
	Node      Node
	Index     int
	IsGlobal  bool
	IsUnicast bool
}

type hwaddrCandidates []hwaddrCandidate

func (list hwaddrCandidates) Len() int {
	return len(list)
}

func (list hwaddrCandidates) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list hwaddrCandidates) Less(i, j int) bool {
	a := list[i]
	b := list[j]
	if a.IsGlobal != b.IsGlobal {
		return a.IsGlobal
	}
	if a.IsUnicast != b.IsUnicast {
		return a.IsUnicast
	}
	return a.Index < b.Index
}

func (list hwaddrCandidates) Sort() {
	sort.Sort(list)
}

var _ sort.Interface = hwaddrCandidates(nil)
