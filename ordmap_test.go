package ordmap_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/eriktate/go-ordmap"
)

func Test_SimpleLifecycle(t *testing.T) {
	om := ordmap.New[string, int](0)

	if _, ok := om.Get("life"); ok {
		t.Fatal("expected nothing back from empty ordmap")
	}

	om.Set("life", 42)

	life, ok := om.Get("life")
	if !ok {
		t.Fatal("expected value back for 'life' key")
	}

	if life != 42 {
		t.Fatalf("expected the meaning of life, but received %d", life)
	}

	if !om.Has("life") {
		t.Fatal("expected ordmap to contain 'life'")
	}

	om.Set("life", 7)
	life, _ = om.Get("life")
	if life != 7 {
		t.Fatalf("expected to change the meaning of life to 7, but got %d", life)
	}

	if om.Len() != 1 {
		t.Fatalf("expected map length to be 1, got %d", om.Len())
	}

	om.Delete("life")

	if _, ok := om.Get("life"); ok {
		t.Fatalf("expected no value for life after deletion")
	}

	if om.Has("life") {
		t.Fatal("expected ordmap to no longer contain life")
	}
}

func Test_Order(t *testing.T) {
	om := ordmap.New[string, int](0)

	entries := make([]ordmap.Entry[string, int], 10_000)
	for idx := range entries {
		entries[idx] = ordmap.Entry[string, int]{
			Key:   fmt.Sprintf("key %d", idx),
			Value: idx,
		}
	}

	om.BulkSet(entries...)
	for idx, entry := range om.Entries() {
		if entry.Key != fmt.Sprintf("key %d", idx) || entry.Value != idx {
			t.Fatalf("expected entry #%d to be %d, received key=%s val=%d", idx, idx, entry.Key, entry.Value)
		}
	}
}

func Test_ConcurrentAccess(t *testing.T) {
	om := ordmap.New[string, int](0)

	wg := sync.WaitGroup{}
	for i := range 10 {
		wg.Add(1)
		go func(idx int) {
			for j := range 1000 {
				om.Set(fmt.Sprintf("%d", j), idx*j)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	if om.Len() != 1000 {
		t.Fatalf("expected final map length to be 1000, found %d", om.Len())
	}
}

func Test_EntriesAndIterators(t *testing.T) {
	testFn := func(om *ordmap.OrdMap[string, int]) {
		for i := 100; i > 0; i-- {
			om.Set(fmt.Sprintf("%d", i), i)
		}

		for idx, entry := range om.Entries() {
			if entry.Value != 100-idx {
				t.Fatalf("expected value of %d, found %d", 100-idx, entry.Value)
			}
		}

		idx := 0
		for key, val := range om.EntryIter() {
			expectedVal := 100 - idx
			expectedKey := fmt.Sprintf("%d", expectedVal)
			if key != expectedKey {
				t.Fatalf("expected key of %q, found %q", expectedKey, key)
			}

			if val != expectedVal {
				t.Fatalf("expected value of %d, found %d", expectedVal, val)
			}

			idx++
		}

		idx = 0
		for val := range om.Values() {
			if val != 100-idx {
				t.Fatalf("expected value of %d, found %d", 100-idx, val)
			}
			idx++
		}

		for idx, val := range om.All() {
			if val != 100-idx {
				t.Fatalf("expected value of %d, found %d", 100-idx, val)
			}
		}

		idx = 0
		for key := range om.Keys() {
			expected := fmt.Sprintf("%d", 100-idx)
			if key != expected {
				t.Fatalf("expected key of %q, found %q", expected, key)
			}
			idx++
		}
	}

	testFn(ordmap.New[string, int](0))
	testFn(ordmap.NewUnsafe[string, int](0))
}

func BenchmarkSafeOrdmap(b *testing.B) {
	om := ordmap.New[int, int](0)
	for b.Loop() {
		for i := range 100 {
			expected := i * 2
			om.Set(i, expected)
			val, ok := om.Get(i)
			if !ok {
				b.Fatalf("expected value for %d", i)
			}

			if val != expected {
				b.Fatalf("expected value of %d, found %d", i, expected)
			}
		}
	}
}

func BenchmarkUnsafeOrdmap(b *testing.B) {
	om := ordmap.NewUnsafe[int, int](0)
	for b.Loop() {
		for i := range 100 {
			expected := i * 2
			om.Set(i, expected)
			val, ok := om.Get(i)
			if !ok {
				b.Fatalf("expected value for %d", i)
			}

			if val != expected {
				b.Fatalf("expected value of %d, found %d", i, expected)
			}
		}
	}
}
