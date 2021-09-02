package queryparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatches(t *testing.T) {
	assert.True(t, Matches("a bc d ef", "bc"))
	assert.False(t, Matches("a bc d ef", "b"))
	assert.False(t, Matches("a bc d ef", "c"))
	assert.False(t, Matches("a bc d ef", "x"))
	assert.False(t, Matches("a bc d ef ", "x"))
	assert.False(t, Matches("a bc d ef ", "x "))
}
