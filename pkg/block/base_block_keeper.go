package block

import (
	"fmt"
	"sync"

	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/emirpasic/gods/queues/circularbuffer"
	"github.com/ethereum/go-ethereum/common"
)

// BaseBlockKeeper is a purely on-memory block keeper.
type BaseBlockKeeper struct {
	mu           sync.RWMutex
	maxNumBlocks int
	head         common.Hash
	blockMap     map[common.Hash]Block
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
		head:         common.Hash{},
		blockMap:     make(map[common.Hash]Block, maxNumBlocks),
		queue:        circularbuffer.New(maxNumBlocks),
	}
}

// Init ...
func (k *BaseBlockKeeper) Init() error {
	if len(k.blockMap) > 0 {
		k.blockMap = make(map[common.Hash]Block, k.maxNumBlocks)
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

func (k *BaseBlockKeeper) exists(hash common.Hash) bool {
	_, ok := k.blockMap[hash]

	return ok
}

// Exists checks whether a block hash is exists or not.
func (k *BaseBlockKeeper) Exists(hash common.Hash) (bool, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.exists(hash), nil
}

func (k *BaseBlockKeeper) removeOld() {
	v, ok := k.queue.Dequeue()
	if !ok {
		return
	}

	hash, ok := v.(common.Hash)
	if !ok {
		return
	}

	delete(k.blockMap, hash)
}

// Add adds new block to the keeper.
func (k *BaseBlockKeeper) Add(block Block) error {
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
			block.ReorgedHash = common.Hash{}
		}
	}

	k.blockMap[block.Hash] = block
	k.head = block.Hash
	k.queue.Enqueue(block.Hash)

	return nil
}

// Head returns the block head of the chain on the keeper.
func (k *BaseBlockKeeper) Head() (Block, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.Get(k.head)
}

func (k *BaseBlockKeeper) get(hash common.Hash) (Block, error) {
	block, ok := k.blockMap[hash]
	if !ok {
		return Block{}, fmt.Errorf("block %v: %w", hash, errors.ErrNotFound)
	}

	return block, nil
}

// Get returns a block for given hash.
func (k *BaseBlockKeeper) Get(hash common.Hash) (Block, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.get(hash)
}

// IsReorg returns true if the chain was re-orged when add new block to it.
func (k *BaseBlockKeeper) IsReorg(block Block) (bool, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return block.ParentHash != k.head, nil
}

// GetRecentBlocks returns a list of recent blocks in descending order.
func (k *BaseBlockKeeper) GetRecentBlocks(n int) ([]Block, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if n <= 0 {
		return nil, fmt.Errorf("%w: n must be positive", errors.ErrInvalidArgument)
	}

	if n > len(k.blockMap) {
		n = len(k.blockMap)
	}

	blocks := make([]Block, 0, n)
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
