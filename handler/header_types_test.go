package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("Content-Type", Normalize("content-type"))
	assert.Equal("Content-Type", Normalize("CONTENT-TYPE"))
	assert.Equal("Content-Type", Normalize("cONtENT-tYpE"))
}
