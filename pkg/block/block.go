package block

import (
	"github.com/KyberNetwork/evmlistener/pkg/types"
)

// Keeper is an interface for interacting with block keeper.
type Keeper interface {
	Init() error
	Len() int
	Cap() int
	Add(b types.Block) error
	Exists(hash string) (bool, error)
	Head() (types.Block, error)
	Get(hash string) (types.Block, error)
	IsReorg(b types.Block) (bool, error)
	GetRecentBlocks(n int) ([]types.Block, error)
	GetHead() string
	SetHead(hash string)
}
