package youyouayedee

import (
	"encoding/binary"
	"fmt"
	"time"

	"golang.org/x/crypto/blake2b"
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

// Decode breaks down this UUID into its component fields.
//
// Only V1 and V6 UUIDs make use of the LeapSecondCalculator argument.  If it
// is required but nil, then a LeapSecondCalculatorDummy will be used instead.
//
func (uuid UUID) Decode(lsc LeapSecondCalculator) Breakdown {
	var breakdown Breakdown
	var ticks uint64

	if !uuid.IsValid() {
		return breakdown
	}

	version := uuid.Version()
	uuid[6] = (uuid[6] & 0x0f)
	uuid[8] = (uuid[8] & 0x3f)

	if lsc == nil {
		lsc = LeapSecondCalculatorDummy{}
	}

	breakdown.Valid = true
	breakdown.Version = version

	switch breakdown.Version {
	case 1:
		ticks = getV1Ticks(uuid[0:8])
		breakdown.HasTicks = true
		breakdown.HasCounter = true
		breakdown.HasNode = true
		breakdown.Time = gregorianTicksToGoTime(lsc, ticks)
		breakdown.Ticks = int64(ticks)
		breakdown.Counter = int(getClock14(uuid[8:10]))
		copy(breakdown.Node[:], uuid[10:16])

	case 2:
		breakdown.HasDomainAndID = true
		breakdown.HasData = true
		breakdown.Domain = DCEDomain(uuid[9])
		breakdown.ID = binary.BigEndian.Uint32(uuid[0:4])
		breakdown.Data = make([]byte, 11)
		copy(breakdown.Data[0:5], uuid[4:9])
		copy(breakdown.Data[5:11], uuid[10:16])

	case 6:
		ticks = getV6Ticks(uuid[0:8])
		breakdown.HasTicks = true
		breakdown.HasCounter = true
		breakdown.HasNode = true
		breakdown.Time = gregorianTicksToGoTime(lsc, ticks)
		breakdown.Ticks = int64(ticks)
		breakdown.Counter = int(getClock14(uuid[8:10]))
		copy(breakdown.Node[:], uuid[10:16])

	case 7:
		ticks = getUint48(uuid[0:6])
		ticks = uint64(signExtendUnixTicks(ticks))
		breakdown.HasTicks = true
		breakdown.HasData = true
		breakdown.Time = unixTicksToGoTime(int64(ticks))
		breakdown.Ticks = int64(ticks)
		if clock, ok := getClock32(uuid[6:11]); ok {
			breakdown.HasCounter = true
			breakdown.Counter = int(clock)
			breakdown.Data = make([]byte, 5)
			copy(breakdown.Data[0:5], uuid[11:16])
		} else {
			breakdown.Data = make([]byte, 10)
			copy(breakdown.Data[0:10], uuid[6:16])
		}

	default:
		breakdown.HasData = true
		breakdown.Data = make([]byte, Size)
		copy(breakdown.Data, uuid[:])
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

// Convert returns a copy of this UUID which has been converted from its
// current UUID version to the given UUID version.  Most version combinations
// are not possible.  This method will return ErrVersionMismatch for such
// combinations.  The primary use for this method is to convert V1 UUIDs into
// V6 UUIDs and back, as this is the only pair for which bidirectional
// conversions exist.
//
// Only V1 and V6 UUIDs make use of the LeapSecondCalculator argument, and only
// when converting to V7 UUIDs.  If it is required but nil, then a
// LeapSecondCalculatorDummy will be used instead.
//
func (uuid UUID) Convert(version Version, lsc LeapSecondCalculator) (UUID, error) {
	if uuid.IsZero() || uuid.IsMax() {
		return uuid, nil
	}

	if !uuid.IsValid() {
		return uuid, ErrInputNotValid{Input: uuid}
	}

	current := uuid.Version()
	if version == current {
		return uuid, nil
	}
	if version == 8 {
		uuid[6] = (uuid[6] & 0x0f) | 0x80
		return uuid, nil
	}

	if lsc == nil {
		lsc = LeapSecondCalculatorDummy{}
	}

	var ok bool
	if current == 1 && version == 6 {
		ticks := getV1Ticks(uuid[0:8])
		putV6Ticks(uuid[0:8], ticks)
		ok = true
	}
	if current == 6 && version == 1 {
		ticks := getV6Ticks(uuid[0:8])
		putV1Ticks(uuid[0:8], ticks)
		ok = true
	}
	if current == 1 && version == 7 {
		ticks := getV1Ticks(uuid[0:8])
		clock := uint32(getClock14(uuid[8:10]))
		sum := blake2b.Sum256(uuid[:])

		putUint48(uuid[0:6], goTimeToUnixTicks(gregorianTicksToGoTime(lsc, ticks)))
		putClock32(uuid[6:11], clock)
		copy(uuid[11:16], sum[0:5])
		ok = true
	}
	if current == 6 && version == 7 {
		ticks := getV6Ticks(uuid[0:8])
		clock := uint32(getClock14(uuid[8:10]))
		sum := blake2b.Sum256(uuid[:])

		putUint48(uuid[0:6], goTimeToUnixTicks(gregorianTicksToGoTime(lsc, ticks)))
		putClock32(uuid[6:11], clock)
		copy(uuid[11:16], sum[0:5])
		ok = true
	}
	if ok {
		uuid[6] = (uuid[6] & 0x0f) | byte(version<<4)
		uuid[8] = (uuid[8] & 0x3f) | 0x80
		return uuid, nil
	}

	var expected []Version
	switch current {
	case 1:
		fallthrough
	case 6:
		expected = []Version{1, 6, 7, 8}
	default:
		expected = []Version{current, 8}
	}

	return uuid, ErrVersionMismatch{
		Requested: version,
		Expected:  expected,
	}
}
