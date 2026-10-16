package remotecache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"time"

	"github.com/MoNecromanCI/mnci/apps/mnci/taskcache"
)

// Tiered is the local cache in front of a remote store. Reads try local
// first and keep what the remote had, once verified; writes go to both
// (unless ReadOnly). The remote never fails a run: a remote error is a
// warning, and an object that fails verification is a miss.
type Tiered struct {
	Local    taskcache.Storage
	Remote   ObjectStore
	Secret   []byte
	ReadOnly bool
	// Warn reports remote problems; each kind of problem once.
	Warn   func(string)
	warned map[string]bool
}

var _ taskcache.Storage = (*Tiered)(nil)

func name(kind taskcache.Kind, key string) string { return string(kind) + "/" + key }

func (t *Tiered) warn(what, detail string) {
	if t.warned == nil {
		t.warned = map[string]bool{}
	}
	if t.Warn != nil && !t.warned[what] {
		t.warned[what] = true
		t.Warn(what + ": " + detail)
	}
}

// pull fetches an entry from the remote into the local cache, verified;
// false when the remote does not have it, could not be asked, or sent
// something that fails verification.
func (t *Tiered) pull(kind taskcache.Kind, key string) bool {
	if kind == taskcache.Last {
		return false
	}
	object, found, err := t.Remote.Get(name(kind, key))
	if err != nil {
		t.warn("the remote cache could not be read", err.Error())
		return false
	}
	if !found {
		return false
	}
	content, err := Verify(t.Secret, string(kind), key, object)
	if err == nil && kind == taskcache.CAS {
		if sum := sha256.Sum256(content); hex.EncodeToString(sum[:]) != key {
			err = ErrTampered
		}
	}
	if err != nil {
		t.warn("a remote cache object was refused", name(kind, key)+": "+err.Error())
		return false
	}
	return t.Local.Put(kind, key, bytes.NewReader(content)) == nil
}

// Has implements taskcache.Storage; a blob only the remote has is fetched,
// so what Has promises, Get delivers.
func (t *Tiered) Has(kind taskcache.Kind, key string) (bool, error) {
	if ok, err := t.Local.Has(kind, key); ok || err != nil {
		return ok, err
	}
	return t.pull(kind, key), nil
}

// Get implements taskcache.Storage.
func (t *Tiered) Get(kind taskcache.Kind, key string) (io.ReadCloser, error) {
	r, err := t.Local.Get(kind, key)
	if err != taskcache.ErrMiss {
		return r, err
	}
	if !t.pull(kind, key) {
		return nil, taskcache.ErrMiss
	}
	return t.Local.Get(kind, key)
}

// Put implements taskcache.Storage.
func (t *Tiered) Put(kind taskcache.Kind, key string, content io.Reader) error {
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}
	if err := t.Local.Put(kind, key, bytes.NewReader(data)); err != nil {
		return err
	}
	if t.ReadOnly || kind == taskcache.Last {
		return nil
	}
	if kind == taskcache.CAS {
		if ok, err := t.Remote.Has(name(kind, key)); err == nil && ok {
			return nil
		}
	}
	if err := t.Remote.Put(name(kind, key), Sign(t.Secret, string(kind), key, data)); err != nil {
		t.warn("the remote cache could not be written", err.Error())
	}
	return nil
}

// Delete implements taskcache.Storage (the local copy only).
func (t *Tiered) Delete(kind taskcache.Kind, key string) error { return t.Local.Delete(kind, key) }

// Walk implements taskcache.Storage (the local copies only).
func (t *Tiered) Walk(kind taskcache.Kind, visit func(string, int64, time.Time) error) error {
	return t.Local.Walk(kind, visit)
}

// Touch implements taskcache.Storage (the local copy only).
func (t *Tiered) Touch(kind taskcache.Kind, key string) error { return t.Local.Touch(kind, key) }
