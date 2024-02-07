package chord

import "sync"

type (
	BlockStore interface{}

	Block struct{}

	blockStoreImpl struct {
		blocks map[string]*Block
		mu     sync.RWMutex
	}
)

func NewStore() BlockStore {
	return &blockStoreImpl{
		blocks: make(map[string]*Block),
	}
}
