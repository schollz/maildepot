package messagebox

import (
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
	// bob sends to jane and jeff a message
	m, _ := New(bob, []string{jeff.Public, jane.Public}, []byte("hello, world"))
	for n := 0; n < b.N; n++ {
		m.Open([]keypair.KeyPair{jeff, jane, bob, bill})
	}
}

func TestMessage(t *testing.T) {
	bob, err := keypair.New()
	assert.Nil(t, err)
	bill, err := keypair.New()
	assert.Nil(t, err)
	jane, err := keypair.New()
	assert.Nil(t, err)
	jeff, err := keypair.New()
	assert.Nil(t, err)

	msg := []byte("hello, world")
	m, err := New(bob, []string{bob.Public}, msg)
	assert.Nil(t, err)
	recipient, opened, err := m.Open([]keypair.KeyPair{bob})
	assert.Nil(t, err)
	assert.Equal(t, recipient, bob.Public)
	assert.Equal(t, msg, opened)

	m, err = New(bob, []string{jane.Public}, msg)
	assert.Nil(t, err)
	recipient, opened, err = m.Open([]keypair.KeyPair{jane})
	assert.Nil(t, err)
	assert.Equal(t, recipient, jane.Public)
	assert.Equal(t, msg, opened)

	// jeff can't open because its addressed to jane
	recipient, opened, err = m.Open([]keypair.KeyPair{jeff})
	assert.NotNil(t, err)

	// jane cna't open if she doesn't know bob exists
	recipient, opened, err = m.Open([]keypair.KeyPair{bill})
	assert.NotNil(t, err)

	// Print out a message
	mJ, _ := json.MarshalIndent(m, "", " ")
	fmt.Println(string(mJ))
}
