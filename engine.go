package main

import (
	badger "github.com/dgraph-io/badger/v4"
)

// Storage Engine Interface
type Engine struct {
	db *badger.DB
}

func CreateEngine(dir string) *Engine {
	opts := badger.DefaultOptions("/tmp/badger/" + dir)
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	return &Engine{
		db: db,
	}
}

func (e *Engine) Read(key string) (string, error) {
	var value string
	err := e.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			value = string(val)
			return nil
		})
		return err
	})
	if err != nil && err != badger.ErrKeyNotFound {
		return "", err
	} else if err == badger.ErrKeyNotFound {
		return "", nil
	}
	return value, nil
}

func (e *Engine) Write(key, value string) error {
	return e.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
}

func (e *Engine) Delete(key string) {
	e.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

const (
	DeletedHash = "hefiwhe783d7qdiq83"
)
