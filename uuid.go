package youyouayedee

import (
	"encoding/binary"
	"fmt"
	"time"
)

// Size is the size of a UUID in bytes.
const Size = 16

// UUID represents a UUID.  Only the RFC 4122 variant is supported.
type UUID [Size]byte

// NilUUID is the nil UUID, "00000000-0000-0000-0000-000000000000".
var NilUUID = UUID{}

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

// Decode breaks down a UUID into its component fields.
//
// Only V1 and V6 UUIDs make use of the LeapSecondCalculator argument.  If it
// is required but nil, then a LeapSecondCalculatorDummy will be used instead.
//
func Decode(uuid UUID, lsc LeapSecondCalculator) Breakdown {
	var breakdown Breakdown
	var ticks uint64

	if !uuid.IsValid() {
		return breakdown
	}

	raw := [Size]byte(uuid)
	raw[6] = (raw[6] & 0x0f)
	raw[8] = (raw[8] & 0x3f)

	var scratch [8]byte

	if lsc == nil {
		lsc = LeapSecondCalculatorDummy{}
	}

	breakdown.Valid = true
	breakdown.Version = uuid.Version()

	switch breakdown.Version {
	case 1:
		ticks = uint64(binary.BigEndian.Uint32(raw[0:4]))
		ticks |= uint64(binary.BigEndian.Uint16(raw[4:6])) << 32
		ticks |= uint64(binary.BigEndian.Uint16(raw[6:8])) << 48
		breakdown.HasTicks = true
		breakdown.HasCounter = true
		breakdown.HasNode = true
		breakdown.Time = gregorianTicksToGoTime(lsc, ticks)
		breakdown.Ticks = int64(ticks)
		breakdown.Counter = int(binary.BigEndian.Uint16(raw[8:10]))
		copy(breakdown.Node[:], raw[10:16])

	case 6:
		ticks = uint64(binary.BigEndian.Uint32(raw[0:4])) << (32 - 4)
		ticks |= uint64(binary.BigEndian.Uint16(raw[4:6])) << (16 - 4)
		ticks |= uint64(binary.BigEndian.Uint16(raw[6:8]))
		breakdown.HasTicks = true
		breakdown.HasCounter = true
		breakdown.HasNode = true
		breakdown.Time = gregorianTicksToGoTime(lsc, ticks)
		breakdown.Ticks = int64(ticks)
		breakdown.Counter = int(binary.BigEndian.Uint16(raw[8:10]))
		copy(breakdown.Node[:], raw[10:16])

	case 7:
		copy(scratch[2:8], raw[0:6])
		ticks = binary.BigEndian.Uint64(scratch[0:8])
		ticks = uint64(signExtendUnixTicks(ticks))
		breakdown.HasTicks = true
		breakdown.HasData = true
		breakdown.Time = unixTicksToGoTime(int64(ticks))
		breakdown.Ticks = int64(ticks)
		breakdown.Data = make([]byte, 10)
		copy(breakdown.Data[0:10], raw[6:16])

	case 2:
		breakdown.HasDomainAndID = true
		breakdown.HasData = true
		breakdown.Domain = DCEDomain(raw[9])
		breakdown.ID = binary.BigEndian.Uint32(raw[0:4])
		breakdown.Data = make([]byte, 11)
		copy(breakdown.Data[0:5], raw[4:9])
		copy(breakdown.Data[5:11], raw[10:16])

	default:
		breakdown.HasData = true
		breakdown.Data = make([]byte, Size)
		copy(breakdown.Data, raw[:])
	}
	return breakdown
}

// Breakdown holds the results of decoding a UUID into its components.
type Breakdown struct {
	// Valid is true iff the UUID was successfully decoded to any degree.
	Valid bool

	// Version holds the detected UUID version.
	Version Version

	// HasTicks is true iff the Time and Ticks fields are valid.
	HasTicks bool

	// HasCounter is true iff the Counter field is valid.
	HasCounter bool

	// HasNode is true iff the Node field is valid.
	HasNode bool

	// HasData is true iff the Data field is valid.
	HasData bool

	// HasDomainAndID is true iff the Domain and ID fields are valid.
	HasDomainAndID bool

	// Time holds the decoded time extracted from a time-based UUID.
	Time time.Time

	// Ticks holds the raw tick count from a time-based UUID.
	//
	// For V1 and V6 UUIDs, this is hectonanoseconds since the start of the
	// Gregorian calendar in 1582, with leap seconds (probably) included.
	//
	// For V7 UUIDs, this is milliseconds since the start of the Unix epoch
	// in 1970, with leap seconds omitted.
	//
	Ticks int64

	// Counter holds the raw counter value from a time-based UUID.
	//
	// Only valid for V1 and V6 UUIDs.
	//
	Counter int

	// Node holds the node identifier from a time-based UUID.
	//
	// Only valid for V1 and V6 UUIDs.
	//
	Node Node

	// Domain holds the DCE domain.
	//
	// Only valid for V2 UUIDs.
	//
	Domain DCEDomain

	// ID holds the DCE identifier.
	//
	// Only valid for V2 UUIDs.
	//
	ID uint32

	// Data contains any additional random or opaque bytes, with the
	// variant and version bits cleared if applicable.
	//
	// For V1 and V6 UUIDs, this field is not used.
	//
	// For V7 UUIDs, this field contains all the bits from the UUID except
	// the timestamp.
	//
	// For V3, V4, V5, and V8 UUIDs, this field contains almost all bits
	// from the UUID.
	//
	Data []byte
}
