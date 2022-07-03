package youyouayedee

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
	// field is required but nil, an instance of DummyLeapSecondCalculator
	// will be used instead.  The value is is needed to calculate the
	// RFC4122-correct timestamp values while generating V1 and V6
	// time-based UUIDs.
	//
	LeapSecondCalculator LeapSecondCalculator

	// ClockStorage provides persistent storage for clock counters.
	//
	// Only time-based UUID generators use this field.  If it is nil but a
	// generator requires it, then UnavailableClockStorage is used instead.
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

// NewGenerator initializes a new Generator instance for the given UUID version.
//
// If this library does not know how to generate UUIDs of the given version,
// then UnsupportedVersionError is returned.
//
// The GeneratorsByVersion global variable may be used to override this
// function's default behavior for chosen UUID versions.
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

// BaseGenerator is a dummy Generator implementation that you can embed into
// your Generator implementation's struct so that your Generator will have
// default implementations of all methods that simply return
// MethodNotSupportedError.
//
// This is a very easy way to future-proof your code for resiliency during
// future major version updates affecting the Generator interface.
//
type BaseGenerator struct{}

func (BaseGenerator) NewUUID() (UUID, error) {
	return NilUUID, MethodNotSupportedError{Method: MethodNewUUID}
}

func (BaseGenerator) NewHashUUID(data []byte) (UUID, error) {
	return NilUUID, MethodNotSupportedError{Method: MethodNewHashUUID}
}

var _ Generator = BaseGenerator{}

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
var GeneratorsByVersion = make(map[Version]GeneratorFactory)
