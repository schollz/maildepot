package depot

import (
	"bytes"
	"fmt"
	"hash/adler32"

	bolt "go.etcd.io/bbolt"
)

func (db *DB) getRangeOfHashes(bucket, first, last string) (rangeHash string, middleKey string, count int, err error) {
	db.RLock()
	defer db.RUnlock()

	a32 := adler32.New()
	count = 0
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
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				a32.Write(k)
				count++
			}
		} else if first == "first" {
			for k, _ := c.First(); k != nil && bytes.Compare(k, max) < 0; k, _ = c.Next() {
				a32.Write(k)
				count++
			}
		} else if last == "last" {
			for k, _ := c.Seek(min); k != nil; k, _ = c.Next() {
				a32.Write(k)
				count++
			}
		} else {
			for k, _ := c.Seek(min); k != nil && bytes.Compare(k, max) < 0; k, _ = c.Next() {
				a32.Write(k)
				count++
			}
		}

		return nil
	})
	if err != nil {
		return
	}
	rangeHash = fmt.Sprintf("%x", a32.Sum32())

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

		cur := 0

		if first == "first" && last == "last" {
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				if cur == count/2 {
					middleKey = string(k)
					return nil
				}
				cur++
			}
		} else if first == "first" {
			for k, _ := c.First(); k != nil && bytes.Compare(k, max) < 0; k, _ = c.Next() {
				if cur == count/2 {
					middleKey = string(k)
					return nil
				}
				cur++
			}
		} else if last == "last" {
			for k, _ := c.Seek(min); k != nil; k, _ = c.Next() {
				if cur == count/2 {
					middleKey = string(k)
					return nil
				}
				cur++
			}
		} else {
			for k, _ := c.Seek(min); k != nil && bytes.Compare(k, max) < 0; k, _ = c.Next() {
				if cur == count/2 {
					middleKey = string(k)
					return nil
				}
				cur++
			}
		}

		return nil
	})
	return
}

func (db *DB) checkRangeOfHashes(bucket, first, last, rangeHashFromOtherDB string) (isEqual bool, err error) {
	rangeHash, _, _, err := db.getRangeOfHashes(bucket, first, last)
	if err != nil {
		return
	}
	isEqual = rangeHash == rangeHashFromOtherDB
	return
}
