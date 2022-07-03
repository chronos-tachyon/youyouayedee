package uuid

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"hash"
	"sync/atomic"
)

type genHash struct {
	ns     UUID
	fn     func() hash.Hash
	ver    Version
	busy   uintptr
	hasher hash.Hash
}

// NewHashGenerator initializes a new Generator that produces hash-based UUIDs.
//
// Versions 3, 5, and 8 are supported.  GeneratorOptions must specify a valid,
// non-nil UUID in the Namespace field.  For versions 3 or 5, the factory
// argument is ignored; md5.New or sha1.New is always used for those respective
// UUID versions.  For version 8, the factory argument must be non-nil.
//
func NewHashGenerator(version Version, factory func() hash.Hash, o GeneratorOptions) (Generator, error) {
	switch version {
	case 3:
		factory = md5.New

	case 5:
		factory = sha1.New

	case 8:
		if factory == nil {
			return nil, fmt.Errorf("factory is nil; must specify a hash.Hash provider")
		}

	default:
		return nil, MismatchedVersionError{Requested: version, Expected: []Version{3, 5, 8}}
	}

	ns := o.Namespace
	if !ns.IsValid() {
		return nil, fmt.Errorf("Namespace is not a valid UUID")
	}

	h := factory()
	return &genHash{ns: ns, fn: factory, ver: version, busy: 0, hasher: h}, nil
}

func (g *genHash) NewUUID() (UUID, error) {
	return NilUUID, MustHashError{Version: g.ver}
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