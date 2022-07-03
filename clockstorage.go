package youyouayedee

import (
	"time"
)

// ClockStorage provides an interface for loading and storing clock counters on
// behalf of UUID generators.
//
// UUID generators based on timestamps (versions 1, 6, and 7) need some form of
// persistent storage in order to prevent collisions.  This interface provides
// that persistent storage.
//
type ClockStorage interface {
	// Load retrieves the last known timestamp and the last known counter
	// value for the given Node.
	//
	// If the implementation returns ClockStorageUnavailableError for any
	// reason, the UUID generator is required to use the current time as
	// the last known timestamp and to generate a new random counter value
	// from scratch.
	//
	// Implementations are not required to retain a (timestamp, counter)
	// tuple for each node; if there is no tuple stored for the given Node,
	// then ClockStorageUnavailableError is the best choice of return
	// value.
	//
	Load(Node) (time.Time, uint32, error)

	// Store associates the given last known timestamp and last known
	// counter value with the given Node, synchronizing the data to
	// persistent storage.
	//
	// Implementations are free to discard the (timestamp, counter) tuples
	// associated with any previous node identifiers upon receiving a Store
	// method call with a new node identifier.  However, it may be
	// beneficial to retain one such tuple for each past node identifier,
	// at least when such node identifiers are flagged as both IsGlobal and
	// IsUnicast.
	//
	Store(Node, time.Time, uint32) error
}

// UnavailableClockStorage is a dummy implementation of ClockStorage that does
// not store anything.
type UnavailableClockStorage struct{}

func (UnavailableClockStorage) Load(Node) (time.Time, uint32, error) {
	return time.Time{}, 0, ClockStorageUnavailableError{}
}

func (UnavailableClockStorage) Store(Node, time.Time, uint32) error {
	return nil
}

var _ ClockStorage = UnavailableClockStorage{}
