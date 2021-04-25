package kv

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func dbCheck(tt *testing.T, fn func(KV, *rapid.T)) {
	rapid.Check(tt, func(t *rapid.T) {
		kv, err := NewInMemoryBadgerKV()
		require.NoError(t, err, "Unable to open a new instance of Badger DB")
		defer func() {
			require.NoError(t, kv.Close())
		}()

		fn(kv, t)
	})
}

func TestRandomKeys(t *testing.T) {
	dbCheck(t, func(kv KV, t *rapid.T) {
		k := rapid.
			SliceOf(rapid.Byte()).
			Filter(func(_k []byte) bool { return len(_k) > 0 }).
			Draw(t, "key").([]byte)
		v := rapid.SliceOf(rapid.Byte()).Draw(t, "value").([]byte)

		err := kv.Set(k, &Object{Value: v})
		require.NoError(t, err)

		obj, err := kv.Get(k)
		require.NoError(t, err)
		if len(v) > 0 {
			require.Equal(t, v, obj.Value)
		} else {
			require.Empty(t, obj.Value)
		}
	})
}

type kvMachine struct {
	kv    KV
	keys  []string
	mapKv map[string]Object
}

func (m *kvMachine) Init(t *rapid.T) {
	kv, err := NewInMemoryBadgerKV()
	require.NoError(t, err)

	m.keys = rapid.
		SliceOf(rapid.StringMatching("\\w+")).
		Filter(func(ks []string) bool { return len(ks) > 0 }).
		Draw(t, "keys").([]string)

	m.kv = kv
	m.mapKv = make(map[string]Object)
}

func genKey(m *kvMachine, t *rapid.T) ([]byte, string) {
	strKey := rapid.SampledFrom(m.keys).Draw(t, "a key").(string)
	return []byte(strKey), strKey
}

func (m *kvMachine) Cleanup() {
	m.kv.Close()
}

func (m *kvMachine) Check(t *rapid.T) {}

func (m *kvMachine) InitKey(t *rapid.T) {
	byteK, k := genKey(m, t)
	initVal := Object{Value: []byte("@")}

	// pre condition
	if _, ok := m.mapKv[k]; ok {
		t.SkipNow()
	}

	err := m.kv.Set(byteK, &initVal)
	require.NoError(t, err)

	// Model upd
	m.mapKv[k] = initVal
}

func (m *kvMachine) Get(t *rapid.T) {
	byteK, k := genKey(m, t)

	obj, err := m.kv.Get(byteK)

	// post condition
	elem, ok := m.mapKv[k]
	if !ok {
		require.Equal(t, ErrNotFound, err)
		require.Nil(t, obj)
	} else {
		require.NoError(t, err)
		require.Equal(t, elem, *obj)
	}
}

func (m *kvMachine) Update(t *rapid.T) {
	byteK, k := genKey(m, t)

	// pre cond.
	if _, ok := m.mapKv[k]; !ok {
		t.SkipNow()
	}

	obj, err := m.kv.Get(byteK)
	require.NotNil(t, obj)
	require.NoError(t, err)

	newVal := append(obj.Value, '!')

	obj.Value = newVal
	err = m.kv.Set(byteK, obj)
	require.NoError(t, err)

	// model upd
	m.mapKv[k] = Object{Value: newVal, Gen: m.mapKv[k].Gen + 1}
}

func TestStatefulKv(t *testing.T) {
	rapid.Check(t, rapid.Run(&kvMachine{}))
}
