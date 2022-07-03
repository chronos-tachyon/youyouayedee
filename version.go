package uuid

import (
	"fmt"
)

// Version represents the version number of a UUID.
type Version byte

// IsValid returns true iff this Version is a UUID version known to this library.
func (version Version) IsValid() bool {
	return version >= 1 && version <= 8
}

// GoString returns a developer-friendly string representation.
func (version Version) GoString() string {
	return fmt.Sprintf("uuid.Version(%d)", byte(version))
}

// String returns a human-friendly string representation.
func (version Version) String() string {
	return fmt.Sprintf("Version %d", byte(version))
}

var (
	_ fmt.GoStringer = Version(0)
	_ fmt.Stringer   = Version(0)
)
