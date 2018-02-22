package site

import (
	"strconv"

	"github.com/u-speak/core/tangle/hash"
	"github.com/vmihailenco/msgpack"
)

// Site represents a single storage node inside the tangle
type Site struct {
	Validates []*Site
	Nonce     uint64
	Content   hash.Hash
	Type      string
}

// Hash computes the hash of the site
func (s *Site) Hash() hash.Hash {
	ts := "C" + s.Content.String() + "N" + strconv.FormatUint(s.Nonce, 10) + "T" + s.Type
	for _, s := range s.Validates {
		ts += "V" + s.Hash().String()
	}
	return hash.New([]byte(ts))
}

// Serialize converts the site to a slice of bytes
func (s *Site) Serialize() []byte {
	b, _ := msgpack.Marshal(s)
	return b
}

// Deserialize restores the site from a slice of bytes
func (s *Site) Deserialize(b []byte) error {
	return msgpack.Unmarshal(b, s)
}

// Mine the block for a specifig weight
func (s *Site) Mine(targetWeight int) {
	for s.Hash().Weight() < targetWeight {
		s.Nonce++
	}
}
