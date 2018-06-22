package keypair

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

func TestBasic(t *testing.T) {
	bob, err := New()
	fmt.Println(bob.Public)
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
