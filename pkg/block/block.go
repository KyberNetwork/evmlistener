package block

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Block contains information of block.
type Block struct {
	Number      *big.Int
	Hash        common.Hash
	ParentHash  common.Hash
	ReorgedHash common.Hash
	Logs        []types.Log
}

// Storage is an interface for interacting with block storage.
type Storage interface {
	Add(b Block) error
	Exists(hash common.Hash) (bool, error)
	Head() (Block, error)
	Get(hash common.Hash) (Block, error)
	IsReorg(b Block) (bool, error)
	GetRecentBlocks(n int) ([]Block, error)
}
