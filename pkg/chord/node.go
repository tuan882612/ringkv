package chord

import (
	"crypto/sha1"
	"math/big"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

type NID = *big.Int // Node ID

// node represents a node in the Chord ring or peer on the network
type (
	Node struct {
		ID          NID
		Address     string
		Successor   *NodeInfo
		Predecessor *NodeInfo
		Table       FingerTable
		Store       BlockStore
		Transport   Transport
	}

	NodeInfo struct {
		ID      NID
		Address string
	}
)

// Use SHA1 to hash the address and convert to a big.Int
func makeNodeID(address string) *big.Int {
	hash := sha1.New()
	hash.Write([]byte(address))
	return new(big.Int).SetBytes(hash.Sum(nil))
}

// NewNode creates a new node with the given address.
// The node is pre-initialized with an empty store, transport and finger table.
func NewNode(address string) *Node {
	node := &Node{
		ID:        makeNodeID(address),
		Address:   address,
		Store:     NewStore(),
		Transport: NewTransport(),
	}
	node.Table = NewFingerTable(node)
	return node
}

func (n *Node) getInfo() *NodeInfo {
	return &NodeInfo{
		ID:      n.ID,
		Address: n.Address,
	}
}

// used to find the successor of a node with a given id in the ring
func (n *Node) findSuccessor(id NID) (*NodeInfo, error) {
	if n.ID.Cmp(id) == 0 {
		return n.getInfo(), nil
	}

	if n.Table.ininterval(id, n.ID, n.Successor.ID) {
		return n.Successor, nil
	}

	// Use the finger table to find the closest preceding node
	node := n.closestPrecedingNode(id)
	req := NewRPCRequest(FindSuccessorMethod, id.Bytes())
	data, err := n.Transport.invokeRPC(node.Address, req)
	if err != nil {
		return nil, err
	}

	successor := &NodeInfo{}
	if err := jsoniter.Unmarshal(data, successor); err != nil {
		return nil, err
	}

	log.Info().
		Str("node_id", node.ID.String()).Str("successor_id", successor.ID.String()).
		Msgf("%s: Found successor at %s", n.Address, successor.Address)
	return successor, nil
}

// closestPrecedingNode finds the closest preceding node to the given id in the finger table
func (n *Node) closestPrecedingNode(id NID) *NodeInfo {
	for i := M - 1; i >= 0; i-- {
		if n.Table.ininterval(n.Table.getFinger(i).ID, n.ID, id) {
			return n.Table.getFinger(i)
		}
	}

	return n.Successor
}

// findPredecessor is used to find the predecessor of a node with a given id in the ring
func (n *Node) findPredecessor(id NID) (*NodeInfo, error) {
	for !(id.Cmp(n.ID) > 0 && id.Cmp(n.Successor.ID) <= 0) {
		if n.ID.Cmp(n.Successor.ID) >= 0 {
			break
		}

		node := n.closestPrecedingNode(id)
		if node.ID.Cmp(n.ID) == 0 {
			break
		}

		req := NewRPCRequest(FindPredecessorMethod, id.Bytes())
		data, err := n.Transport.invokeRPC(node.Address, req)
		if err != nil {
			return nil, err
		}

		pred := &NodeInfo{}
		if err := jsoniter.Unmarshal(data, pred); err != nil {
			log.Error().
				Err(err).Str("node_id", n.ID.String()).
				Msgf("%s: Failed to decode predecessor", n.Address)
			return nil, err
		}

		n.Predecessor = pred
	}

	log.Info().
		Str("node_id", n.ID.String()).Str("predecessor_id", n.Predecessor.ID.String()).
		Msgf("%s: Found predecessor at %s", n.Address, n.Predecessor.Address)
	return n.Predecessor, nil
}

// notify is called periodically to verify the node's immediate successor and update its predecessor
func (n *Node) notify(node *NodeInfo) error {
	if n.Predecessor == nil || n.Table.ininterval(node.ID, n.Predecessor.ID, n.ID) {
		n.Predecessor = node
	}

	return nil
}

// stabilize is called periodically to verify the node's immediate successor and tell the successor about the node
func (n *Node) stabilize() {
	log.Info().Str("node_id", n.ID.String()).Msgf("%s: Starting stabilize", n.Address)
	ticker := time.NewTicker(time.Second * 10) // Run every 10 seconds
	go func() {
		for range ticker.C {
			req := NewRPCRequest(FindPredecessorMethod, n.Successor.ID.Bytes())
			res, err := n.Transport.invokeRPC(n.Successor.Address, req)
			if err != nil {
				log.Error().
					Err(err).Str("node_id", n.ID.String()).
					Msgf("%s: Failed to find predecessor", n.Address)
				continue
			}

			pred := &NodeInfo{}
			if err := jsoniter.Unmarshal(res, pred); err != nil {
				log.Error().
					Err(err).Str("node_id", n.ID.String()).
					Msgf("%s: Failed to decode predecessor", n.Address)
				continue
			}

			start := new(big.Int).Add(n.ID, big.NewInt(1))
			if n.Table.ininterval(pred.ID, start, n.Successor.ID) {
				n.Successor = pred
			}

			data, err := jsoniter.Marshal(n.getInfo())
			if err != nil {
				log.Error().
					Err(err).Str("node_id", n.ID.String()).
					Msgf("%s: Failed to encode node info", n.Address)
				continue
			}

			req = NewRPCRequest(NotifyMethod, data)
			if _, err := n.Transport.invokeRPC(n.Successor.Address, req); err != nil {
				log.Error().
					Err(err).Str("node_id", n.ID.String()).
					Msgf("%s: Failed to notify successor", n.Address)
			}
		}
	}()
}
