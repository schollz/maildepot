package keypair

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

func TestTweetNaClKeyPair(t *testing.T) {
	// keypair from https://tweetnacl.js.org/#/box
	me, err := New(KeyPair{
		Public:  "4Bu5tqhJ1qSbbnytbpNYZw+I8kOVQ/4y9VjUyGaL9Rg=",
		Private: "CyPIQzF7xdE/rR6Uc/fV2pO0epXNhTpTbvRvOb3osv0=",
	})
	assert.Nil(t, err)
	enc, _ := me.Encrypt([]byte("hello, world"), me.Public)
	fmt.Println("Nonce:", base64.StdEncoding.EncodeToString(GetNonce(enc)))
	fmt.Println("Box:", base64.StdEncoding.EncodeToString(enc[24:]))
}

func TestIdea(t *testing.T) {
	world, _ := New()
	bob, _ := New()
	enc, _ := world.Encrypt([]byte("hello, world"), bob.Public)
	_, err := world.Decrypt(enc, world.Public)
	assert.NotNil(t, err)
	_, err = world.Decrypt(enc, bob.Public)
	assert.Nil(t, err)
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
