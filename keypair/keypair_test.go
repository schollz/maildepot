package keypair

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	bob, err := New()
	assert.Nil(t, err)
	jane, err := New()
	assert.Nil(t, err)
	jeff, err := New()
	assert.Nil(t, err)

	msg := []byte("hello, world")
	enc, err := bob.Encrypt(msg, jane.Public)
	assert.Nil(t, err)
	dec, err := jane.Decrypt(enc, bob.Public)
	assert.Nil(t, err)
	assert.Equal(t, msg, dec)

	dec, err = jeff.Decrypt(enc, bob.Public)
	assert.NotNil(t, err)
	assert.NotEqual(t, msg, dec)
}
