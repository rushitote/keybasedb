package main

import (
	"sort"

	badger "github.com/dgraph-io/badger/v4"
)

// Storage Engine Interface
type Engine struct {
	db *badger.DB
	mt *MerkleTree
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

func (e *Engine) Stream(f func(key string, value string) error) {
	e.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			var value string
			err := item.Value(func(val []byte) error {
				value = string(val)
				return nil
			})
			if err != nil {
				panic(err)
			}
			err = f(key, value)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (e *Engine) CreateMerkleTree(hashRange HashRange) {
	e.mt = CreateMerkleTree(hashRange)
	kvHashLists := make([][]string, len(e.mt.LeafNodes))
	// iterate over all keys in the range and add them to the merkle tree
	e.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())
			keyHash := GenerateHash(key)
			idx := GetMTLeafIndex(keyHash, e.mt.Root)
			if idx == -1 {
				continue
			}
			var value string
			err := item.Value(func(val []byte) error {
				value = string(val)
				return nil
			})
			if err != nil {
				panic(err)
			}
			kvHash := GenerateHash(key + value)
			kvHashLists[idx] = append(kvHashLists[idx], kvHash)
		}
		return nil
	})
	for i, kvHashList := range kvHashLists {
		sort.Strings(kvHashList)
		e.mt.LeafNodes[i].Hash = GenerateHashOfList(kvHashList)
	}
}

const (
	DeletedHash = "hefiwhe783d7qdiq83"
)
