package ordmap

import "sync"

// An Entry is a generic key/value pair within an OrdMap.
type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

// An OrdMap is a generic, concurrency safe ordered map implementation. It works by storing entries in a slice to
// preserve ordering while tracking key lookups to indices in order to fulfill typical O(1) map semantics. Storage
// requirements should be roughly equivalent to map[K]V + map[K]int. Deletes are potentially slow because the
// underlying slice has to be spliced.
type OrdMap[K comparable, V any] struct {
	m sync.RWMutex

	lookup map[K]int
	data   []Entry[K, V]
}

// New returns a new OrdMap with allocations for data and lookup.
func New[K comparable, V any](initialSize int) OrdMap[K, V] {
	return OrdMap[K, V]{
		lookup: make(map[K]int),
		data:   make([]Entry[K, V], initialSize),
	}
}

// Entries returns the ordered slice of Entry structs which can be iterated on.
func (om *OrdMap[K, V]) Entries() []Entry[K, V] {
	om.m.RLock()
	defer om.m.RUnlock()
	return om.data
}

// Get implements a map lookup. This should semantically be O(1) and equivalent to val, ok := map[key].
func (om *OrdMap[K, V]) Get(key K) (V, bool) {
	om.m.RLock()
	defer om.m.RUnlock()
	idx, ok := om.lookup[key]
	if !ok {
		var zero V
		return zero, false
	}

	return om.data[idx].Value, true
}

// Index returns the ordered index associated with the given key.
func (om *OrdMap[K, V]) Index(key K) (int, bool) {
	om.m.RLock()
	defer om.m.RUnlock()
	idx, ok := om.lookup[key]
	return idx, ok
}

// Set a key/value pair within the OrdMap. When the underlying data slice has capacity, this should be a O(1)
// operation. Extra cost is incurred when the slice has to be grown.
func (om *OrdMap[K, V]) Set(key K, val V) {
	om.BulkSet(Entry[K, V]{Key: key, Value: val})
}

// BulkSet allows for setting many entries at once. BulkSet should be preferred over Set when setting many keys at
// once since it only locks the mutex once per operation instead of once per entry. In the case of duplicated keys,
// earlier values in the list will be overwritten.
func (om *OrdMap[K, V]) BulkSet(entries ...Entry[K, V]) {
	om.m.Lock()
	defer om.m.Unlock()
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
	om.m.RLock()
	_, ok := om.lookup[key]
	om.m.RUnlock()
	return ok
}

// Delete a key from an OrdMap. This is not terribly performant, so be careful using this method in hot paths.
func (om *OrdMap[K, V]) Delete(key K) {
	om.m.Lock()
	defer om.m.Unlock()

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
	om.m.RLock()
	defer om.m.RUnlock()
	return len(om.data)
}
