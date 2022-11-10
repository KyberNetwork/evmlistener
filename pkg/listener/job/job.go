package job

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Job represents a job for updating pool's state.
type Job struct {
	Topic       string
	BlockNumber *big.Int // Latest block number.
	PoolAddress common.Address
	Logs        []types.Log
}

// Splitter is an interface support for splitting jobs from logs.
type Splitter interface {
	Split(logs []types.Log) []Job
}

// Publisher is an interface for publishing jobs.
type Publisher interface {
	Publish(ctx context.Context, job Job) error
}
