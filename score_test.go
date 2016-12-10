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

func TestSplode(t *testing.T) {
	c := buildScoreCache(100)
	assert.NotNil(t, c)
	c.Set(1, 2)
	v, err := c.Get(1)
	assert.Nil(t, err)
	assert.Equal(t, 2, v)
}
