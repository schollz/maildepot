package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDB(t *testing.T) {
	db, err := New("1.db")
	assert.Nil(t, err)
	defer db.Close()

	assert.Nil(t, db.NewBucket("test1"))
	assert.Nil(t, db.Set("test1", "hello", "world"))
	assert.Nil(t, db.Set("test1", "hello1", "world"))
	assert.Nil(t, db.Set("test1", "hello2", "world"))
	assert.Nil(t, db.Set("test1", "hello3", "world"))
	assert.Nil(t, db.Set("test1", "hello4", "world"))
	assert.Nil(t, db.Set("test1", "hello5", "world"))
	var world string
	err = db.Get("test1", "hello", &world)
	assert.Nil(t, err)
	assert.Equal(t, "world", world)

	rangeHash, middleKey, err := db.getRangeOfHashes("test1", "hello1", "hello1")
	assert.Equal(t, "hello2", middleKey)
	assert.Nil(t, err)

	db2, err := New("2.db")
	assert.Nil(t, err)
	defer db2.Close()
	fmt.Println(db2.checkRangeOfHashes("test1", "first", "last", rangeHash))
}
