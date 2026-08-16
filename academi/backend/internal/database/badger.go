package database

import (
	"log"
	"sync"

	"github.com/academi/backend/internal/config"
	"github.com/dgraph-io/badger/v4"
)

var (
	db   *badger.DB
	once sync.Once
)

func GetDB() *badger.DB {
	once.Do(func() {
		cfg := config.Load()
		opts := badger.DefaultOptions(cfg.Database.Path).
			WithLogger(nil).
			WithLoggingLevel(badger.ERROR)

		var err error
		db, err = badger.Open(opts)
		if err != nil {
			log.Fatalf("Failed to open BadgerDB: %v", err)
		}
	})
	return db
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}

func Set(key []byte, value []byte) error {
	return GetDB().Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func Get(key []byte) ([]byte, error) {
	var val []byte
	err := GetDB().View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	return val, err
}

func Delete(key []byte) error {
	return GetDB().Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func Iterate(prefix []byte, fn func(key, value []byte) error) error {
	return GetDB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			if err := fn(item.Key(), val); err != nil {
				return err
			}
		}
		return nil
	})
}
