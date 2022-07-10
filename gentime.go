package youyouayedee

import (
	"encoding/binary"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/crypto/blake2b"
)

// NewTimeGenerator constructs a new Generator that produces time-based UUIDs
// of the given version.
//
// Versions 1, 6, 7, and 8 are supported.
//
func NewTimeGenerator(version Version, o Options) (Generator, error) {
	var err error

	if version != 1 && version != 6 && version != 7 && version != 8 {
		return nil, ErrVersionMismatch{Requested: version, Expected: []Version{1, 6, 7, 8}}
	}

	node := o.Node
	if node.IsZero() {
		node, err = GenerateNode(o)
		if err != nil {
			return nil, ErrOperationFailed{Operation: GenerateNodeOp, Err: err}
		}
	}

	now := o.TimeSource
	if now == nil {
		now = time.Now
	}

	lsc := o.LeapSecondCalculator
	if lsc == nil {
		lsc = LeapSecondCalculatorDummy{}
	}

	cs := o.ClockStorage
	if cs == nil {
		cs = ClockStorageUnavailable{}
	}

	last, clock, err := cs.Load(node)
	if err != nil {
		if !isErrClockNotFound(err) {
			return nil, ErrOperationFailed{Operation: ClockStorageLoadOp, Err: err}
		}

		last = now()
		clock = rand.Uint32()
	}

	return &genTime{
		node:  node,
		now:   now,
		lsc:   lsc,
		cs:    cs,
		ver:   version,
		last:  last,
		clock: clock,
	}, nil
}

type genTime struct {
	GeneratorBase

	node  Node
	now   func() time.Time
	lsc   LeapSecondCalculator
	cs    ClockStorage
	ver   Version
	mu    sync.Mutex
	last  time.Time
	clock uint32
}

func (g *genTime) NewUUID() (UUID, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := g.now()

	if g.last.Before(now) {
		g.last = now
	} else {
		now = g.last
		g.clock++
	}

	err := g.cs.Store(g.node, g.last, g.clock)
	if err != nil {
		return Nil, ErrOperationFailed{Operation: ClockStorageStoreOp, Err: err}
	}

	var uuid UUID
	var ticks uint64

	if g.ver == 1 {
		ticks = goTimeToGregorianTicks(g.lsc, now)
		putV1Ticks(uuid[0:8], ticks)
		putClock14(uuid[8:10], g.clock)
		copy(uuid[10:16], g.node[0:6])
	} else if g.ver == 6 {
		ticks = goTimeToGregorianTicks(g.lsc, now)
		putV6Ticks(uuid[0:8], ticks)
		putClock14(uuid[8:10], g.clock)
		copy(uuid[10:16], g.node[0:6])
	} else {
		ticks = goTimeToUnixTicks(now)

		var hashInput [18]byte
		binary.BigEndian.PutUint64(hashInput[0:8], ticks)
		binary.BigEndian.PutUint32(hashInput[8:12], g.clock)
		copy(hashInput[12:18], g.node[0:6])
		sum := blake2b.Sum256(hashInput[:])

		putUint48(uuid[0:6], ticks)
		putClock32(uuid[6:11], g.clock)
		copy(uuid[11:16], sum[0:5])
	}

	uuid[6] = (uuid[6] & 0x0f) | byte(g.ver<<4)
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return uuid, nil
}

var _ Generator = (*genTime)(nil)

func goTimeToGregorianTicks(lsc LeapSecondCalculator, now time.Time) uint64 {
	s := now.Unix()
	ns := now.Nanosecond()

	if ns < 0 {
		ns = nanosPerSecond - ns
		s--
	}

	s += int64(lsc.LeapSecondsSinceUnixEpoch(s, false))
	s += secondsFromGregorianEpochToUnixEpoch

	if s < 0 {
		s = 0
	}

	num := uint64(s*ticksPerSecond) + uint64(ns/nanosPerTick)
	num &= tickMask
	return num
}

func gregorianTicksToGoTime(lsc LeapSecondCalculator, num uint64) time.Time {
	num &= tickMask

	ns := int64(num%ticksPerSecond) * nanosPerTick
	s := int64(num / ticksPerSecond)

	s -= secondsFromGregorianEpochToUnixEpoch
	s -= int64(lsc.LeapSecondsSinceUnixEpoch(s, true))

	if s < 0 && ns > 0 {
		ns = nanosPerSecond - ns
		s++
	}

	return time.Unix(s, ns)
}

func goTimeToUnixTicks(now time.Time) uint64 {
	s := now.Unix()
	ns := now.Nanosecond()

	if ns < 0 {
		ns = nanosPerSecond - ns
		s--
	}

	s64 := (s * millisPerSecond) + int64(ns/nanosPerMilli)
	return uint64(s64) & milliMask
}

func signExtendUnixTicks(num uint64) int64 {
	num &= milliMask
	if (num & milliSignBit) == milliSignBit {
		num |= ^uint64(milliMask)
	}
	return int64(num)
}

func unixTicksToGoTime(num int64) time.Time {
	ns := (num % millisPerSecond) * nanosPerMilli
	s := (num / millisPerSecond)

	if s < 0 && ns > 0 {
		ns = nanosPerSecond - ns
		s++
	}

	return time.Unix(s, ns)
}
