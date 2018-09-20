package mail

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/schollz/maildepot/keypair"
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

func TestBasic(t *testing.T) {
	world, err := keypair.New()
	assert.Nil(t, err)
	fmt.Printf("world: %+v\n", world)
	msg, err := New(world, world, []string{world.Public}, []byte("hello, world"))
	assert.Nil(t, err)
	fmt.Printf("msg: %+v\n", msg)
	openMsg, err := msg.Open(world, []keypair.KeyPair{world})
	assert.Nil(t, err)
	fmt.Printf("open msg: %+v\n", openMsg)
}

func TestMulti(t *testing.T) {
	world, err := keypair.New()
	assert.Nil(t, err)
	bob, _ := keypair.New()

	fmt.Printf("world: %+v\n", world)
	msg, err := New(world, world, []string{bob.Public}, []byte("hello, world"))
	assert.Nil(t, err)
	fmt.Printf("msg: %+v\n", msg)
	openMsg, err := msg.Open(world, []keypair.KeyPair{world})
	assert.Nil(t, err)
	fmt.Printf("open msg: %+v\n", openMsg)
}
