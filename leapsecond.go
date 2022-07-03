package youyouayedee

// LeapSecondCalculator is an interface for calculating the number of leap
// seconds by which Unix time differs from the number of SI seconds since
// 1970-01-01T00:00:00Z.
//
// Leap seconds cannot be predicted sooner than about six months or so ahead of
// time, so a proper implementation must source this data from the
// International Earth Rotation and Reference Systems Service (IERS), or from
// an organization which redistributes it in a timely manner such as IETF or
// IANA, and it must be updated regularly.  On Linux hosts, this data is
// usually provided as part of your distribution's regularly updated timezone
// database.  Other operating systems may vary.
//
type LeapSecondCalculator interface {
	LeapSecondsSinceUnixEpoch(seconds int64, includesLeapSeconds bool) int
}

// DummyLeapSecondCalculator falsely claims that there has never been a leap
// second.  Although this is not actually true, many existing Version 1 UUID
// generators behave as though this is the case.
//
type DummyLeapSecondCalculator struct{}

func (DummyLeapSecondCalculator) LeapSecondsSinceUnixEpoch(seconds int64, includesLeapSeconds bool) int {
	return 0
}

var _ LeapSecondCalculator = DummyLeapSecondCalculator{}

// FixedLeapSecondCalculator makes a best effort calculation of the number of
// leap seconds that have elapsed before a given time.  The list of leap
// seconds is statically compiled into the library and is subject to updates.
//
// In particular, although historical leap seconds never change once recorded,
// differences may arise between past computations of a given recent date and
// present computations if this library is updated between computations.
//
type FixedLeapSecondCalculator struct{}

func (FixedLeapSecondCalculator) LeapSecondsSinceUnixEpoch(seconds int64, includesLeapSeconds bool) int {

	// NB: there are corner cases at each leap second boundary.
	//
	// Actual positive leap second:
	//
	//     UTC Time              Unix seconds    SI seconds  "Total" field
	//     --------------------  ------------  ------------  -------------
	//     2016-12-31T23:59:59Z    1483228799    1483228825             26
	//     2016-12-31T23:59:60Z             -    1483228826             26
	//     2017-01-01T00:00:00Z    1483228800    1483228827             27
	//
	//     This affects (includesLeapSeconds == true).
	//
	// Hypothetical negative leap second:
	//
	//     UTC Time              Unix seconds    SI seconds  "Total" field
	//     --------------------  ------------  ------------  -------------
	//     2016-12-31T23:59:58Z    1483228798    1483228825             27
	//                        -    1483228799             -             27
	//     2017-01-01T00:00:00Z    1483228800    1483228826             26
	//
	//     This affects (includesLeapSeconds == false).

	tableLen := uint(len(fixedLeapSecondTable))
	ti := uint(1)
	total := int(0)
	for ti < tableLen {
		row := fixedLeapSecondTable[ti]
		t := row.Time
		if includesLeapSeconds {
			minTotal := row.Total
			if minTotal > total {
				minTotal = total
			}
			t += int64(minTotal)
		}
		total = row.Total
		if seconds < t {
			break
		}
		ti++
	}
	return total
}

var _ LeapSecondCalculator = FixedLeapSecondCalculator{}

type leapSecond struct {
	Time  int64
	Total int
}

// Data captured from IETF timezone data file "leap-seconds.list".
//
// Retrieved from https://www.ietf.org/timezones/data/leap-seconds.list on
// 2022-07-02 and converted to the Unix time_t epoch.
//
var fixedLeapSecondTable = [...]leapSecond{
	{Time: -(1 << 63), Total: 0},
	{Time: 0x03c26700, Total: 1},  // 1 Jan 1972
	{Time: 0x04b25800, Total: 2},  // 1 Jul 1972
	{Time: 0x05a4ec00, Total: 3},  // 1 Jan 1973
	{Time: 0x07861f80, Total: 4},  // 1 Jan 1974
	{Time: 0x09675300, Total: 5},  // 1 Jan 1975
	{Time: 0x0b488680, Total: 6},  // 1 Jan 1976
	{Time: 0x0d2b0b80, Total: 7},  // 1 Jan 1977
	{Time: 0x0f0c3f00, Total: 8},  // 1 Jan 1978
	{Time: 0x10ed7280, Total: 9},  // 1 Jan 1979
	{Time: 0x12cea600, Total: 10}, // 1 Jan 1980
	{Time: 0x159fca80, Total: 11}, // 1 Jul 1981
	{Time: 0x1780fe00, Total: 12}, // 1 Jul 1982
	{Time: 0x19623180, Total: 13}, // 1 Jul 1983
	{Time: 0x1d25ea00, Total: 14}, // 1 Jul 1985
	{Time: 0x21dae500, Total: 15}, // 1 Jan 1988
	{Time: 0x259e9d80, Total: 16}, // 1 Jan 1990
	{Time: 0x277fd100, Total: 17}, // 1 Jan 1991
	{Time: 0x2a50f580, Total: 18}, // 1 Jul 1992
	{Time: 0x2c322900, Total: 19}, // 1 Jul 1993
	{Time: 0x2e135c80, Total: 20}, // 1 Jul 1994
	{Time: 0x30e72400, Total: 21}, // 1 Jan 1996
	{Time: 0x33b84880, Total: 22}, // 1 Jul 1997
	{Time: 0x368c1000, Total: 23}, // 1 Jan 1999
	{Time: 0x43b71b80, Total: 24}, // 1 Jan 2006
	{Time: 0x495c0780, Total: 25}, // 1 Jan 2009
	{Time: 0x4fef9300, Total: 26}, // 1 Jul 2012
	{Time: 0x55932d80, Total: 27}, // 1 Jul 2015
	{Time: 0x58684680, Total: 28}, // 1 Jan 2017
}
