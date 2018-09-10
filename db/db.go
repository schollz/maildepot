package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/schollz/messagebox/keypair"
	bolt "go.etcd.io/bbolt"
)

// NoSuchKeyError is thrown when calling Get with invalid key
type NoSuchKeyError struct {
	key string
}

func (err NoSuchKeyError) Error() string {
	return "BoltStore: no such key \"" + err.key + "\""
}

// DB is the main structure for the distributed DB
type DB struct {
	worldKey keypair.KeyPair
	db       *bolt.DB
	sync.RWMutex
}

// New generates a new DDB
func New(dbname string) (db *DB, err error) {
	db = new(DB)
	db.worldKey, err = keypair.NewDeterministic("world1")
	if err != nil {
		return
	}

	db.db, err = bolt.Open(dbname, 0600, nil)
	return
}

// Close closes the database
func (db *DB) Close() {
	db.db.Close()
}

// NewBucket creates a new bucket
func (db *DB) NewBucket(bucket string) error {
	db.Lock()
	defer db.Unlock()
	return db.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

// Set saves a value at the given key.
func (db *DB) Set(bucket, key string, value interface{}) error {
	db.Lock()
	defer db.Unlock()
	bValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(key), bValue)
		return err
	})
}

// Get will return the value associated with a key.
func (db *DB) Get(bucket, key string, v interface{}) (err error) {
	db.RLock()
	defer db.RUnlock()
	return db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		val := b.Get([]byte(key))
		if val == nil {
			return NoSuchKeyError{key}
		}
		return json.Unmarshal(val, &v)
	})
}

// Delete removes a key from the store.
func (db *DB) Delete(bucket, key string) error {
	db.Lock()
	defer db.Unlock()
	return db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Delete([]byte(key))
	})
}

// GetKeysInRange will return list of keys in that range
func (db *DB) GetKeysInRange(bucket, first, last string) (keys []string, err error) {
	db.RLock()
	defer db.RUnlock()
	keys = []string{}
	err = db.db.View(func(tx *bolt.Tx) error {
		// Assume our events bucket exists and has RFC3339 encoded time keys.
		c := tx.Bucket([]byte(bucket)).Cursor()

		// Our time range spans the 90's decade.
		min := []byte(first)
		max := []byte(last)

		if first == "first" && last == "last" {
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				keys = append(keys, string(k))
			}
		} else if first == "first" {
			for k, _ := c.First(); k != nil && bytes.Compare(k, max) < 0; k, _ = c.Next() {
				keys = append(keys, string(k))
			}
		} else if last == "last" {
			for k, _ := c.Seek(min); k != nil; k, _ = c.Next() {
				keys = append(keys, string(k))
			}
		} else {
			for k, _ := c.Seek(min); k != nil && bytes.Compare(k, max) < 0; k, _ = c.Next() {
				keys = append(keys, string(k))
			}
		}

		return nil
	})
	return
}
