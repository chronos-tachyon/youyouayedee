package youyouayedee

import (
	"crypto/md5"
	"crypto/sha1"
	"hash"
	"sync/atomic"
)

// NewHashGenerator initializes a new Generator that produces hash-based UUIDs.
//
// Versions 3, 5, and 8 are supported.  Options must specify a valid, non-nil
// UUID in the Namespace field.  For version 8 only, Options must also specify
// a valid, non-nil HashFactory callback.
//
func NewHashGenerator(version Version, o Options) (Generator, error) {
	var factory func() hash.Hash
	switch version {
	case 3:
		factory = md5.New

	case 5:
		factory = sha1.New

	case 8:
		factory = o.HashFactory
		if factory == nil {
			return nil, ErrHashFactoryIsNil{Version: version}
		}

	default:
		return nil, ErrVersionMismatch{Requested: version, Expected: []Version{3, 5, 8}}
	}

	ns := o.Namespace
	if !ns.IsValid() {
		return nil, ErrNamespaceNotValid{Version: version, Namespace: ns}
	}

	h := factory()
	return &genHash{ns: ns, fn: factory, ver: version, busy: 0, hasher: h}, nil
}

type genHash struct {
	GeneratorBase

	ns     UUID
	fn     func() hash.Hash
	ver    Version
	busy   uintptr
	hasher hash.Hash
}

func (g *genHash) NewHashUUID(data []byte) (UUID, error) {
	var uuid UUID
	if atomic.CompareAndSwapUintptr(&g.busy, 0, 1) {
		uuid = hashImpl(g.hasher, g.ver, g.ns, data)
		atomic.StoreUintptr(&g.busy, 0)
	} else {
		h := g.fn()
		uuid = hashImpl(h, g.ver, g.ns, data)
	}
	return uuid, nil
}

func hashImpl(h hash.Hash, v Version, ns UUID, data []byte) UUID {
	h.Reset()
	_, _ = h.Write(ns[:])
	_, _ = h.Write(data)
	s := h.Sum(nil)
	var uuid UUID
	copy(uuid[:], s)
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	uuid[6] = (uuid[6] & 0x0f) | byte(v<<4)
	return uuid
}

var _ Generator = (*genHash)(nil)
