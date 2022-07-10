package youyouayedee

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"strconv"
)

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

func decodeHexByte(input []byte, ii uint) (bool, byte) {
	hi := hexDecode[input[ii]]
	lo := hexDecode[input[ii+1]]
	if hi < 0x10 && lo < 0x10 {
		value := (hi << 4) | lo
		return true, value
	}
	return false, 0
}

func isHex(input []byte, ii uint) bool {
	hex := hexDecode[input[ii]]
	return hex < 0x10
}

func readRandom(rng io.Reader, out []byte) error {
	if rng == nil {
		rng = rand.Reader
	}

	_, err := io.ReadFull(rng, out)
	if err != nil {
		return ErrOperationFailed{Operation: ReadRandomOp, Err: err}
	}
	return nil
}

func isErrClockNotFound(err error) bool {
	var unavailable ErrClockNotFound
	return errors.Is(err, &unavailable)
}

func parse(input []byte, isBytes bool) (UUID, error) {
	var output UUID
	var requiredByteIndices []uint
	var requiredByteValues []byte
	var requiredByteCount uint
	var a, b, c, d, e uint
	var okToParse, allZeroes, allOnes bool

	inputLen := uint(len(input))

	if bytes.ContainsAny(input, upperCase) {
		dupe := make([]byte, inputLen)
		copy(dupe, input)
		input = bytes.ToLower(dupe)
	}

	switch inputLen {
	case 0:
		return Nil, nil

	case 3:
		if string(input) == "nil" {
			return Nil, nil
		}
		if string(input) == "max" {
			return Max, nil
		}

	case 4:
		if string(input) == "null" {
			return Nil, nil
		}

	case 16:
		if isBytes {
			copy(output[:], input)
			allZeroes = true
			allOnes = true
			for oi := uint(0); oi < Size; oi++ {
				allZeroes = allZeroes && (output[oi] == 0x00)
				allOnes = allOnes && (output[oi] == 0xff)
			}
			return checkParse(input, output, allZeroes, allOnes)
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

	case 45:
		requiredByteIndices = []uint{0, 1, 2, 3, 4, 5, 6, 7, 8, 17, 22, 27, 32}
		requiredByteValues = []byte{'u', 'r', 'n', ':', 'u', 'u', 'i', 'd', ':', '-', '-', '-', '-'}
		requiredByteCount = 13
		a, b, c, d, e = 9, 18, 23, 28, 33
		okToParse = true
	}

	if !okToParse {
		problem := WrongTextLength
		if isBytes {
			problem = WrongBinaryLength
		}
		return Nil, ErrParseFailed{
			Input:   input,
			Problem: problem,
			Args:    mkargs(inputLen),
		}
	}

	for xi := uint(0); xi < requiredByteCount; xi++ {
		ii := requiredByteIndices[xi]
		ch := requiredByteValues[xi]
		if input[ii] != ch {
			return Nil, ErrParseFailed{
				Input:      input,
				Problem:    UnexpectedCharacter,
				Args:       mkargs(input[ii], ii, strconv.QuoteRune(rune(ch))),
				Index:      ii,
				ExpectByte: ch,
				ActualByte: input[ii],
			}
		}
	}

	inputIndex := [Size]uint{
		a + 0x0, a + 0x2, a + 0x4, a + 0x6,
		b + 0x0, b + 0x2, c + 0x0, c + 0x2,
		d + 0x0, d + 0x2, e + 0x0, e + 0x2,
		e + 0x4, e + 0x6, e + 0x8, e + 0xa,
	}

	var ok [Size]bool
	for oi := uint(0); oi < Size; oi++ {
		ok[oi], output[oi] = decodeHexByte(input, inputIndex[oi])
	}

	allZeroes = true
	allOnes = true
	for oi := uint(0); oi < Size; oi++ {
		if !ok[oi] {
			ii := inputIndex[oi]
			if isHex(input, ii) {
				ii++
			}
			return Nil, ErrParseFailed{
				Input:      input,
				Problem:    UnexpectedCharacter,
				Args:       mkargs(input[ii], ii, "hex digit [0-9a-f]"),
				Index:      ii,
				ActualByte: input[ii],
			}
		}
		allZeroes = allZeroes && (output[oi] == 0x00)
		allOnes = allOnes && (output[oi] == 0xff)
	}

	return checkParse(input, output, allZeroes, allOnes)
}

func checkParse(input []byte, output UUID, allZeroes bool, allOnes bool) (UUID, error) {
	actualVB := output[8]
	expectVB := (actualVB & 0x3f) | 0x80
	if actualVB == expectVB || allZeroes || allOnes {
		return output, nil
	}

	return Nil, ErrParseFailed{
		Input:      input,
		Problem:    WrongVariant,
		Args:       mkargs(actualVB, expectVB),
		ExpectByte: expectVB,
		ActualByte: actualVB,
	}
}

func mkargs(args ...interface{}) []interface{} {
	return args
}

func getUint48(in []byte) uint64 {
	var tmp [8]byte
	copy(tmp[2:8], in[0:6])
	return binary.BigEndian.Uint64(tmp[:])
}

func getV1Ticks(in []byte) uint64 {
	lo := binary.BigEndian.Uint32(in[0:4])
	mid := binary.BigEndian.Uint16(in[4:6])
	hi := binary.BigEndian.Uint16(in[6:8]) & 0x0fff
	return (uint64(hi) << 48) | (uint64(mid) << 32) | uint64(lo)
}

func getV6Ticks(in []byte) uint64 {
	hi := binary.BigEndian.Uint32(in[0:4])
	mid := binary.BigEndian.Uint16(in[4:6])
	lo := binary.BigEndian.Uint16(in[6:8]) & 0x0fff
	return (uint64(hi) << 28) | (uint64(mid) << 12) | uint64(lo)
}

func getClock14(in []byte) uint32 {
	return uint32(binary.BigEndian.Uint16(in[0:2]) & 0x3fff)
}

func getClock32(in []byte) (uint32, bool) {
	hi := binary.BigEndian.Uint16(in[0:2])
	mid := binary.BigEndian.Uint16(in[2:4])
	lo := in[4]
	ok := (mid & 0x3000) == 0x0000
	return (uint32(hi) << 20) | (uint32(mid) << 8) | uint32(lo), ok
}

func putUint48(out []byte, value uint64) {
	var tmp [8]byte
	binary.BigEndian.PutUint64(tmp[:], value)
	copy(out[0:6], tmp[2:8])
}

func putV1Ticks(out []byte, value uint64) {
	hi := uint16(value>>48) & 0x0fff
	mid := uint16(value >> 32)
	lo := uint32(value)

	binary.BigEndian.PutUint32(out[0:4], lo)
	binary.BigEndian.PutUint16(out[4:6], mid)
	binary.BigEndian.PutUint16(out[6:8], hi)
}

func putV6Ticks(out []byte, value uint64) {
	hi := uint32(value >> 28)
	mid := uint16(value >> 12)
	lo := uint16(value) & 0x0fff

	binary.BigEndian.PutUint32(out[0:4], hi)
	binary.BigEndian.PutUint16(out[4:6], mid)
	binary.BigEndian.PutUint16(out[6:8], lo)
}

func putClock14(out []byte, value uint32) {
	u16 := uint16(value & 0x3fff)
	binary.BigEndian.PutUint16(out[0:2], u16)
}

func putClock32(out []byte, value uint32) {
	hi := uint16(value>>20) & 0x0fff
	mid := uint16(value>>8) & 0x0fff
	lo := uint8(value)

	binary.BigEndian.PutUint16(out[0:2], hi)
	binary.BigEndian.PutUint16(out[2:4], mid)
	out[4] = lo
}
