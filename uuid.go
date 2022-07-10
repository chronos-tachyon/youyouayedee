package youyouayedee

import (
	"encoding/binary"
	"fmt"
)

// Size is the size of a UUID in bytes.
const Size = 16

// UUID represents a UUID.  Only the RFC 4122 variant is supported.
type UUID [Size]byte

// Nil is the nil UUID, "00000000-0000-0000-0000-000000000000".
var Nil = UUID{}

// Max is the maximum UUID, "ffffffff-ffff-ffff-ffff-ffffffffffff".
var Max = UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

// IsZero returns true iff this UUID is the nil UUID.
func (uuid UUID) IsZero() bool {
	return uuid == Nil
}

// IsMax returns true iff this UUID is the maximum UUID.
func (uuid UUID) IsMax() bool {
	return uuid == Max
}

// IsValid returns true iff this UUID is a valid UUID according to the rules
// for the RFC 4122 variant format.
//
// The nil UUID and the max UUID are *not* valid UUIDs according to this test.
//
func (uuid UUID) IsValid() bool {
	return (uuid[8] & 0xc0) == 0x80
}

// Version returns the UUID's version field.
//
// This value may not be meaningful if IsValid would return false.
//
func (uuid UUID) Version() Version {
	return Version(uuid[6] >> 4)
}

// Domain returns the DCE domain of the UUID.
//
// This value is only meaningful for valid version 2 UUIDs.
//
func (uuid UUID) Domain() DCEDomain {
	return DCEDomain(uuid[9])
}

// Domain returns the DCE domain of the UUID.
//
// This value is only meaningful for valid version 2 UUIDs.
//
func (uuid UUID) ID() uint32 {
	return binary.BigEndian.Uint32(uuid[0:4])
}

// GoString formats the UUID as a developer-friendly string.
func (uuid UUID) GoString() string {
	if uuid.IsZero() {
		return "youyouayedee.Nil"
	}

	if uuid.IsMax() {
		return "youyouayedee.Max"
	}

	//  1 * "youyouayedee.UUID{}" = 19 bytes
	//  1 * "0x??"                =  4 bytes
	// 15 * ", 0x??"              = 90 bytes (15 * 6)
	//                              --------
	//                             113 bytes
	//
	// We round up to 128 because 2.

	var tmp [128]byte
	buf := tmp[:0]
	buf = append(buf, "youyouayedee.UUID{"...)
	for bi := uint(0); bi < Size; bi++ {
		if bi == 0 {
			buf = append(buf, '0', 'x')
		} else {
			buf = append(buf, ',', ' ', '0', 'x')
		}
		buf = appendHexByte(buf, uuid[bi])
	}
	buf = append(buf, '}')
	return string(buf)
}

// String formats the UUID using the standard string representation.
func (uuid UUID) String() string {
	var tmp [64]byte
	return string(uuid.AppendTo(tmp[:0]))
}

// URN formats the UUID as a Uniform Resource Name in the "uuid" namespace.
func (uuid UUID) URN() string {
	var tmp [64]byte
	return string(uuid.AppendTo(append(tmp[:0], "urn:uuid:"...)))
}

// AppendTo appends the UUID's string representation to the given []byte.
func (uuid UUID) AppendTo(out []byte) []byte {
	for bi := uint(0); bi < Size; bi++ {
		if bi == 4 || bi == 6 || bi == 8 || bi == 10 {
			out = append(out, '-')
		}
		out = appendHexByte(out, uuid[bi])
	}
	return out
}

// Parse parses a UUID from a string.
func Parse(str string) (UUID, error) {
	var tmp [64]byte
	input := append(tmp[:0], str...)
	return parse(input, false)
}

// Parse parses a UUID from a []byte.
func ParseBytes(buf []byte) (UUID, error) {
	return parse(buf, true)
}

// Must is a panic wrapper for Parse, NewUUID, et al.
func Must(uuid UUID, err error) UUID {
	if err != nil {
		panic(err)
	}
	return uuid
}

var (
	_ fmt.GoStringer = UUID{}
	_ fmt.Stringer   = UUID{}
)
