package db

import (
	"bytes"
	"fmt"

	"github.com/OneOfOne/xxhash"
	bolt "go.etcd.io/bbolt"
)

func (db *DB) getRangeOfHashes(bucket, first, last string) (rangeHash string, middleKey string, err error) {
	db.RLock()
	defer db.RUnlock()

	h64 := xxhash.New64()
	count := 0
	err = db.db.View(func(tx *bolt.Tx) error {
		// Assume our events bucket exists and has RFC3339 encoded time keys.
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket '%s' does not exist", bucket)
		}
		c := b.Cursor()

		// Our time range spans the 90's decade.
		min := []byte(first)
		max := []byte(last)

		if first == "first" && last == "last" {
			for k, v := c.First(); k != nil; k, v = c.Next() {
				fmt.Printf("%s: %s\n", k, v)
				h64.WriteString(string(k))
				count++
			}
		} else if first == "first" {
			for k, v := c.First(); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
				fmt.Printf("%s: %s\n", k, v)
				h64.WriteString(string(k))
				count++
			}
		} else if last == "last" {
			for k, v := c.Seek(min); k != nil; k, v = c.Next() {
				fmt.Printf("%s: %s\n", k, v)
				h64.WriteString(string(k))
				count++
			}
		} else {
			for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
				fmt.Printf("%s: %s\n", k, v)
				h64.WriteString(string(k))
				count++
			}
		}

		return nil
	})
	if err != nil {
		return
	}
	rangeHash = fmt.Sprintf("%x", h64.Sum64())

	err = db.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucket)).Cursor()
		cur := 0
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			cur++
			if cur == count/2 {
				middleKey = string(k)
				return nil
			}
		}
		return nil
	})
	return
}

func (db *DB) checkRangeOfHashes(bucket, first, last, rangeHashFromOtherDB string) (isEqual bool, err error) {
	rangeHash, _, err := db.getRangeOfHashes(bucket, first, last)
	if err != nil {
		return
	}
	isEqual = rangeHash == rangeHashFromOtherDB
	return
}
