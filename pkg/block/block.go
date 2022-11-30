package block

import (
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

// Keeper is an interface for interacting with block keeper.
type Keeper interface {
	Init() error
	Len() int
	Cap() int
	Add(b types.Block) error
	Exists(hash common.Hash) (bool, error)
	Head() (types.Block, error)
	Get(hash common.Hash) (types.Block, error)
	IsReorg(b types.Block) (bool, error)
	GetRecentBlocks(n int) ([]types.Block, error)
}
