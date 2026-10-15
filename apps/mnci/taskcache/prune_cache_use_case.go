package taskcache

import (
	"encoding/json"
	"sort"
	"time"
)

// PruneResult reports what pruning removed.
type PruneResult struct {
	Entries, Blobs int
	Freed          int64
}

// Prune keeps the cache under maxBytes: it removes the least recently used
// action-cache entries until the blobs the rest reference fit, then sweeps
// every blob no remaining entry references.
func Prune(s Storage, maxBytes int64) (PruneResult, error) {
	type entry struct {
		key      string
		lastUsed time.Time
		blobs    []string
	}
	var entries []entry
	err := s.Walk(AC, func(key string, _ int64, last time.Time) error {
		e := entry{key: key, lastUsed: last}
		if r, err := s.Get(AC, key); err == nil {
			var parsed Entry
			if json.NewDecoder(r).Decode(&parsed) == nil {
				e.blobs = append(e.blobs, parsed.Log)
				for _, o := range parsed.Outputs {
					e.blobs = append(e.blobs, o.Digest)
				}
			}
			_ = r.Close()
		}
		entries = append(entries, e)
		return nil
	})
	if err != nil {
		return PruneResult{}, err
	}
	sizes := map[string]int64{}
	if err := s.Walk(CAS, func(key string, size int64, _ time.Time) error { sizes[key] = size; return nil }); err != nil {
		return PruneResult{}, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].lastUsed.After(entries[j].lastUsed) })

	var result PruneResult
	referenced := map[string]bool{}
	var total int64
	for _, e := range entries {
		var add int64
		for _, b := range e.blobs {
			if !referenced[b] {
				add += sizes[b]
			}
		}
		if total+add > maxBytes {
			if err := s.Delete(AC, e.key); err != nil {
				return result, err
			}
			result.Entries++
			continue
		}
		for _, b := range e.blobs {
			referenced[b] = true
		}
		total += add
	}
	for key, size := range sizes {
		if !referenced[key] {
			if err := s.Delete(CAS, key); err != nil {
				return result, err
			}
			result.Blobs++
			result.Freed += size
		}
	}
	return result, nil
}
