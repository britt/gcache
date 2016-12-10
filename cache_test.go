package gcache

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoaderFunc(t *testing.T) {
	size := 2
	var testCaches = []*CacheBuilder{
		New(size).Simple(),
		New(size).LRU(),
		New(size).LFU(),
		New(size).ARC(),
	}
	for _, builder := range testCaches {
		var testCounter int64
		counter := 1000
		cache := builder.
			LoaderFunc(func(key interface{}) (interface{}, error) {
				time.Sleep(10 * time.Millisecond)
				return atomic.AddInt64(&testCounter, 1), nil
			}).
			EvictedFunc(func(key, value interface{}) {
				panic(key)
			}).Build()

		var wg sync.WaitGroup
		for i := 0; i < counter; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := cache.Get(0)
				if err != nil {
					t.Error(err)
				}
			}()
		}
		wg.Wait()

		if testCounter != 1 {
			t.Errorf("testCounter != %v", testCounter)
		}
	}
}
