package youyouayedee

const (
	tickBits = 60
	tickMask = (1 << tickBits) - 1

	clockBits = 14
	clockMask = (1 << clockBits) - 1

	milliBits    = 48
	milliMask    = (1 << milliBits) - 1
	milliSignBit = 1 << (milliBits - 1)

	daysFromGregorianEpochToUnixEpoch    = 141427
	secondsFromGregorianEpochToUnixEpoch = daysFromGregorianEpochToUnixEpoch * 86400

	nanosPerSecond  = 1000000000
	nanosPerMilli   = 1000000
	nanosPerTick    = 100
	millisPerSecond = nanosPerSecond / nanosPerMilli
	ticksPerSecond  = nanosPerSecond / nanosPerTick
)
