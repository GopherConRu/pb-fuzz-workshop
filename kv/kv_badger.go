package kv

import (
	"bytes"
	"encoding/gob"
	"hash/maphash"
	"sync"

	badger "github.com/dgraph-io/badger/v3"
)

const cacheSize = 3000

type cachedGenPair struct {
	gen uint64
	key []byte
}

type badgerKV struct {
	db *badger.DB

	seed     maphash.Seed
	m        sync.Mutex
	genCache map[uint16]cachedGenPair
}

func NewInMemoryBadgerKV() (KV, error) {
	opts := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &badgerKV{db: db, seed: maphash.MakeSeed(), genCache: make(map[uint16]cachedGenPair)}, err
}

func (kv *badgerKV) computeCacheKey(k []byte) uint16 {
	var h maphash.Hash
	h.SetSeed(kv.seed)
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
	kv.genCache[kv.computeCacheKey(k)] = cachedGenPair{gen: obj.Gen, key: k}
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
		if cachedGen.gen > o.Gen && bytes.Equal(cachedGen.key, k) {
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
				if oldObj.Gen > o.Gen {
					return ErrOldGen
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

	newObj := Object{
		Value: o.Value,
		Gen:   o.Gen + 1,
	}
	err := enc.Encode(&newObj)
	if err != nil {
		panic(err)
	}

	err = tx.Set(k, buf.Bytes())
	if err != nil {
		return err
	}
	kv.genCache[cacheKey] = cachedGenPair{gen: newObj.Gen, key: k}

	return tx.Commit()
}

func (kv *badgerKV) Close() error {
	return kv.db.Close()
}
