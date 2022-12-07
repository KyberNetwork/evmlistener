package listener

import (
	"github.com/KyberNetwork/evmlistener/pkg/block"
	ltypes "github.com/KyberNetwork/evmlistener/pkg/types"
)

type BlockKeeperMock struct {
	*block.BaseBlockKeeper

	initBlocks []ltypes.Block
}

func NewBlockKeeperMock(n int) *BlockKeeperMock {
	return &BlockKeeperMock{
		BaseBlockKeeper: block.NewBaseBlockKeeper(n),
		initBlocks:      nil,
	}
}

func (k *BlockKeeperMock) SetInitData(blocks []ltypes.Block) {
	k.initBlocks = blocks
}

func (k *BlockKeeperMock) Init() error {
	err := k.BaseBlockKeeper.Init()
	if err != nil {
		return err
	}

	if len(k.initBlocks) > 0 {
		for _, b := range k.initBlocks {
			err = k.BaseBlockKeeper.Add(b)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
