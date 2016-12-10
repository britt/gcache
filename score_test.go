package gcache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func buildScoreCache(size, weight int) Cache {
	return New(size).
		SCORE().
		ScoringFunc(func(_ interface{}) int { return 1 }).
		WeightingFunc(func(_ interface{}) int { return weight }).
		Build()
}

func TestScoreCache_GetSet(t *testing.T) {
	size := 1000
	c := buildScoreCache(size, 1)
	testSetCache(t, c, size)
	testGetCache(t, c, size)
}

func TestScoreCache_Get_WithLoader(t *testing.T) {
	c := New(100).
		SCORE().
		ScoringFunc(func(_ interface{}) int { return 1 }).
		WeightingFunc(func(_ interface{}) int { return 2 }).
		LoaderFunc(func(key interface{}) (interface{}, error) {
			return fmt.Sprintf("%v", key), nil
		}).
		Build()

	v, err := c.Get(1)
	assert.Equal(t, "1", v)
	assert.Nil(t, err)

	v, err = c.Get(2)
	assert.Equal(t, "2", v)
	assert.Nil(t, err)

	v, err = c.Get(3)
	assert.Equal(t, "3", v)
	assert.Nil(t, err)
}

func TestScoreCache_GetIFPresent(t *testing.T) {
	c := buildScoreCache(10, 2)

	for i := 0; i < 5; i++ {
		c.Set(i, i)
	}

	for i := 0; i < 10; i++ {
		v, err := c.GetIFPresent(i)
		if i < 5 {
			assert.Equal(t, i, v)
			assert.Nil(t, err)
		} else {
			assert.Nil(t, v)
			assert.Equal(t, err, KeyNotFoundError)
		}
	}
}

func TestScoreCache_GetALL(t *testing.T) {
	c := buildScoreCache(10, 2)

	for i := 0; i < 5; i++ {
		c.Set(i, i)
	}

	all := c.GetALL()
	assert.Equal(t, 5, len(all))
	for i := 0; i < 5; i++ {
		assert.Equal(t, i, all[i])
	}
}

func TestScoreCache_Set_Callback(t *testing.T) {
	added := make(map[interface{}]interface{})
	c := New(10).
		SCORE().
		ScoringFunc(func(_ interface{}) int { return 1 }).
		WeightingFunc(func(_ interface{}) int { return 1 }).
		AddedFunc(func(key, value interface{}) {
			added[key] = value
		}).
		Build()

	for i := 0; i < 5; i++ {
		c.Set(i, i)
	}

	assert.Equal(t, 5, len(added))
	for i := 0; i < 5; i++ {
		assert.Equal(t, i, added[i])
	}
}

func TestScoreCache_MaxSize(t *testing.T) {
	c := buildScoreCache(20, 2)

	for i := 0; i < 50; i++ {
		c.Set(i, i)
	}

	assert.Equal(t, 10, c.Len())
	assert.Equal(t, 20, c.(*ScoreCache).totalWeight)
}

func TestScoreCache_Eviction(t *testing.T) {
	evictions := 0

	c := New(10).
		SCORE().
		ScoringFunc(func(_ interface{}) int { return 1 }).
		WeightingFunc(func(_ interface{}) int { return 1 }).
		EvictedFunc(func(key, value interface{}) {
			evictions++
		}).
		Build()

	for i := 0; i < 30; i++ {
		c.Set(i, i)
	}

	assert.Equal(t, 10, c.Len())
	assert.Equal(t, 20, evictions)
}

func TestScoreCache_Len(t *testing.T) {
	c := buildScoreCache(100, 2)

	for i := 0; i < 17; i++ {
		c.Set(i, i)
	}

	assert.Equal(t, 17, c.Len())
}

func TestScoreCache_Remove(t *testing.T) {}
func TestScoreCache_Purge(t *testing.T)  {}
func TestScoreCache_Keys(t *testing.T)   {}
func TestScoreCache_Stats(t *testing.T)  {}
