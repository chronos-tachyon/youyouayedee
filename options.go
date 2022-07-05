package youyouayedee

import (
	"hash"
	"io"
	"time"
)

// Options supplies options for generating UUID values.
type Options struct {
	// Node indicates the node identifier to use.  The node identifier is
	// typically either a unique EUI-48 network hardware address associated
	// with the current host, or else a random 46-bit value modified to
	// look like a non-globally-unique EUI-48.
	//
	// Only time-based UUID generators use this field.  If this field is
	// the zero value but a generator requires it, then the generator will
	// query the local host during generator initialization by using the
	// GenerateNode function provided by this library.
	//
	// If the caller wishes to guarantee uniqueness between multiple
	// Generator instances for the same UUID version, it is the
	// responsibility of the caller to verify that a globally unique Node
	// can be obtained for the current host, or else to generate a random
	// node identifier and then preserve it for re-use, e.g. by writing it
	// to a file so that it is not forgotten across reboots and program
	// restarts.
	//
	Node Node

	// TimeSource is a function which returns the current time.
	//
	// Only time-based UUID generators use this field.  If this field is
	// nil but a generator requires it, then the built-in Go "time".Now
	// function is used instead.
	//
	TimeSource func() time.Time

	// LeapSecondCalculator provides the number of leap seconds by which
	// Unix time differs from the actual number of SI seconds since
	// 1970-01-01T00:00:00Z.
	//
	// Only V1 and V6 time-based UUID generators use this field.  If this
	// field is required but nil, an instance of LeapSecondCalculatorDummy
	// will be used instead.  The value is is needed to calculate the
	// RFC4122-correct timestamp values while generating V1 and V6
	// time-based UUIDs.
	//
	LeapSecondCalculator LeapSecondCalculator

	// ClockStorage provides persistent storage for clock counters.
	//
	// Only time-based UUID generators use this field.  If it is nil but a
	// generator requires it, then ClockStorageUnavailable is used instead.
	// This will degrade the uniqueness guarantees provided by those
	// generators; if the caller needs stronger guarantees, they must
	// provide their own ClockStorage instance that meets their
	// requirements.
	//
	ClockStorage ClockStorage

	// Namespace is the base UUID for namespacing data inputs when hashing.
	//
	// Only hash-based UUID generators use this field, but for those UUID
	// generators it is a mandatory field.  Generator initialization will
	// fail if this field is not initialized to a valid UUID.
	//
	Namespace UUID

	// HashFactory is a callback to produce new instances of hash.Hash on demand.
	//
	// Only hash-based UUID generators for V8 UUIDs use this field, and for
	// that case the field is mandatory.  V3 and V5 UUID generators ignore
	// this field and always use md5.New or sha1.New, respectively.
	//
	HashFactory func() hash.Hash

	// ForceRandomNode controls the behavior of GenerateNode when a node
	// identifier is required but Node is the zero value.
	//
	ForceRandomNode bool

	// RandomSource specifies a source of random bytes.
	//
	// Both random-based and time-based UUID generators use this field,
	// although the latter only use it to generate a node identifier if one
	// cannot otherwise be obtained.  If this field is nil but a source of
	// random bytes is required, then "crypto/rand".Reader will be used
	// instead.
	//
	RandomSource io.Reader
}
