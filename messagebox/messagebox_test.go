package messagebox

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/schollz/messagebox/keypair"
	"github.com/stretchr/testify/assert"
)

func BenchmarkOpen(b *testing.B) {
	bob, _ := keypair.New()
	bill, _ := keypair.New()
	jane, _ := keypair.New()
	jeff, _ := keypair.New()
	world, _ := keypair.New()
	// bob sends to jane and jeff a message
	m, _ := New(world, bob, []string{jeff.Public, jane.Public}, []byte("hello, world"))
	for n := 0; n < b.N; n++ {
		m.Open(world, []keypair.KeyPair{jeff, jane, bob, bill})
	}
}

func TestTweetNacl(t *testing.T) {
	// from https://tweetnacl.js.org/#/secretbox
	key := "RXIBnxEifbR1xHjrTjS05Dr9q4u4gNwwLH7jYqZfy5o="
	nonce := "/Cc4CCyMTK2n6wFUx7ZiDNbj3cIqR11I"
	box := "hSFSqfPaz46d5A5K77sSahDnYjvQvBNap0zirA=="

	nonceBytes, _ := base64.StdEncoding.DecodeString(nonce)
	encryptedBytes, _ := base64.StdEncoding.DecodeString(box)
	encryptedBytes = append(nonceBytes, encryptedBytes...)
	keyBytes, _ := base64.StdEncoding.DecodeString(key)
	var keybytes32 [32]byte
	copy(keybytes32[:], keyBytes[:])
	m, err := decrypt(encryptedBytes, keybytes32)
	assert.Equal(t, []byte("hello, world"), m)
	assert.Nil(t, err)
}

func TestMessage(t *testing.T) {
	world, err := keypair.New()
	assert.Nil(t, err)
	bob, err := keypair.New()
	assert.Nil(t, err)
	bill, err := keypair.New()
	assert.Nil(t, err)
	jane, err := keypair.New()
	assert.Nil(t, err)
	jeff, err := keypair.New()
	assert.Nil(t, err)
	everyone, err := keypair.New()
	assert.Nil(t, err)

	everyone, err = keypair.NewFromPublic(everyone.Public)
	assert.Nil(t, err)
	fmt.Println(everyone)

	msg := []byte("hello, world")
	m, err := New(world, bob, []string{bob.Public}, msg)
	// Print out a message
	mJ, _ := json.MarshalIndent(m, "", " ")
	fmt.Println(string(mJ))

	assert.Nil(t, err)
	recipient, opened, err := m.Open(world, []keypair.KeyPair{bob})
	assert.Nil(t, err)
	assert.Equal(t, recipient, bob)
	assert.Equal(t, msg, opened)

	m, err = New(world, bob, []string{jane.Public, bob.Public, everyone.Public}, msg)
	assert.Nil(t, err)
	recipient, opened, err = m.Open(world, []keypair.KeyPair{jane})
	assert.Nil(t, err)
	assert.Equal(t, recipient, jane)
	assert.Equal(t, msg, opened)

	// jeff can't open because its addressed to jane
	recipient, opened, err = m.Open(world, []keypair.KeyPair{jeff})
	assert.NotNil(t, err)

	// jane can't open if she doesn't know bob exists
	recipient, opened, err = m.Open(world, []keypair.KeyPair{bill})
	assert.NotNil(t, err)

	// can't open, wrong world
	recipient, opened, err = m.Open(jane, []keypair.KeyPair{jane})
	assert.NotNil(t, err)

}
