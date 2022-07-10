package youyouayedee

import (
	"fmt"
	"sort"
)

// Node represents an EUI-48 network card hardware address, or something that
// has been formatted to look like one.
type Node [6]byte

// NilNode represents the invalid nil node identifier, "00:00:00:00:00:00".
var NilNode = Node{}

// IsZero returns true iff this Node is the nil node identifier.
func (node Node) IsZero() bool {
	return node == NilNode
}

// IsGlobal returns true if the G/L bit is set to "G", meaning that it is
// globally unique a.k.a. "OUI ENFORCED".
//
// As a special case, it returns false for the nil node identifier.
//
func (node Node) IsGlobal() bool {
	return (node != NilNode) && ((node[0] & 0x02) == 0)
}

// IsLocal returns true if the G/L bit is set to "L", meaning that it is
// locally defined a.k.a. "LOCALLY ADMINISTERED".
//
// As a special case, it also returns true for the nil node identifier.
//
func (node Node) IsLocal() bool {
	return !node.IsGlobal()
}

// IsUnicast returns true if the U/M bit is set to "U", meaning that it is a
// unicast EUI-48 address.
//
// As a special case, it returns false for the nil node identifier.
//
func (node Node) IsUnicast() bool {
	return (node != NilNode) && ((node[0] & 0x01) == 0)
}

// IsMulticast returns true if the U/M bit is set to "M", meaning that it is a
// multicast EUI-48 address.
//
// As a special case, it also returns true for the nil node identifier.
//
// Multicast addresses are usually generated on-the-fly and are not unique to
// one host, so they are rarely the best choice for UUID uniqueness.
//
func (node Node) IsMulticast() bool {
	return !node.IsUnicast()
}

// GoString formats the Node as a developer-friendly string.
func (node Node) GoString() string {
	if node.IsZero() {
		return "youyouayedee.NilNode"
	}

	var tmp [64]byte
	buf := tmp[:0]
	buf = append(buf, "youyouayedee.Node{"...)
	for bi := uint(0); bi < 6; bi++ {
		if bi == 0 {
			buf = append(buf, '0', 'x')
		} else {
			buf = append(buf, ',', ' ', '0', 'x')
		}
		buf = appendHexByte(buf, node[bi])
	}
	buf = append(buf, '}')
	return string(buf)
}

// String formats the Node in the standard colon-delimited way for an EUI-48.
func (node Node) String() string {
	var tmp [64]byte
	return string(node.AppendTo(tmp[:0]))
}

// AppendTo appends the Node's colon-delimited EUI-48 to the given []byte.
func (node Node) AppendTo(out []byte) []byte {
	for bi := uint(0); bi < 6; bi++ {
		if bi != 0 {
			out = append(out, ':')
		}
		out = appendHexByte(out, node[bi])
	}
	return out
}

// GenerateNode returns the best available node identifier given the current
// host's EUI-48 and EUI-64 network addresses, or else it generates one at
// random as a fallback.
func GenerateNode(o Options) (Node, error) {
	var node Node
	if !o.ForceRandomNode {
		nodes, err := listHardwareAddresses()
		if err != nil {
			return NilNode, err
		}
		if len(nodes) > 0 {
			return nodes[0], nil
		}
	}

	if err := readRandom(o.RandomSource, node[:]); err != nil {
		return NilNode, err
	}

	// Set the G/L bit to L and the U/M bit to M.
	node[0] = (node[0] | 0x03)
	return node, nil
}

var (
	_ fmt.GoStringer = Node{}
	_ fmt.Stringer   = Node{}
)

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

func convertHardwareAddrToNode(hwaddr []byte) (Node, bool) {
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
