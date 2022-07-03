package youyouayedee

import (
	"fmt"
	"io"
)

// Node represents an EUI-48 network card hardware address, or something that
// has been formatted to look like one.
type Node [6]byte

// NilNode represents the invalid nil node identifier.
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

// NodeOptions supplies options for obtaining a reasonably unique node
// identifier.
type NodeOptions struct {
	// ForceRandomNode specifies whether or not to force the use of a
	// randomly chosen node identifier.
	//
	// If true, the host's hardware addresses are ignored completely and a
	// random node identifier is always chosen.
	//
	// If false, a random node identifier is only used if no appropriate
	// hardware address (EUI-48, or EUI-64 that's backward compatible with
	// EUI-48) can be found.
	//
	ForceRandomNode bool

	// RandomSource specifies a source of random bytes.
	//
	// If this field is nil but a source of random bytes is required, then
	// "crypto/rand".Reader will be used instead.
	//
	RandomSource io.Reader
}

// GenerateNode returns the best available node identifier given the current
// host's EUI-48 and EUI-64 network addresses, or else it generates one at
// random as a fallback.
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

var (
	_ fmt.GoStringer = Node{}
	_ fmt.Stringer   = Node{}
)
