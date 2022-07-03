package uuid

import (
	"fmt"
	"strings"
)

// Size is the size of a UUID in bytes.
const Size = 16

// UUID represents a UUID.  Only the RFC 4122 variant is supported.
type UUID [Size]byte

// NilUUID is the nil UUID, "00000000-0000-0000-0000-000000000000".
var NilUUID = UUID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

// MaxUUID is the maximum UUID, "ffffffff-ffff-ffff-ffff-ffffffffffff".
var MaxUUID = UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

// IsZero returns true iff this UUID is the nil UUID.
func (uuid UUID) IsZero() bool {
	return uuid == NilUUID
}

// IsMax returns true iff this UUID is the maximum UUID.
func (uuid UUID) IsMax() bool {
	return uuid == MaxUUID
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

// GoString formats the UUID as a developer-friendly string.
func (uuid UUID) GoString() string {
	//  1 * "uuid.UUID{}" = 11 bytes
	//  1 * "0x??"        =  4 bytes
	// 15 * ", 0x??"      = 90 bytes (15 * 6)
	//                      --------
	//                     105 bytes
	//
	// We round up to 128 because 2.

	var tmp [128]byte
	buf := tmp[:0]
	buf = append(buf, "uuid.UUID{"...)
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
	if strings.ContainsAny(str, upperCase) {
		str = strings.ToLower(str)
	}

	var requiredByteIndices []uint
	var requiredByteValues []byte
	var requiredByteCount uint
	var a, b, c, d, e uint
	var okToParse bool

	strLen := uint(len(str))
	switch strLen {
	case 0:
		return NilUUID, nil

	case 3:
		if str == "nil" {
			return NilUUID, nil
		}
		if str == "max" {
			return MaxUUID, nil
		}

	case 4:
		if str == "null" {
			return NilUUID, nil
		}

	case 32:
		a, b, c, d, e = 0, 8, 12, 16, 20
		okToParse = true

	case 36:
		requiredByteIndices = []uint{8, 13, 18, 23}
		requiredByteValues = []byte{'-', '-', '-', '-'}
		requiredByteCount = 4
		a, b, c, d, e = 0, 9, 14, 19, 24
		okToParse = true

	case 38:
		requiredByteIndices = []uint{0, 9, 14, 19, 24, 37}
		requiredByteValues = []byte{'{', '-', '-', '-', '-', '}'}
		requiredByteCount = 6
		a, b, c, d, e = 1, 10, 15, 20, 25
		okToParse = true

	case 41:
		requiredByteIndices = []uint{0, 1, 2, 3, 4, 5, 6, 7, 8, 17, 22, 27, 32}
		requiredByteValues = []byte{'u', 'r', 'n', ':', 'u', 'u', 'i', 'd', ':', '-', '-', '-', '-'}
		requiredByteCount = 13
		a, b, c, d, e = 9, 18, 23, 28, 33
		okToParse = true
	}

	if okToParse {
		for xi := uint(0); xi < requiredByteCount; xi++ {
			si := requiredByteIndices[xi]
			ch := requiredByteValues[xi]
			if str[si] != ch {
				return NilUUID, ParseError{
					Input:   str,
					Problem: UnexpectedCharacter,
					Args:    []interface{}{si},
					Index:   si,
				}
			}
		}
		return parseImpl(str, a, b, c, d, e)
	}

	return NilUUID, ParseError{
		Input:   str,
		Problem: WrongLength,
		Args:    []interface{}{strLen},
		Length:  strLen,
	}
}

// Parse parses a UUID from a []byte.
func ParseBytes(buf []byte) (UUID, error) {
	return Parse(string(buf))
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
