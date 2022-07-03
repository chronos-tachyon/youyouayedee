package uuid

import (
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
// The nil UUID and the max UUID are not valid UUIDs according to this test.
//
func (uuid UUID) IsValid() bool {
	return (uuid[8] & 0xc0) == 0x80
}

// Version returns the UUID's version field.  May not be meaningful if IsValid
// returns false.
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
	// "00112233-4455-6677-8899-aabbccddeeff"
	//           ^^   ^^   ^^   ^^
	const hyphenMap = (1 << 0xa) | (1 << 0x8) | (1 << 0x6) | (1 << 0x4)

	for bi := uint(0); bi < Size; bi++ {
		if bit := uint(1) << bi; (hyphenMap & bit) != 0 {
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

func parseImpl(str string, a, b, c, d, e uint) (UUID, error) {
	strIndex := [Size]uint{
		a + 0x0, a + 0x2, a + 0x4, a + 0x6,
		b + 0x0, b + 0x2, c + 0x0, c + 0x2,
		d + 0x0, d + 0x2, e + 0x0, e + 0x2,
		e + 0x4, e + 0x6, e + 0x8, e + 0xa,
	}

	var uuid UUID
	var ok [Size]bool
	for bi := uint(0); bi < Size; bi++ {
		si := strIndex[bi]
		ok[bi], uuid[bi] = decodeHexByte(str, si)
	}

	allZeroes := true
	allOnes := true
	for bi := uint(0); bi < Size; bi++ {
		if !ok[bi] {
			si := strIndex[bi]
			if isHex(str[si]) {
				si++
			}
			return NilUUID, ParseError{
				Input:   str,
				Problem: UnexpectedCharacter,
				Args:    []interface{}{si},
				Index:   si,
			}
		}
		allZeroes = allZeroes && (uuid[bi] == 0x00)
		allOnes = allOnes && (uuid[bi] == 0xff)
	}

	if !allZeroes && !allOnes && (uuid[8]&0xc0) != 0x80 {
		vb := uuid[8]
		return NilUUID, ParseError{
			Input:       str,
			Problem:     WrongVariant,
			Args:        []interface{}{vb},
			VariantByte: vb,
		}
	}

	return uuid, nil
}

const upperCase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

const hexEncode = "0123456789abcdef"

var hexDecode = [256]byte{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x00 .. 0x07
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x08 .. 0x0f
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x10 .. 0x17
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x18 .. 0x1f
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x20 .. 0x27
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x28 .. 0x2f
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, // 0x30 .. 0x37
	0x08, 0x09, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x38 .. 0x3f
	0xff, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0xff, // 0x40 .. 0x47
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x48 .. 0x4f
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x50 .. 0x57
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x58 .. 0x5f
	0xff, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0xff, // 0x60 .. 0x67
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x68 .. 0x6f
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x70 .. 0x77
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x78 .. 0x7f
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x80 .. 0x87
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x88 .. 0x8f
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x90 .. 0x97
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0x98 .. 0x9f
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xa0 .. 0xa7
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xa8 .. 0xaf
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xb0 .. 0xb7
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xb8 .. 0xbf
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xc0 .. 0xc7
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xc8 .. 0xcf
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xd0 .. 0xd7
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xd8 .. 0xdf
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xe0 .. 0xe7
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xe8 .. 0xef
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xf0 .. 0xf7
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // 0xf8 .. 0xff
}

func appendHexByte(out []byte, value byte) []byte {
	hi := (value >> 4)
	lo := (value & 0xf)
	return append(out, hexEncode[hi], hexEncode[lo])
}

func decodeHexByte(str string, si uint) (bool, byte) {
	hi := hexDecode[str[si]]
	lo := hexDecode[str[si+1]]
	if hi == 0xff || lo == 0xff {
		return false, 0
	}
	value := (hi << 4) | lo
	return true, value
}

func isHex(ch byte) bool {
	return hexDecode[ch] != 0xff
}
