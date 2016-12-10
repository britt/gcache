package gcache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func buildScoreCache(size int) Cache {
	return New(size).
		SCORE().
		ScoringFunc(func(_ interface{}) int { return 1 }).
		WeightingFunc(func(_ interface{}) int { return 1 }).
		Build()
}

func TestScoreCache_GetSet(t *testing.T) {
	size := 1000
	c := buildScoreCache(size)
	testSetCache(t, c, size)
	testGetCache(t, c, size)
}

func TestScoreCache_Get_WithLoader(t *testing.T) {}
func TestScoreCache_GetIFPresent(t *testing.T)   {}
func TestScoreCache_GetALL(t *testing.T)         {}
func TestScoreCache_Eviction(t *testing.T)       {}
func TestScoreCache_Len(t *testing.T)            {}
func TestScoreCache_Remove(t *testing.T)         {}
func TestScoreCache_Purge(t *testing.T)          {}
func TestScoreCache_Keys(t *testing.T)           {}

func TestSplode(t *testing.T) {
	c := buildScoreCache(100)
	assert.NotNil(t, c)
	c.Set(1, 2)
	v, err := c.Get(1)
	assert.Nil(t, err)
	assert.Equal(t, 2, v)
}
