package uuid

import (
	"io"
)

// Node represents an EUI-48 network hardware address.
type Node [6]byte

// NilNode represents the invalid nil hardware address.
var NilNode = Node{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

// IsZero returns true iff this Node is the nil hardware address.
func (node Node) IsZero() bool {
	return node == NilNode
}

// IsGlobal returns true if the G/L bit is set to "G", meaning that it is
// globally unique a.k.a. "OUI ENFORCED".
//
// As a special case, it returns false for the nil hardware address.
//
func (node Node) IsGlobal() bool {
	return (node != NilNode) && ((node[0] & 0x02) == 0)
}

// IsLocal returns true if the G/L bit is set to "L", meaning that it is
// locally defined a.k.a. "LOCALLY ADMINISTERED".
//
// As a special case, it also returns true for the nil hardware address.
//
func (node Node) IsLocal() bool {
	return !node.IsGlobal()
}

// IsUnicast returns true if the U/M bit is set to "U", meaning that it is a
// unicast hardware address.
//
// As a special case, it returns false for the nil hardware address.
//
func (node Node) IsUnicast() bool {
	return (node != NilNode) && ((node[0] & 0x01) == 0)
}

// IsMulticast returns true if the U/M bit is set to "M", meaning that it is a
// multicast hardware address.  Multicast addresses are usually generated
// on-the-fly and are not unique to one host.
//
// As a special case, it also returns true for the nil hardware address.
//
func (node Node) IsMulticast() bool {
	return !node.IsUnicast()
}

// NodeOptions supplies options for looking up and/or generating a Node value.
type NodeOptions struct {
	// ForceRandomNode specifies whether or not to force the use of a
	// randomly chosen Node value.  If true, the host's network hardware
	// addresses are ignored completely.  If false, a random Node value is
	// only used if no appropriate hardware address can be found.
	ForceRandomNode bool

	// RandomSource specifies a source of random bytes.  If this field is
	// nil but a source of random bytes is required, then
	// "crypto/rand".Reader will be used instead.
	RandomSource io.Reader
}

// GenerateNode looks up the best Node value for the current host's network
// hardware address, or else it generates one at random.
func GenerateNode(o NodeOptions) (Node, error) {
	var node Node
	if !o.ForceRandomNode {
		nodes, err := listHardwareAddresses()
		if err != nil {
			return NilNode, IOError{Err: err}
		}
		if len(nodes) > 0 {
			return nodes[0], nil
		}
	}

	if err := readRandom(o.RandomSource, node[:]); err != nil {
		return NilNode, IOError{Err: err}
	}

	// Set the G/L bit to L and the U/M bit to M.
	node[0] = (node[0] | 0x03)
	return node, nil
}
