package post

//go:generate msgp

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/u-speak/core/tangle/hash"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// Post contains all information needed for a complete post representation
type Post struct {
	Content   string            `json:"content"`
	Pubkey    *openpgp.Entity   `msg:"-" json:"-"`
	Signature *packet.Signature `msg:"-" json:"-"`
	PubkeyStr string            `json:"pubkey"`
	SigStr    string            `json:"signature"`
	Timestamp int64             `json:"date"`
}

type serializable interface {
	Serialize(w io.Writer) error
}

// Hash returns the hashed post for storage
func (p *Post) Hash() (hash.Hash, error) {
	buff := bytes.NewBuffer(nil)
	p.Signature.Serialize(buff)
	sigdata := buff.Bytes()[3:]
	sighash := hash.New(sigdata)
	h := "C" + p.Content + "D" + strconv.FormatInt(p.Timestamp, 10) + "P" + p.Pubkey.PrimaryKey.KeyIdString() + "S" + sighash.String()
	return hash.New([]byte(h)), nil
}

// Verify returns no error when the signature is valid
func (p *Post) Verify() error {
	hash := p.Signature.Hash.New()
	tr := strings.Trim(p.Content, "\t\n\v\f\r \u0085\u00A0")
	_, err := io.Copy(hash, strings.NewReader(tr))
	if err != nil {
		return err
	}
	return p.Pubkey.PrimaryKey.VerifySignature(hash, p.Signature)
}

// Serialize implements tangle/datastore.serializable
func (p *Post) Serialize() ([]byte, error) {
	err := p.storePGPStr()
	if err != nil {
		return nil, err
	}
	return p.MarshalMsg(nil)
}

func (p *Post) storePGPStr() error {
	pk, err := asciiEncode(p.Pubkey, openpgp.PublicKeyType)
	if err != nil {
		return err
	}
	p.PubkeyStr = pk
	ss, err := asciiEncode(p.Signature, openpgp.SignatureType)
	if err != nil {
		return err
	}
	p.SigStr = ss
	return nil
}

// Deserialize implements tangle/datastore.serializable
func (p *Post) Deserialize(bts []byte) error {
	_, err := p.UnmarshalMsg(bts)
	if err != nil {
		return err
	}
	return p.ReInit()
}

// JSON prepares for json encoding
func (p *Post) JSON() error {
	return p.storePGPStr()
}

// ReInit restores the original field after serialization
func (p *Post) ReInit() error {
	pub, err := asciiDecodeEntity(p.PubkeyStr)
	if err != nil {
		return err
	}
	p.Pubkey = pub

	sigpkt, err := asciiDecode(p.SigStr)
	if err != nil {
		return err
	}
	sig, ok := sigpkt.(*packet.Signature)
	if !ok {
		return errors.New("Wrong Block type for signature")
	}
	p.Signature = sig
	return nil
}

// Type implements tangle/datastore.serializable
func (p *Post) Type() string {
	return "post"
}

func asciiDecodeEntity(s string) (*openpgp.Entity, error) {
	buff := strings.NewReader(s)
	block, err := armor.Decode(buff)
	if err != nil {
		return nil, err
	}
	reader := packet.NewReader(block.Body)
	return openpgp.ReadEntity(reader)
}

func asciiEncode(s serializable, blockType string) (string, error) {
	buff := bytes.NewBuffer(nil)
	wr, err := armor.Encode(buff, blockType, make(map[string]string))
	if err != nil {
		return "", err
	}
	err = s.Serialize(wr)
	if err != nil {
		return "", err
	}
	err = wr.Close()
	if err != nil {
		return "", err
	}
	return buff.String(), nil

}

func asciiDecode(s string) (packet.Packet, error) {
	buff := strings.NewReader(s)
	block, err := armor.Decode(buff)
	if err != nil {
		return nil, err
	}
	reader := packet.NewReader(block.Body)
	return reader.Next()
}
