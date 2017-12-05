package node

import (
	"container/list"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core/chain"
	"github.com/u-speak/core/config"
	d "github.com/u-speak/core/node/protoc"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	// MaxMsgSize specifies the largest packet size for grpc calls
	MaxMsgSize = 5242880
)

// Node is a wrapper around the chain. Nodes are the backbone of the network
type Node struct {
	PostChain        *chain.Chain
	ImageChain       *chain.Chain
	KeyChain         *chain.Chain
	ListenInterface  string
	Version          string
	remoteInterfaces map[string]struct{}
	Hooks            struct {
		PreAdd string
	}
}

type ChainStatus struct {
	Valid    bool   `json:"valid"`
	Length   uint64 `json:"length"`
	LastHash string `json:"last_hash"`
}

type ChainStatusList struct {
	Post  ChainStatus `json:"post"`
	Image ChainStatus `json:"image"`
	Key   ChainStatus `json:"key"`
}

// Status is used for reporting this nodes configuration to other nodes
type Status struct {
	Address     string          `json:"address"`
	Version     string          `json:"version"`
	Length      uint64          `json:"length"`
	Connections []string        `json:"connections"`
	Chains      ChainStatusList `json:"chains"`
}

func validateAll(chain.Hash) bool {
	return true
}

// New constructs a new node from the configuration
func New(c config.Configuration) (*Node, error) {
	ic, err := chain.New(&chain.BoltStore{Path: c.Storage.BoltStore.ImagePath}, validateAll)
	if err != nil {
		return nil, err
	}
	kc, err := chain.New(&chain.BoltStore{Path: c.Storage.BoltStore.KeyPath}, validateAll)
	if err != nil {
		return nil, err
	}
	pc, err := chain.New(&chain.BoltStore{Path: c.Storage.BoltStore.PostPath}, validateAll)
	if err != nil {
		return nil, err
	}
	return &Node{
		ListenInterface:  c.NodeNetwork.Interface + ":" + strconv.Itoa(c.NodeNetwork.Port),
		ImageChain:       ic,
		KeyChain:         kc,
		PostChain:        pc,
		Version:          c.Version,
		remoteInterfaces: make(map[string]struct{}),
		Hooks:            c.Hooks,
	}, nil
}

// encHash returns the String encoded Hash
func encHash(h [32]byte) string {
	return base64.URLEncoding.EncodeToString(h[:])
}

// Status returns the current running configuration of the node
func (n *Node) Status() Status {
	cons := []string{}
	for k := range n.remoteInterfaces {
		cons = append(cons, k)
	}
	return Status{
		Address: n.ListenInterface,
		Length:  n.PostChain.Length() + n.KeyChain.Length() + n.ImageChain.Length(),
		Chains: ChainStatusList{
			Post:  ChainStatus{Length: n.PostChain.Length(), Valid: n.PostChain.Valid(), LastHash: encHash(n.PostChain.LastHash())},
			Image: ChainStatus{Length: n.ImageChain.Length(), Valid: n.ImageChain.Valid(), LastHash: encHash(n.ImageChain.LastHash())},
			Key:   ChainStatus{Length: n.KeyChain.Length(), Valid: n.KeyChain.Valid(), LastHash: encHash(n.KeyChain.LastHash())},
		},
		Connections: cons,
		Version:     n.Version,
	}
}

// Info returns the serializable info struct
func (n *Node) Info() *d.Info {
	s := n.Status()
	return &d.Info{
		Length:          s.Length,
		Valid:           n.PostChain.Valid() && n.ImageChain.Valid() && n.KeyChain.Valid(),
		ListenInterface: s.Address,
	}
}

// GetInfo is a all purpose status request
func (n *Node) GetInfo(ctx context.Context, r *d.Info) (*d.Info, error) {
	if _, ok := n.remoteInterfaces[r.ListenInterface]; !ok && n.ListenInterface != r.ListenInterface {
		log.Infof("Establishing reverse connection with %s", r.ListenInterface)
		n.Connect(r.ListenInterface)
	}
	return n.Info(), nil
}

// Run listens for connections to this node
func (n *Node) Run() {
	log.Infof("Starting Nodeserver on %s", n.ListenInterface)
	lis, err := net.Listen("tcp", n.ListenInterface)
	if err != nil {
		log.Errorf("Could not listen on %s: %s", n.ListenInterface, err)
	}
	// Set MsgSize to 5MB
	grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(MaxMsgSize), grpc.MaxRecvMsgSize(MaxMsgSize))
	d.RegisterDistributionServiceServer(grpcServer, n)
	log.Fatal(grpcServer.Serve(lis))
}

// Connect connects to a new remote
func (n *Node) Connect(remote string) error {
	if _, ok := n.remoteInterfaces[remote]; ok {
		return errors.New("Attempted to add an allready established interface")
	}
	n.remoteInterfaces[remote] = struct{}{}
	conn, err := dial(remote)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := d.NewDistributionServiceClient(conn)
	i, err := client.GetInfo(context.Background(), n.Info())
	if err != nil {
		delete(n.remoteInterfaces, remote)
		return err
	}
	if !i.Valid {
		delete(n.remoteInterfaces, remote)
		return errors.New("Remote chain invalid")
	}
	if i.Length > n.Status().Length {
		err := n.SynchronizeChain(remote)
		if err != nil {
			delete(n.remoteInterfaces, remote)
			return err
		}
	}
	log.Infof("Added connection %s", remote)
	return nil
}

// SubmitBlock is called whenever a new block is submitted to the network
func (n *Node) SubmitBlock(b chain.Block) {
	log.Debug(n.PostChain)
	log.Infof("Pushing block %x to network", b.Hash())
	n.Push(&b)
}

// Push sends a block to all connected nodes
func (n *Node) Push(b *chain.Block) {
	h := b.PrevHash
	pb := &d.Block{
		Content:   b.Content,
		Nonce:     b.Nonce,
		Previous:  h[:],
		Signature: b.Signature,
		Date:      b.Date.Unix(),
		Type:      b.Type,
		PubKey:    b.PubKey,
	}
	for r := range n.remoteInterfaces {
		conn, err := dial(r)
		if err != nil {
			continue
		}
		client := d.NewDistributionServiceClient(conn)
		_, err = client.AddBlock(context.Background(), pb)
		if err != nil {
			log.Error(err)
		}
		err = conn.Close()
		if err != nil {
			log.Error(err)
		}
	}
}

// SmartAdd Adds Blocks to the specified chain
func (n *Node) SmartAdd(b chain.Block) {
	var c *chain.Chain
	switch b.Type {
	case "post":
		c = n.PostChain
	case "image":
		c = n.ImageChain
	case "key":
		c = n.KeyChain
	}
	c.Add(b)
}

// AddBlock receives a sent Block from other node or repl
func (n *Node) AddBlock(ctx context.Context, block *d.Block) (*d.PushReturn, error) {
	var p [32]byte
	copy(p[:], block.Previous)
	b := chain.Block{
		Content:   block.Content,
		Type:      block.Type,
		PubKey:    block.PubKey,
		Date:      time.Unix(block.Date, 0),
		Signature: block.Signature,
		PrevHash:  p,
		Nonce:     block.Nonce,
	}
	log.Debugf("Received Block with hash: %s", base64.URLEncoding.EncodeToString(b.Hash().Bytes()))

	switch b.Type {
	case "post":
		if p != n.PostChain.LastHash() {
			log.Errorf("Tried to add invalid Block! Previous hash %v is not valid. Please synchronize the nodes", p)
			return &d.PushReturn{}, errors.New("Received block had invalid previous hash")
		}

	case "image":
		if p != n.ImageChain.LastHash() {
			log.Errorf("Tried to add invalid Block! Previous hash %v is not valid. Please synchronize the nodes", p)
			return &d.PushReturn{}, errors.New("Received block had invalid previous hash")
		}
	case "key":
		if p != n.KeyChain.LastHash() {
			log.Errorf("Tried to add invalid Block! Previous hash %v is not valid. Please synchronize the nodes", p)
			return &d.PushReturn{}, errors.New("Received block had invalid previous hash")
		}

	}
	// PreAdd hook
	if n.Hooks.PreAdd != "" {
		u, err := url.Parse(n.Hooks.PreAdd)
		if err != nil {
			log.Errorf("Error running PreAdd hook: %s", err.Error())
		}
		q := u.Query()
		q.Add("hash", base64.URLEncoding.EncodeToString(b.Hash().Bytes()))
		u.RawQuery = q.Encode()
		log.Debugf("Calling PreAdd Hook with URL: %s", u.String())
		_, err = http.Get(u.String())
		if err != nil {
			log.Errorf("Error running PreAdd hook: %s", err.Error())
		}
	}
	n.SmartAdd(b)
	return &d.PushReturn{}, nil
}

// Synchronize sends all Blocks from all chains to an other node
func (n *Node) Synchronize(p *d.SyncParams, stream d.DistributionService_SynchronizeServer) error {
	log.Infof("Synchronization started. Sending all Blocks to another Node.")
	names := []string{"postchain", "imagechain", "keychain"}
	c := [32]byte{}
	h := n.PostChain.LastHash()
	b := n.PostChain.Get(h)
	var blst list.List
	for k := 0; k < 3; k++ {
		if k == 0 {
			h = n.PostChain.LastHash()
			b = n.PostChain.Get(h)
		} else if k == 1 {
			h = n.ImageChain.LastHash()
			b = n.ImageChain.Get(h)
		} else {
			h = n.KeyChain.LastHash()
			b = n.KeyChain.Get(h)
		}

		for {
			blst.PushBack(b.Content)
			if b.PrevHash == c {
				break
			}
			if k == 0 {
				b = n.PostChain.Get(b.PrevHash)
			} else if k == 1 {
				b = n.ImageChain.Get(b.PrevHash)
			} else {
				b = n.KeyChain.Get(b.PrevHash)
			}
		}
		blk := []*chain.Block{}
		if k == 0 {
			blk, _ = n.PostChain.DumpChain()
		} else if k == 1 {
			blk, _ = n.ImageChain.DumpChain()
		} else {
			blk, _ = n.KeyChain.DumpChain()
		}

		for i := len(blk) - 2; i >= 0; i-- {
			err := stream.Send(&d.Block{
				Content:   blk[i].Content,
				Nonce:     blk[i].Nonce,
				Previous:  blk[i].PrevHash[:],
				Type:      blk[i].Type,
				PubKey:    blk[i].PubKey,
				Date:      blk[i].Date.Unix(),
				Signature: blk[i].Signature,
			})
			if err != nil {
				log.Error(err)
			}
		}
		log.Infof("Synchronization for %v finished successfully.", names[k])
	}
	return nil
}

// ReinitializeChain Re-Initializes all chains
func (n *Node) ReinitializeChain() {
	n.PostChain.Reinitialize()
	n.ImageChain.Reinitialize()
	n.KeyChain.Reinitialize()
}

// SynchronizeChain receives all the Blocks sent from an other node
func (n *Node) SynchronizeChain(remote string) error {
	n.ReinitializeChain()
	lhp := n.PostChain.LastHash()
	log.Infof("Synchronization started. Receiving Blocks from other node.")

	params := &d.SyncParams{LastHash: lhp[:]}
	conn, err := dial(remote)
	if err != nil {
		return err
	}
	client := d.NewDistributionServiceClient(conn)
	stream, err := client.Synchronize(context.Background(), params)
	if err != nil {
		return err
	}
	for {
		block, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		var p [32]byte
		copy(p[:], block.Previous)
		b := chain.Block{
			Content:   block.Content,
			Type:      block.Type,
			PubKey:    block.PubKey,
			Date:      time.Unix(block.Date, 0),
			Signature: block.Signature,
			PrevHash:  p,
			Nonce:     block.Nonce,
		}

		log.Infof("Got a new Block: %v", b.Type)
		log.Debugf("Received %+v", b)
		n.SmartAdd(b)
	}
	conn.Close()
	log.Infof("Synchronization finished successfully.")
	return nil
}

func dial(r string) (*grpc.ClientConn, error) {
	return grpc.Dial(r,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(MaxMsgSize),
			grpc.MaxCallSendMsgSize(MaxMsgSize),
		))
}

func runHook(s string) {
}
