package messagebox

import (
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
		m.Open(jane, []string{jeff.Public, jane.Public, bob.Public, bill.Public})
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

	fmt.Println(jeff, bill, jane, m)

	sender, opened, err := m.Open(bob, []string{bob.Public})
	assert.Nil(t, err)
	assert.Equal(t, sender, bob.Public)
	assert.Equal(t, msg, opened)

	m, err = New(bob, []string{jane.Public}, msg)
	assert.Nil(t, err)
	sender, opened, err = m.Open(jane, []string{jeff.Public, jane.Public, bob.Public, bill.Public})
	assert.Nil(t, err)
	assert.Equal(t, sender, bob.Public)
	assert.Equal(t, msg, opened)

	// jeff can't open because its addressed to jane
	sender, opened, err = m.Open(jeff, []string{jeff.Public, jane.Public, bob.Public, bill.Public})
	assert.NotNil(t, err)

	// jane cna't open if she doesn't know bob exists
	sender, opened, err = m.Open(jane, []string{bill.Public})
	assert.NotNil(t, err)
}
