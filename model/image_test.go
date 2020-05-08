package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImage_AddSize(t *testing.T) {
	i := Image{}
	i.AddSize("test", 1, 2)
	assert.Len(t, i.Sizes, 1)
	assert.Equal(t, "test", i.Sizes[0].Path)
	assert.Equal(t, 1, i.Sizes[0].Width)
	assert.Equal(t, 2, i.Sizes[0].Height)

	i.AddSize("test2", 1, 2)
	assert.Len(t, i.Sizes, 2)
}

func TestImage_HasResizedSize(t *testing.T) {
	i := Image{}
	i.AddSize("test", 1, 2)

	assert.True(t, i.HasResizedSize(1, 2))
	assert.False(t, i.HasResizedSize(1, 1))
}
