package depot

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDB(t *testing.T) {
	os.Remove("1.db")
	os.Remove("2.db")

	db, err := New("1.db")
	assert.Nil(t, err)
	defer db.Close()

	assert.Nil(t, db.NewBucket("test1"))
	assert.Nil(t, db.Set("test1", "hello0", "world"))
	assert.Nil(t, db.Set("test1", "hello1", "world"))
	assert.Nil(t, db.Set("test1", "hello3", "world"))
	assert.Nil(t, db.Set("test1", "hello4", "world"))
	assert.Nil(t, db.Set("test1", "hello5", "world"))
	assert.Nil(t, db.Set("test1", "hello6", "world"))
	assert.Nil(t, db.Set("test1", "hello7", "world"))
	assert.Nil(t, db.Set("test1", "hello8", "world"))
	assert.Nil(t, db.Set("test1", "hello9", "world"))
	var world string
	err = db.Get("test1", "hello0", &world)
	assert.Nil(t, err)
	assert.Equal(t, "world", world)

	rangeHash, middleKey, count, err := db.getRangeOfHashes("test1", "hello3", "hello5")
	assert.Equal(t, "hello4", middleKey)
	assert.Equal(t, 2, count)
	assert.Nil(t, err)

	db2, err := New("2.db")
	assert.Nil(t, db2.NewBucket("test1"))
	assert.Nil(t, err)
	defer db2.Close()

	// assert.Nil(t, db2.Set("test1", "hello0", "world"))
	// assert.Nil(t, db2.Set("test1", "hello1", "world"))
	// assert.Nil(t, db2.Set("test1", "hello2", "world"))
	// assert.Nil(t, db2.Set("test1", "hello3", "world"))
	// assert.Nil(t, db2.Set("test1", "hello4", "world"))
	// assert.Nil(t, db2.Set("test1", "hello5", "world"))
	// assert.Nil(t, db2.Set("test1", "hello6", "world"))

	isEqual, err := db2.checkRangeOfHashes("test1", "hello3", "hello5", rangeHash)
	assert.True(t, isEqual)
	assert.Nil(t, err)

	fmt.Println(findExchange("test1", "first", "last", db, db2, []toExchange{}))
}

type toExchange struct {
	first string
	last  string
}

func findExchange(bucket, first, last string, db *DB, db2 *DB, exchange1 []toExchange) (finalExchange []toExchange, err error) {
	exchange2, err := find(bucket, first, last, db, db2, exchange1)
	if err != nil {
		return
	}
	finalExchange = make([]toExchange, len(exchange2))
	num := 0
	for i, ex := range exchange2 {
		if i == 0 {
			finalExchange[num] = ex
			num++
			continue
		}
		if finalExchange[num-1].last == ex.first {
			finalExchange[num-1].last = ex.last
		} else {
			finalExchange[num] = ex
			num++
		}
	}
	finalExchange = finalExchange[:num]
	return
}

func find(bucket, first, last string, db *DB, db2 *DB, exchange1 []toExchange) (exchange2 []toExchange, err error) {
	exchange2 = exchange1
	rangeHash, middleKey, count, err := db.getRangeOfHashes(bucket, first, last)
	if err != nil {
		return
	}
	isEqual, err := db2.checkRangeOfHashes(bucket, first, last, rangeHash)
	if err != nil {
		return
	}
	if isEqual {
		return
	}

	// not equal, search some more
	if count == 1 {
		log.Println("exchange", first, last)
		exchange2 = append(exchange2, toExchange{first, last})
		return
	}

	exchange2, err = find(bucket, first, middleKey, db, db2, exchange2)
	if err != nil {
		return
	}
	exchange2, err = find(bucket, middleKey, last, db, db2, exchange2)
	if err != nil {
		return
	}

	return
}
