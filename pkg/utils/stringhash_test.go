package utils

import (
	"testing"

	tassert "github.com/stretchr/testify/assert"
)

func TestHashFromString(t *testing.T) {
	assert := tassert.New(t)

	stringHash, err := HashFromString("some string")
	assert.NoError(err)
	assert.NotZero(stringHash)

	emptyStringHash, err := HashFromString("")
	assert.NoError(err)
	assert.Equal(emptyStringHash, uint32(0x11c9dc5))

	assert.NotEqual(stringHash, emptyStringHash)
}
