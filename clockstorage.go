package youyouayedee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"
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

// FileClockStorage is an implementation of ClockStorage that reads from and
// writes to a file while holding a lock.
type FileClockStorage struct {
	mu     sync.Mutex
	name   string
	file   *os.File
	data   map[string]*clockRow
	closed bool
}

type clockRow struct {
	Time    time.Time `json:"time"`
	Counter uint32    `json:"counter"`
}

// OpenClockStorageFile constructs an instance of FileClockStorage.
func OpenClockStorageFile(fileName string) (*FileClockStorage, error) {
	if !lockFileSupported {
		return nil, fmt.Errorf("clock sequence files must be locked for exclusive access, but package youyouayedee doesn't know how to lock files on your OS")
	}

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open clock sequence file: %q: %w", fileName, err)
	}

	needClose := true
	defer func() {
		if needClose {
			_ = f.Close()
		}
	}()

	err = lockFile(f)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire exclusive lock on clock sequence file: %q: %w", fileName, err)
	}

	cs := &FileClockStorage{
		name: fileName,
		file: f,
	}

	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read clock sequence data from file: %q: %w", fileName, err)
	}

	if len(raw) == 0 {
		cs.data = make(map[string]*clockRow)
	} else {
		err = json.Unmarshal(raw, &cs.data)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal clock sequence data from file as JSON: %q: %w", fileName, err)
		}
	}

	needClose = false
	return cs, nil
}

func (cs *FileClockStorage) Load(node Node) (time.Time, uint32, error) {
	if cs == nil {
		return time.Time{}, 0, ClockStorageUnavailableError{}
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.file == nil || cs.closed {
		return time.Time{}, 0, fs.ErrClosed
	}

	key := node.String()
	if row, found := cs.data[key]; found {
		return row.Time, row.Counter, nil
	}
	return time.Time{}, 0, ClockStorageUnavailableError{}
}

func (cs *FileClockStorage) Store(node Node, t time.Time, c uint32) error {
	if cs == nil {
		return nil
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.file == nil || cs.closed {
		return fs.ErrClosed
	}

	key := node.String()
	row := cs.data[key]
	if row == nil {
		row = new(clockRow)
		cs.data[key] = row
	}
	row.Time = t
	row.Counter = c

	// Delete all IsLocal && IsMulticast keys except the current key.
	//
	// This reduces clutter in the database if the node identifier is being
	// generated at random each time the application runs.
	//
	for rowKey := range cs.data {
		if rowKey != key && len(rowKey) >= 2 && bytes.ContainsAny(rowKey[1], "37bf") {
			delete(cs.data, rowKey)
		}
	}

	raw, err := json.Marshal(cs.data)
	if err != nil {
		return fmt.Errorf("failed to marshal clock sequence data to JSON: %q: %w", cs.name, err)
	}

	_, err = cs.file.Seek(0, os.SEEK_SET)
	if err != nil {
		return fmt.Errorf("failed to seek to start of the clock sequence file: %q: %w", cs.name, err)
	}

	err = cs.file.Truncate(0)
	if err != nil {
		return fmt.Errorf("failed to truncate the clock sequence file to zero bytes: %q: %w", cs.name, err)
	}

	_, err = cs.file.Write(raw)
	if err != nil {
		return fmt.Errorf("failed to write clock sequence data to file: %q: %w", cs.name, err)
	}

	err = cs.file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync clock sequence data to disk: %q: %w", cs.name, err)
	}

	return nil
}

func (cs *FileClockStorage) Close() error {
	if cs == nil {
		return nil
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.file == nil || cs.closed {
		return fs.ErrClosed
	}

	cs.closed = true
	return cs.file.Close()
}

var (
	_ ClockStorage = (*FileClockStorage)(nil)
	_ io.Closer    = (*FileClockStorage)(nil)
)
