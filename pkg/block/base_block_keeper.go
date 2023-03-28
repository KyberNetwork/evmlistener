package block

import (
	"fmt"
	"sync"

	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/emirpasic/gods/queues/circularbuffer"
)

// BaseBlockKeeper is a purely on-memory block keeper.
type BaseBlockKeeper struct {
	mu           sync.RWMutex
	maxNumBlocks int
	head         string
	blockMap     map[string]types.Block
	queue        *circularbuffer.Queue
}

// NewBaseBlockKeeper ...
func NewBaseBlockKeeper(maxNumBlocks int) *BaseBlockKeeper {
	if maxNumBlocks <= 0 {
		panic("max-num-blocks is not positive")
	}

	return &BaseBlockKeeper{
		mu:           sync.RWMutex{},
		maxNumBlocks: maxNumBlocks,
		head:         "",
		blockMap:     make(map[string]types.Block, maxNumBlocks),
		queue:        circularbuffer.New(maxNumBlocks),
	}
}

// Init ...
func (k *BaseBlockKeeper) Init() error {
	if len(k.blockMap) > 0 {
		k.blockMap = make(map[string]types.Block, k.maxNumBlocks)
		k.queue.Clear()
	}

	return nil
}

// Len ...
func (k *BaseBlockKeeper) Len() int {
	return len(k.blockMap)
}

// Cap ...
func (k *BaseBlockKeeper) Cap() int {
	return k.maxNumBlocks
}

func (k *BaseBlockKeeper) exists(hash string) bool {
	_, ok := k.blockMap[hash]

	return ok
}

// Exists checks whether a block hash is exists or not.
func (k *BaseBlockKeeper) Exists(hash string) (bool, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.exists(hash), nil
}

func (k *BaseBlockKeeper) removeOld() {
	v, ok := k.queue.Dequeue()
	if !ok {
		return
	}

	hash, ok := v.(string)
	if !ok {
		return
	}

	delete(k.blockMap, hash)
}

// Add adds new block to the keeper.
func (k *BaseBlockKeeper) Add(block types.Block) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if len(k.blockMap) > 0 {
		if k.exists(block.Hash) {
			return fmt.Errorf("block %v: %w", block.Hash, errors.ErrAlreadyExists)
		}

		if k.queue.Full() {
			k.removeOld()
		}

		if block.ParentHash != k.head {
			block.ReorgedHash = k.head
		} else {
			block.ReorgedHash = ""
		}
	}

	k.blockMap[block.Hash] = block
	k.head = block.Hash
	k.queue.Enqueue(block.Hash)

	return nil
}

// Head returns the block head of the chain on the keeper.
func (k *BaseBlockKeeper) Head() (types.Block, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.Get(k.head)
}

func (k *BaseBlockKeeper) get(hash string) (types.Block, error) {
	block, ok := k.blockMap[hash]
	if !ok {
		return types.Block{}, fmt.Errorf("block %v: %w", hash, errors.ErrNotFound)
	}

	return block, nil
}

// Get returns a block for given hash.
func (k *BaseBlockKeeper) Get(hash string) (types.Block, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.get(hash)
}

// IsReorg returns true if the chain was re-orged when add new block to it.
func (k *BaseBlockKeeper) IsReorg(block types.Block) (bool, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return block.ParentHash != k.head, nil
}

// GetRecentBlocks returns a list of recent blocks in descending order.
func (k *BaseBlockKeeper) GetRecentBlocks(n int) ([]types.Block, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if n <= 0 {
		return nil, fmt.Errorf("%w: n must be positive", errors.ErrInvalidArgument)
	}

	if n > len(k.blockMap) {
		n = len(k.blockMap)
	}

	blocks := make([]types.Block, 0, n)
	hash := k.head
	for i := 0; i < n; i++ {
		block, ok := k.blockMap[hash]
		if !ok {
			break
		}

		blocks = append(blocks, block)
		hash = block.ParentHash
	}

	return blocks, nil
}
