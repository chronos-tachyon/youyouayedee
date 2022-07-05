package youyouayedee

// Generator is an interface for generating new UUID values.
type Generator interface {
	// NewUUID generates a new unpredictable UUID.
	//
	// Generators are not required to support this operation, and should
	// return ErrMethodNotSupported{MethodNewUUID} if it is not.
	//
	NewUUID() (UUID, error)

	// NewHashUUID generates a new deterministic UUID by hashing the given
	// input data.
	//
	// Generators are not required to support this operation, and should
	// return ErrMethodNotSupported{MethodNewHashUUID} if it is not.
	//
	NewHashUUID(data []byte) (UUID, error)
}

// NewGenerator initializes a new Generator instance for the given UUID version.
//
// If this library does not know how to generate UUIDs of the given version,
// then ErrVersionNotSupported is returned.
//
// The GeneratorsByVersion global variable may be used to override this
// function's default behavior for chosen UUID versions.
//
func NewGenerator(version Version, o Options) (Generator, error) {
	if factory, found := GeneratorsByVersion[version]; found {
		return factory.NewGenerator(version, o)
	}
	switch version {
	case 1:
		return NewTimeGenerator(1, o)
	case 3:
		return NewHashGenerator(3, o)
	case 4:
		return NewRandomGenerator(4, o)
	case 5:
		return NewHashGenerator(5, o)
	case 6:
		return NewTimeGenerator(6, o)
	case 7:
		return NewTimeGenerator(7, o)
	}
	return nil, ErrVersionNotSupported{Version: version}
}

// GeneratorBase is a base Generator implementation that you can embed into the
// struct of your custom Generator implementation.  All Generator methods will
// then have default implementations that simply return ErrMethodNotSupported.
//
// This is a very easy way to future-proof your code for resiliency during
// future major version updates affecting the Generator interface.
//
type GeneratorBase struct{}

func (GeneratorBase) NewUUID() (UUID, error) {
	return NilUUID, ErrMethodNotSupported{Method: MethodNewUUID}
}

func (GeneratorBase) NewHashUUID(data []byte) (UUID, error) {
	return NilUUID, ErrMethodNotSupported{Method: MethodNewHashUUID}
}

var _ Generator = GeneratorBase{}

// GeneratorFactory is an interface for constructing Generator instances.
//
// It is used to create hooks placed in GeneratorsByVersion, in order to modify
// the behavior of this library's NewGenerator function.
//
type GeneratorFactory interface {
	NewGenerator(Version, Options) (Generator, error)
}

// GeneratorFactoryFunc implements GeneratorFactory with a function.
type GeneratorFactoryFunc func(Version, Options) (Generator, error)

func (fn GeneratorFactoryFunc) NewGenerator(version Version, o Options) (Generator, error) {
	return fn(version, o)
}

var _ GeneratorFactory = GeneratorFactoryFunc(nil)

// GeneratorsByVersion provides a hook for the NewGenerator function to
// construct user-defined Generator instances.
//
var GeneratorsByVersion = make(map[Version]GeneratorFactory)
