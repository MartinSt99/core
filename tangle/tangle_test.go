package tangle

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/site"
	"github.com/u-speak/core/tangle/store"
	"github.com/u-speak/core/tangle/store/boltstore"
	"github.com/u-speak/core/tangle/store/memorystore"
)

func ms() *memorystore.MemoryStore {
	ms := memorystore.MemoryStore{}
	_ = ms.Init(store.Options{})
	return &ms
}

func TestInit(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init(Options{Store: ms()})
	assert.NoError(t, err)
	assert.Equal(t, 2, tngl.Size())
}

func TestTips(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init(Options{Store: ms()})
	assert.NoError(t, err)
	assert.Len(t, tngl.Tips(), 2)
}

func TestGet(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init(Options{Store: ms()})
	assert.NoError(t, err)
	assert.Nil(t, tngl.Get(hash.Hash{}))
	for _, s := range tngl.Tips() {
		assert.Equal(t, s, tngl.Get(s.Hash()))
	}
}

func TestAdd(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init(Options{Store: ms()})
	assert.NoError(t, err)
	tips := tngl.Tips()
	err = tngl.Add(&site.Site{Content: hash.Hash{1, 3, 3, 7}, Nonce: 0})
	assert.Equal(t, ErrWeightTooLow, err)
	err = tngl.Add(&site.Site{Content: hash.Hash{1, 3, 3, 7}, Nonce: 263})
	assert.Equal(t, ErrTooFewValidations, err)

	sub := &site.Site{Content: hash.Hash{1, 3, 3, 7}, Nonce: 0, Validates: []*site.Site{tips[0], tips[1]}}
	sub.Mine(1)
	err = tngl.Add(sub)
	assert.NoError(t, err)
	assert.False(t, tngl.tips[tips[0]])
	assert.False(t, tngl.tips[tips[1]])
	assert.True(t, tngl.tips[sub])
	assert.Equal(t, sub, tngl.Get(sub.Hash()))
}

func TestRestore(t *testing.T) {
	tngl := Tangle{}
	bs := boltstore.BoltStore{}
	err := bs.Init(store.Options{Path: path.Join(os.TempDir(), "testRestore.db")})
	assert.NoError(t, err)
	err = tngl.Init(Options{Store: &bs})
	assert.NoError(t, err)
	tips := tngl.Tips()
	sub := &site.Site{Content: hash.Hash{1, 3, 3, 7}, Nonce: 0, Validates: []*site.Site{tips[0], tips[1]}}
	sub.Mine(1)
	err = tngl.Add(sub)
	assert.NoError(t, err)
	tips = tngl.Tips()
	tngl.Close()

	bs2 := boltstore.BoltStore{}
	err = bs2.Init(store.Options{Path: path.Join(os.TempDir(), "testRestore.db")})
	assert.NoError(t, err)
	tngl2 := Tangle{}
	err = tngl2.Init(Options{Store: &bs2})
	assert.Equal(t, tips, tngl2.Tips())

}
