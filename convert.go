package youyouayedee

import (
	"golang.org/x/crypto/blake2b"
)

// Convert returns a copy of this UUID which has been converted from its
// current UUID version to the given UUID version.  Most version combinations
// are not possible.  This method will return ErrVersionMismatch for such
// combinations.
//
// The primary use for this method is to convert V1 UUIDs into V6 UUIDs and
// back, as this is the only pair for which bidirectional conversions exist.
// You can also convert V1/V6 into V7 unidirectionally, and you can convert
// absolutely any valid UUID into V8 unidirectionally.
//
// NilUUID and MaxUUID are always returned unmodified.
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

	expected := make([]Version, 0, 4)
	if current == 1 || current == 6 {
		expected = append(expected, 1, 6, 7, 8)
	} else {
		expected = append(expected, current, 8)
	}
	return uuid, ErrVersionMismatch{
		Requested: version,
		Expected:  expected,
	}
}
