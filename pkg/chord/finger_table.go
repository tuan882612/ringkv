package chord

import (
	"math/big"

	"github.com/rs/zerolog/log"
)

var (
	two      = big.NewInt(2)
	modValue = big.NewInt(0).Exp(two, big.NewInt(M), nil) // 2^M for mod operation
)

type (
	// FingerTable is the interface for the finger table.
	// Keeps track of the successor nodes for each position in the ring
	FingerTable interface {
		getFinger(i int) *NodeInfo
		update(node *Node)
		ininterval(id, start, end *big.Int) bool
	}

	fingerTableImpl struct {
		table []*NodeInfo
	}
)

// Initialize the finger table with the node itself as the successor for all positions
func NewFingerTable(node *Node) FingerTable {
	tab := &fingerTableImpl{
		make([]*NodeInfo, M),
	}

	for i := 0; i < M; i++ {
		tab.table[i] = node.getInfo()
	}

	return tab
}

func (ft *fingerTableImpl) getFinger(i int) *NodeInfo {
	return ft.table[i]
}

// update updates the finger table of the node
func (ft *fingerTableImpl) update(node *Node) {
	log.Info().Str("node_id", node.ID.String()).Msgf("%s: Updating finger table", node.Address)
	for i := 0; i < M; i++ {
		// Calculate start = (node.ID + 2^i) % 2^M
		exp := big.NewInt(0).Exp(two, big.NewInt(int64(i)), nil) // 2^i
		start := big.NewInt(0).Add(node.ID, exp)                 // node.ID + 2^i
		start.Mod(start, modValue)                               // (node.ID + 2^i) % 2^M

		// Use start to find the successor node for this position in the finger table
		succ, err := node.findSuccessor(start)
		if err != nil {
			log.Error().
				Err(err).Str("node_id", node.ID.String()).
				Msgf("%s: Failed to update finger table", node.Address)
			continue
		}

		ft.table[i] = succ
	}
}

// ininterval checks if id is in the interval (start, end]
func (ft *fingerTableImpl) ininterval(id, start, end *big.Int) bool {
	if start.Cmp(end) == 0 {
		return true
	}

	if start.Cmp(end) < 0 {
		return id.Cmp(start) > 0 && id.Cmp(end) <= 0
	}

	return id.Cmp(start) > 0 || id.Cmp(end) <= 0
}
