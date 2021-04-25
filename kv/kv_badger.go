package kv

import (
	"bytes"
	"encoding/gob"
	"hash/maphash"
	"sync"

	badger "github.com/dgraph-io/badger/v3"
)

const cacheSize = 3000

type badgerKV struct {
	db *badger.DB

	m        sync.Mutex
	genCache map[uint16]uint64
}

func NewInMemoryBadgerKV() (KV, error) {
	opts := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &badgerKV{db: db, genCache: make(map[uint16]uint64)}, err
}

func (kv *badgerKV) computeCacheKey(k []byte) uint16 {
	var h maphash.Hash
	h.Write(k)
	return uint16(h.Sum64()) % cacheSize
}

func (kv *badgerKV) Get(k Key) (*Object, error) {
	var obj Object
	err := kv.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(k)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotFound
			}
			return err
		}

		val, err := item.ValueCopy(nil)
		dec := gob.NewDecoder(bytes.NewBuffer(val))
		if err != nil {
			return err
		}

		return dec.Decode(&obj)
	})
	if err != nil {
		return nil, err
	}

	kv.m.Lock()
	kv.genCache[kv.computeCacheKey(k)] = obj.Gen
	kv.m.Unlock()
	return &obj, nil
}

func (kv *badgerKV) Set(k Key, o *Object) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	tx := kv.db.NewTransaction(true)
	defer tx.Discard()

	cacheKey := kv.computeCacheKey(k)

	kv.m.Lock()
	defer kv.m.Unlock()

	if cachedGen, ok := kv.genCache[cacheKey]; ok {
		if cachedGen > o.Gen {
			return ErrOldGen
		}
	} else {
		item, err := tx.Get(k)
		if err == nil {
			var oldObj Object
			err = item.Value(func(val []byte) error {
				dec := gob.NewDecoder(bytes.NewBuffer(val))
				decodeError := dec.Decode(&oldObj)
				if decodeError != nil {
					return decodeError
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else if err != badger.ErrKeyNotFound {
			return err
		}
	}

	o.Gen += 1
	err := enc.Encode(&o)
	if err != nil {
		panic(err)
	}

	err = tx.Set(k, buf.Bytes())
	if err != nil {
		return err
	}

	kv.genCache[cacheKey] = o.Gen
	return tx.Commit()
}

func (kv *badgerKV) Close() error {
	return kv.db.Close()
}
