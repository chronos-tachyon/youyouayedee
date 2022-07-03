package uuid

import (
	"crypto/rand"
	"errors"
	"io"
)

func readRandom(rng io.Reader, out []byte) error {
	if rng == nil {
		rng = rand.Reader
	}

	_, err := io.ReadFull(rng, out)
	return err
}

func isClockStorageUnavailable(err error) bool {
	var unavailable ClockStorageUnavailableError
	return errors.Is(err, &unavailable)
}
