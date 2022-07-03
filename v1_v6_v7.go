package youyouayedee

import (
	"encoding/binary"
	"hash"
	mathrand "math/rand"
	"sync"
	"time"

	"golang.org/x/crypto/blake2b"
)

type genTime struct {
	BaseGenerator

	node  Node
	now   func() time.Time
	lsc   LeapSecondCalculator
	cs    ClockStorage
	ver   Version
	mu    sync.Mutex
	h     hash.Hash
	last  time.Time
	clock uint32
}

// NewTimeGenerator initializes a new Generator that produces time-based UUIDs
// of the given version.
//
// Versions 1, 6, 7, and 8 are supported.
//
func NewTimeGenerator(version Version, o GeneratorOptions) (Generator, error) {
	var err error

	var needHash bool
	switch version {
	case 1:
		needHash = false

	case 6:
		needHash = false

	case 7:
		needHash = true

	case 8:
		needHash = true

	default:
		return nil, MismatchedVersionError{Requested: version, Expected: []Version{1, 6, 7, 8}}
	}

	node := o.Node
	if node.IsZero() {
		node, err = GenerateNode(NodeOptions{
			ForceRandomNode: o.ForceRandomNode,
			RandomSource:    o.RandomSource,
		})
		if err != nil {
			return nil, FailedOperationError{Operation: GenerateNodeOp, Err: err}
		}
	}

	now := o.TimeSource
	if now == nil {
		now = time.Now
	}

	lsc := o.LeapSecondCalculator
	if lsc == nil {
		lsc = DummyLeapSecondCalculator{}
	}

	cs := o.ClockStorage
	if cs == nil {
		cs = UnavailableClockStorage{}
	}

	var h hash.Hash
	if needHash {
		h, err = blake2b.New256(node[:])
		if err != nil {
			return nil, FailedOperationError{Operation: InitializeBlakeHashOp, Err: err}
		}
	}

	last, clock, err := cs.Load(node)
	if err != nil {
		if !isClockStorageUnavailable(err) {
			return nil, FailedOperationError{Operation: ClockStorageLoadOp, Err: err}
		}

		last = now()
		clock = mathrand.Uint32()
	}

	return &genTime{
		node:  node,
		now:   now,
		lsc:   lsc,
		cs:    cs,
		ver:   version,
		h:     h,
		last:  last,
		clock: clock,
	}, nil
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
		return NilUUID, FailedOperationError{Operation: ClockStorageStoreOp, Err: err}
	}

	var uuid UUID
	var hashInput [12]byte
	var hashOutput [blake2b.Size256]byte
	var sum []byte
	var ticks uint64

	switch g.ver {
	case 1:
		ticks = goTimeToGregorianTicks(g.lsc, now)

		binary.BigEndian.PutUint32(uuid[0:4], uint32(ticks))
		binary.BigEndian.PutUint16(uuid[4:6], uint16(ticks>>32))
		binary.BigEndian.PutUint16(uuid[6:8], uint16(ticks>>48))
		binary.BigEndian.PutUint16(uuid[8:10], uint16(g.clock))
		copy(uuid[10:16], g.node[0:6])

	case 6:
		ticks = goTimeToGregorianTicks(g.lsc, now)
		binary.BigEndian.PutUint32(uuid[0:4], uint32(ticks>>(32-4)))
		binary.BigEndian.PutUint16(uuid[4:6], uint16(ticks>>(16-4)))
		binary.BigEndian.PutUint16(uuid[6:8], uint16(ticks))
		binary.BigEndian.PutUint16(uuid[8:10], uint16(g.clock))
		copy(uuid[10:16], g.node[0:6])

	case 7:
		fallthrough
	case 8:
		ticks = goTimeToUnixTicks(now)

		binary.BigEndian.PutUint64(hashInput[0:8], ticks)
		binary.BigEndian.PutUint32(hashInput[8:12], g.clock)

		g.h.Reset()
		_, _ = g.h.Write(hashInput[:])
		sum = g.h.Sum(hashOutput[:])

		copy(uuid[0:6], hashInput[2:8])
		copy(uuid[6:16], sum[0:10])

	default:
		panic("bad version")
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

	s += int64(lsc.LeapSecondsSinceUnixEpoch(s))
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
	s -= int64(lsc.LeapSecondsSinceUnixEpoch(s))

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

func unixTicksToGoTime(num uint64) time.Time {
	num &= milliMask
	if (num & milliSignBit) == milliSignBit {
		num |= ^uint64(milliMask)
	}
	s64 := int64(num)

	ns := (s64 % millisPerSecond) * nanosPerMilli
	s := (s64 / millisPerSecond)

	if s < 0 && ns > 0 {
		ns = nanosPerSecond - ns
		s++
	}

	return time.Unix(s, ns)
}
