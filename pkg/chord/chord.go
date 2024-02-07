package chord

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

const (
	// M is the number of bits in the hash
	M = 160 // SHA1
)

// Bootstrap initializes the node with itself as the successor and predecessor creating a new ring
func (n *Node) Bootstrap() error {
	n.Successor = n.getInfo()
	n.Predecessor = n.getInfo()
	if err := n.Transport.listen(n); err != nil {
		return err
	}
	return nil
}

// Join initializes the node by joining an existing ring
func (n *Node) JoinRing(bootstrapAddr string) error {
	if err := n.Transport.listen(n); err != nil {
		return err
	}

	// Find the successor of the new node
	req := NewRPCRequest(FindSuccessorMethod, n.ID.Bytes())
	data, err := n.Transport.invokeRPC(bootstrapAddr, req)
	if err != nil {
		return err
	}

	successor := &NodeInfo{}
	if err := jsoniter.Unmarshal(data, successor); err != nil {
		log.Error().Err(err).Msgf("%s: Failed to unmarshal successor", n.Address)
		return err
	}

	// Update the finger table and successor
	n.Successor = successor
	n.Table.update(n)
	return nil
}
