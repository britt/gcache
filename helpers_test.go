package gcache

import (
	"fmt"
	"testing"
)

func loader(key interface{}) (interface{}, error) {
	return fmt.Sprintf("valueFor%s", key), nil
}

func testSetCache(t *testing.T, gc Cache, numbers int) {
	for i := 0; i < numbers; i++ {
		key := fmt.Sprintf("Key-%d", i)
		value, err := loader(key)
		if err != nil {
			t.Error(err)
			return
		}
		gc.Set(key, value)
	}
}

func testGetCache(t *testing.T, gc Cache, numbers int) {
	for i := 0; i < numbers; i++ {
		key := fmt.Sprintf("Key-%d", i)
		v, err := gc.Get(key)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expectedV, _ := loader(key)
		if v != expectedV {
			t.Errorf("Expected value is %v, not %v", expectedV, v)
		}
	}
}
