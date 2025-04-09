package ordmap

import (
	"iter"
	"sync"
)

// An Entry is a generic key/value pair within an OrdMap.
type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

// An OrdMap is a generic, concurrency-safe, ordered map implementation. It works by storing entries in a slice to
// preserve ordering while tracking key lookups to indices in order to fulfill typical O(1) map semantics. Storage
// requirements should be roughly equivalent to map[K]V + map[K]int. Deletes are potentially slow because the
// underlying slice has to be spliced.
type OrdMap[K comparable, V any] struct {
	m sync.RWMutex

	// unsafe trades concurrency safety for the performance gain of avoiding locks
	unsafe bool
	lookup map[K]int
	data   []Entry[K, V]
}

// New returns a new concurrency safe OrdMap with allocations for data and lookup.
func New[K comparable, V any](initialSize int) *OrdMap[K, V] {
	return &OrdMap[K, V]{
		lookup: make(map[K]int),
		data:   make([]Entry[K, V], initialSize),
	}
}

// NewUnsafe returns a new concurrency unsafe OrdMap with allocations for data and lookup
func NewUnsafe[K comparable, V any](initialSize int) *OrdMap[K, V] {
	return &OrdMap[K, V]{
		lookup: make(map[K]int),
		data:   make([]Entry[K, V], initialSize),
		unsafe: true,
	}
}

// Entries returns the ordered slice of Entry structs which can be iterated on.
func (om *OrdMap[K, V]) Entries() []Entry[K, V] {
	if !om.unsafe {
		om.m.RLock()
		defer om.m.RUnlock()
	}
	return om.data
}

// Get implements a map lookup. This should semantically be O(1) and equivalent to val, ok := map[key].
func (om *OrdMap[K, V]) Get(key K) (V, bool) {
	if !om.unsafe {
		om.m.RLock()
		defer om.m.RUnlock()
	}

	idx, ok := om.lookup[key]
	if !ok {
		var zero V
		return zero, false
	}

	return om.data[idx].Value, true
}

// Index returns the ordered index associated with the given key.
func (om *OrdMap[K, V]) Index(key K) (int, bool) {
	if !om.unsafe {
		om.m.RLock()
		defer om.m.RUnlock()
	}

	idx, ok := om.lookup[key]
	return idx, ok
}

// Set a key/value pair within the OrdMap. When the underlying data slice has capacity, this should be a O(1)
// operation. Extra cost is incurred when the slice has to be grown.
func (om *OrdMap[K, V]) Set(key K, val V) {
	om.BulkSet(Entry[K, V]{Key: key, Value: val})
}

// BulkSet allows for setting many entries at once. BulkSet should be preferred over multiple calls to [Set]
// since it only locks the mutex once per operation instead of once per entry. In the case of duplicated
// keys, earlier values in the list will be overwritten.
func (om *OrdMap[K, V]) BulkSet(entries ...Entry[K, V]) {
	if !om.unsafe {
		om.m.Lock()
		defer om.m.Unlock()
	}

	for _, entry := range entries {
		idx, ok := om.lookup[entry.Key]
		if ok {
			om.data[idx] = entry
			return
		}

		om.lookup[entry.Key] = len(om.data)
		om.data = append(om.data, entry)
	}
}

// Has works the same as Get but does not return the value. It's included for convenience.
func (om *OrdMap[K, V]) Has(key K) bool {
	if !om.unsafe {
		om.m.RLock()
		defer om.m.RUnlock()
	}

	_, ok := om.lookup[key]
	return ok
}

// Delete a key from an OrdMap. This is not terribly performant, so be careful using this method in hot paths.
func (om *OrdMap[K, V]) Delete(key K) {
	if !om.unsafe {
		om.m.Lock()
		defer om.m.Unlock()
	}

	idx, ok := om.lookup[key]
	if !ok {
		return
	}

	defer delete(om.lookup, key)

	if idx == 0 {
		om.data = om.data[1:]
		return
	}

	if idx == len(om.data) {
		om.data = om.data[0 : len(om.data)-1]
		return
	}

	om.data = append(om.data[0:idx], om.data[idx+1:]...)
}

// Len returns the current length of the OrdMap.
func (om *OrdMap[K, V]) Len() int {
	if !om.unsafe {
		om.m.RLock()
		defer om.m.RUnlock()
	}

	return len(om.data)
}

// Values returns an iterator over the values present in the OrdMap. In order
// to ensure the iterator is safe to use concurrently, each iteration must
// acquire a read lock which can affect performance.
func (om *OrdMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		if om.unsafe {
			for _, entry := range om.data {
				if !yield(entry.Value) {
					break
				}
			}
			return
		}

		var idx int
		for {
			om.m.RLock()
			if idx >= len(om.data) {
				om.m.RUnlock()
				return
			}
			entry := om.data[idx]
			om.m.RUnlock()
			if !yield(entry.Value) {
				return
			}
			idx++
		}
	}
}

// Keys returns an iterator over the keys present in the OrdMap. In order
// to ensure the iterator is safe to use concurrently, each iteration must
// acquire a read lock which can affect performance.
func (om *OrdMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		if om.unsafe {
			for _, entry := range om.data {
				if !yield(entry.Key) {
					break
				}
			}
			return
		}

		var idx int
		for {
			om.m.RLock()
			if idx >= len(om.data) {
				om.m.RUnlock()
				return
			}
			entry := om.data[idx]
			om.m.RUnlock()
			if !yield(entry.Key) {
				return
			}
			idx++
		}
	}
}

// All returns an iterator returning pairs of ordered indices and entry Values.
// In order to ensure the iterator is safe to use concurrently, each iteration
// must acquire a read lock which can affect performance.
func (om *OrdMap[K, V]) All() iter.Seq2[int, V] {
	return func(yield func(int, V) bool) {
		if om.unsafe {
			for idx, entry := range om.data {
				if !yield(idx, entry.Value) {
					break
				}
			}
			return
		}

		var idx int
		for {
			om.m.RLock()
			if idx >= len(om.data) {
				om.m.RUnlock()
				return
			}
			entry := om.data[idx]
			om.m.RUnlock()
			if !yield(idx, entry.Value) {
				return
			}
			idx++
		}
	}
}

// EntryIter is the iterator form of [Entries]. In order to ensure the iterator
// is safe to use concurrently, each iteration must acquire a read lock which
// can affect performance.
func (om *OrdMap[K, V]) EntryIter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if om.unsafe {
			for _, entry := range om.data {
				if !yield(entry.Key, entry.Value) {
					break
				}
			}
			return
		}

		var idx int
		for {
			om.m.RLock()
			if idx >= len(om.data) {
				om.m.RUnlock()
				return
			}
			entry := om.data[idx]
			om.m.RUnlock()
			if !yield(entry.Key, entry.Value) {
				return
			}
			idx++
		}
	}
}
