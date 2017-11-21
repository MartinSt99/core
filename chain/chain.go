package chain

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ValidationFunc is the requirement for mining
type ValidationFunc func([32]byte) bool

var (
	// ErrInvalidChain gets returned when the chain validation fails
	ErrInvalidChain = errors.New("Chain Validation Failed")
	// ErrStoreInitialized gets returned when a store is tried to be initialized twice
	ErrStoreInitialized = errors.New("Store already initialized")
)

// Chain is a Blockchain Implementation
type Chain struct {
	blocks   BlockStore
	lastHash [32]byte
	validate ValidationFunc
}

// New initializes a new Chain
func New(b BlockStore, validate ValidationFunc) (*Chain, error) {
	lh, err := b.Init()
	if err != nil {
		return nil, err
	}
	c := &Chain{blocks: b, validate: validate}
	c.lastHash = lh
	if !c.Valid() {
		log.WithField("store", b).Error("Could not initialize Chain")
		return nil, ErrInvalidChain
	}
	return c, nil
}

// Add adds a block to the chain
func (c *Chain) Add(b Block) ([32]byte, error) {
	if !c.Valid() {
		return [32]byte{}, ErrInvalidChain
	}
	hash := b.Hash()
	if !c.validate(hash) {
		return [32]byte{}, errors.New("Block did not pass the validation function")
	}
	if b.PrevHash != c.lastHash {
		return [32]byte{}, errors.New("Blocks PrevHash was not the lasthash")
	}
	err := c.blocks.Add(b)
	if err != nil {
		return [32]byte{}, err
	}
	c.lastHash = hash
	return hash, nil
}

// DumpChain dumps the whole ordered chain in an array
func (c *Chain) DumpChain() ([]*Block, error) {
	if !c.Valid() {
		return []*Block{}, ErrInvalidChain
	}
	h := c.lastHash
	bl := []*Block{}
	for h != [32]byte{} {
		b := c.Get(h)
		bl = append(bl, b)
		h = b.PrevHash
	}
	return bl, nil
}

// Get retrieves a block
func (c *Chain) Get(hash [32]byte) *Block {
	return c.blocks.Get(hash)
}

// Valid checks the chain for integrity and validation compliance
func (c *Chain) Valid() bool {
	return c.blocks.Valid(c.validate) && c.Get(c.lastHash) != nil
}

// LastHash returns the hash of the last block in the chain
func (c *Chain) LastHash() [32]byte {
	return c.lastHash
}

// Length returns the length of the whole chain
func (c *Chain) Length() uint64 {
	return c.blocks.Length()
}

// Latest returns the latest n blocks
func (c *Chain) Latest(n int) ([]*Block, error) {
	if !c.Valid() {
		return nil, ErrInvalidChain
	}
	b := c.Get(c.lastHash)
	bs := []*Block{b}
	for i := 0; i < n; i++ {
		b = c.Get(b.PrevHash)
		if b == nil {
			break
		}
		bs = append(bs, b)
	}
	return bs, nil
}

// Search performs a simple string search on the Content of each block
func (c *Chain) Search(query string) []*Block {
	if !c.Valid() {
		return []*Block{}
	}
	b := c.Get(c.lastHash)
	bs := []*Block{}
	for b != nil {
		if strings.Contains(b.Content, query) {
			bs = append(bs, b)
		}
		b = c.Get(b.PrevHash)
	}
	return bs
}

//Reinitialize clears the Chain
func (c *Chain)Reinitialize() ([32]byte, error) {
	lh, err := c.blocks.Reinitialize()
	if err != nil {
	log.Errorf("Error initializing Chain. %+v", err)
	}
	c.lastHash = lh
	return c.blocks.Reinitialize()

}
