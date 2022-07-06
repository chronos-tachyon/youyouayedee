package youyouayedee

import (
	"encoding/binary"
	"time"
)

// Decode breaks down this UUID into its component fields.
//
// Only V1 and V6 UUIDs make use of the LeapSecondCalculator argument.  If it
// is required but nil, then a LeapSecondCalculatorDummy will be used instead.
//
func (uuid UUID) Decode(lsc LeapSecondCalculator) Decoded {
	var result Decoded
	var ticks uint64

	if !uuid.IsValid() {
		return result
	}

	version := uuid.Version()
	uuid[6] = (uuid[6] & 0x0f)
	uuid[8] = (uuid[8] & 0x3f)

	if lsc == nil {
		lsc = LeapSecondCalculatorDummy{}
	}

	result.Valid = true
	result.Version = version

	switch result.Version {
	case 1:
		ticks = getV1Ticks(uuid[0:8])
		result.HasTicks = true
		result.HasCounter = true
		result.HasNode = true
		result.Time = gregorianTicksToGoTime(lsc, ticks)
		result.Ticks = int64(ticks)
		result.Counter = int(getClock14(uuid[8:10]))
		copy(result.Node[:], uuid[10:16])

	case 2:
		result.HasDomainAndID = true
		result.HasData = true
		result.Domain = DCEDomain(uuid[9])
		result.ID = binary.BigEndian.Uint32(uuid[0:4])
		result.Data = make([]byte, 11)
		copy(result.Data[0:5], uuid[4:9])
		copy(result.Data[5:11], uuid[10:16])

	case 6:
		ticks = getV6Ticks(uuid[0:8])
		result.HasTicks = true
		result.HasCounter = true
		result.HasNode = true
		result.Time = gregorianTicksToGoTime(lsc, ticks)
		result.Ticks = int64(ticks)
		result.Counter = int(getClock14(uuid[8:10]))
		copy(result.Node[:], uuid[10:16])

	case 7:
		ticks = getUint48(uuid[0:6])
		ticks = uint64(signExtendUnixTicks(ticks))
		result.HasTicks = true
		result.HasData = true
		result.Time = unixTicksToGoTime(int64(ticks))
		result.Ticks = int64(ticks)
		if clock, ok := getClock32(uuid[6:11]); ok {
			result.HasCounter = true
			result.Counter = int(clock)
			result.Data = make([]byte, 5)
			copy(result.Data[0:5], uuid[11:16])
		} else {
			result.Data = make([]byte, 10)
			copy(result.Data[0:10], uuid[6:16])
		}

	default:
		result.HasData = true
		result.Data = make([]byte, Size)
		copy(result.Data, uuid[:])
	}
	return result
}

// Decoded holds the results of decoding a UUID into its components.
type Decoded struct {
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
