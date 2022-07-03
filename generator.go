package uuid

import (
	"crypto/md5"
	"crypto/sha1"
	"io"
	"time"
)

// Generator is an interface for generating new UUID values.
type Generator interface {
	// NewUUID generates a new unpredictable UUID.
	//
	// Generators are not required to support this operation, and should
	// return MustHashError if only NewHashUUID is implemented.
	//
	NewUUID() (UUID, error)

	// NewHashUUID generates a new deterministic UUID by hashing the given
	// input data.
	//
	// Generators are not required to support this operation, and should
	// return MustNotHashError if only NewUUID is implemented.
	//
	NewHashUUID(data []byte) (UUID, error)
}

// GeneratorOptions supplies options for generating UUID values.
type GeneratorOptions struct {
	// Node represents the node identifier.  The node identifier is
	// typically either a unique EUI-48 network hardware address associated
	// with the current host, or else a random 46-bit value modified to
	// look like a non-globally-unique EUI-48.
	//
	// Only some UUID generators require this field.  If this field is the
	// zero value but a generator requires it, then the generator will
	// query the local host during generator initialization by using the
	// GenerateNode function.
	//
	Node Node

	// TimeSource is a function which returns the current time.
	//
	// Only some UUID generators require this field.  If it is nil but a
	// generator requires it, then the built-in "time".Now function is
	// used instead.
	//
	TimeSource func() time.Time

	// LeapSecondCalculator provides the number of leap seconds by which
	// Unix time differs from the number of SI seconds since
	// 1970-01-01T00:00:00Z.  This is needed to calculate the correct
	// timestamps while generating V1 and V6 time-based UUIDs.
	//
	// If LeapSecondCalculator is required but nil, an instance of
	// DummyLeapSecondCalculator will be used instead.
	//
	LeapSecondCalculator LeapSecondCalculator

	// ClockStorage provides persistent storage for clock counters.
	//
	// Only some UUID generators require this field.  If it is nil but a
	// generator requires it, then UnavailableClockStorage is used instead.
	//
	ClockStorage ClockStorage

	// Namespace is the base UUID for namespacing data inputs when hashing.
	//
	// Only some UUID generators require this field.  Those UUID generators
	// will fail during initialization if this field is not initialized to
	// a valid UUID.
	//
	Namespace UUID

	// ForceRandomNode controls the behavior of GenerateNode when Node is
	// required but zero.
	//
	ForceRandomNode bool

	// RandomSource specifies a source of random bytes.  If this field is
	// nil but a source of random bytes is required, then
	// "crypto/rand".Reader will be used instead.
	//
	RandomSource io.Reader
}

// NewGenerator initializes a new Generator instance for the given UUID version.
//
// If this library does not know how to generate UUIDs of the given version,
// then UnsupportedVersionError is returned.
//
func NewGenerator(version Version, o GeneratorOptions) (Generator, error) {
	if factory, found := GeneratorsByVersion[version]; found {
		return factory.NewGenerator(version, o)
	}
	switch version {
	case 1:
		return NewTimeGenerator(1, o)
	case 3:
		return NewHashGenerator(3, md5.New, o)
	case 4:
		return NewRandomGenerator(4, o)
	case 5:
		return NewHashGenerator(5, sha1.New, o)
	case 6:
		return NewTimeGenerator(6, o)
	case 7:
		return NewTimeGenerator(7, o)
	}
	return nil, UnsupportedVersionError{Version: version}
}

// GeneratorFactory is an interface for constructing Generator instances.
//
// It is used to create hooks placed in GeneratorsByVersion, in order to modify
// the behavior of this library's NewGenerator function.
//
type GeneratorFactory interface {
	NewGenerator(Version, GeneratorOptions) (Generator, error)
}

// GeneratorFactoryFunc implements GeneratorFactory with a function.
type GeneratorFactoryFunc func(GeneratorOptions) (Generator, error)

func (fn GeneratorFactoryFunc) NewGenerator(version Version, o GeneratorOptions) (Generator, error) {
	return fn(o)
}

var _ GeneratorFactory = GeneratorFactoryFunc(nil)

// GeneratorsByVersion provides a hook for the NewGenerator function to
// construct user-defined Generator instances.
//
var GeneratorsByVersion map[Version]GeneratorFactory
