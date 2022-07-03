package uuid

import (
	"io"
)

type genRandom struct {
	rng io.Reader
	ver Version
}

// NewRandomGenerator initializes a new Generator that produces new randomly
// generated UUIDs.
//
// Versions 4 and 8 are supported.
//
func NewRandomGenerator(version Version, o GeneratorOptions) (Generator, error) {
	switch version {
	case 4:
		// pass

	case 8:
		// pass

	default:
		return nil, MismatchedVersionError{Requested: version, Expected: []Version{4, 8}}
	}

	rng := o.RandomSource
	return &genRandom{rng: rng, ver: version}, nil
}

func (g *genRandom) NewUUID() (UUID, error) {
	var uuid UUID

	if err := readRandom(g.rng, uuid[:]); err != nil {
		return NilUUID, IOError{Err: err}
	}

	uuid[6] = (uuid[6] & 0x0f) | byte(g.ver<<4)
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return uuid, nil
}

func (g *genRandom) NewHashUUID(data []byte) (UUID, error) {
	return NilUUID, MustNotHashError{Version: 4}
}

var _ Generator = (*genRandom)(nil)
