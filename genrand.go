package youyouayedee

import (
	"io"
)

// NewRandomGenerator constructs a new Generator that produces new randomly
// generated UUIDs.
//
// Versions 4 and 8 are supported.
//
func NewRandomGenerator(version Version, o Options) (Generator, error) {
	switch version {
	case 4:
		// pass

	case 8:
		// pass

	default:
		return nil, ErrVersionMismatch{Requested: version, Expected: []Version{4, 8}}
	}

	rng := o.RandomSource
	return &genRandom{rng: rng, ver: version}, nil
}

type genRandom struct {
	GeneratorBase

	rng io.Reader
	ver Version
}

func (g *genRandom) NewUUID() (UUID, error) {
	var uuid UUID

	if err := readRandom(g.rng, uuid[:]); err != nil {
		return Nil, err
	}

	uuid[6] = (uuid[6] & 0x0f) | byte(g.ver<<4)
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return uuid, nil
}

var _ Generator = (*genRandom)(nil)
