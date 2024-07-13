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
	life, ok = om.Get("life")
	if life != 7 {
		t.Fatalf("expected to change the meaning of life to 7, but got %d", life)
	}

	if om.Len() != 1 {
		t.Fatalf("expected map length to be 1, got %d", om.Len())
	}

	om.Delete("life")

	life, ok = om.Get("life")
	if ok {
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
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			for j := 0; j < 1000; j++ {
				om.Set(fmt.Sprintf("%d", j), idx*j)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	if om.Len() != 1000 {
		t.Fatalf("expected final map length to be 1000, got %d", om.Len())
	}
}
