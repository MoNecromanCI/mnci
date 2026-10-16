package taskcache

import (
	"errors"
	"io"
	"time"
)

// Kind separates action-cache entries from content-addressed blobs.
type Kind string

// Storage kinds.
const (
	AC  Kind = "ac"
	CAS Kind = "cas"
	// Last holds, per task, the key of its most recent cached run (keyed by
	// the SHA-256 of the task id), so a miss can be compared with it.
	Last Kind = "last"
)

// ErrMiss is returned by Get when an entry does not exist.
var ErrMiss = errors.New("cache miss")

// Storage is a cache backend. Keys are lowercase hex SHA-256 digests. A CAS
// Put must verify the content matches its digest; putting an existing digest
// is a success.
type Storage interface {
	Has(kind Kind, key string) (bool, error)
	Get(kind Kind, key string) (io.ReadCloser, error)
	Put(kind Kind, key string, content io.Reader) error
	Delete(kind Kind, key string) error
	// Walk visits every entry of a kind with its size and last use. Backends
	// that cannot enumerate return ErrNotSupported.
	Walk(kind Kind, visit func(key string, size int64, lastUsed time.Time) error) error
	// Touch records a use, so pruning keeps recently used entries.
	Touch(kind Kind, key string) error
}

// ErrNotSupported is returned by optional Storage operations.
var ErrNotSupported = errors.New("not supported by this storage backend")
