package listener

import (
	"context"
	"math/big"

	ltypes "github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const errStringUnknownBlock = "unknown block"

// getLogsByBlockHash returns logs by block hash, retry up to 3 times.
func getLogsByBlockHash(
	ctx context.Context, evmClient EVMClient, hash common.Hash,
) (logs []types.Log, err error) {
	for i := 0; i < 3; i++ {
		logs, err = evmClient.FilterLogs(ctx, ethereum.FilterQuery{BlockHash: &hash})
		if err == nil {
			return logs, nil
		}

		if err.Error() != errStringUnknownBlock {
			return nil, err
		}
	}

	return nil, err
}

func getBlocks(
	ctx context.Context, evmClient EVMClient, fromBlock uint64, toBlock uint64,
) ([]ltypes.Block, error) {
	// Get latest block by number.
	b, err := getBlockByNumber(ctx, evmClient, new(big.Int).SetUint64(toBlock))
	if err != nil {
		return nil, err
	}

	// Get block headers and its logs.
	n := int(toBlock - fromBlock + 1)
	blocks := make([]ltypes.Block, n)
	blocks[n-1] = b

	hash := b.ParentHash
	for i := n - 2; i >= 0; i-- { //nolint:gomnd
		b, err = getBlockByHash(ctx, evmClient, hash)
		if err != nil {
			return nil, err
		}
		blocks[i] = b
		hash = b.ParentHash
	}

	return blocks, nil
}

func getHeaderByHash(
	ctx context.Context, evmClient EVMClient, hash common.Hash,
) (header *types.Header, err error) {
	for i := 0; i < 3; i++ {
		header, err = evmClient.HeaderByHash(ctx, hash)
		if err == nil {
			return header, nil
		}

		if err.Error() != errStringUnknownBlock {
			return nil, err
		}
	}

	return nil, err
}

func getBlockByHash(
	ctx context.Context, evmClient EVMClient, hash common.Hash,
) (ltypes.Block, error) {
	header, err := getHeaderByHash(ctx, evmClient, hash)
	if err != nil {
		return ltypes.Block{}, err
	}

	logs, err := getLogsByBlockHash(ctx, evmClient, hash)
	if err != nil {
		return ltypes.Block{}, err
	}

	return headerToBlock(header, logs), nil
}

func getHeaderByNumber(
	ctx context.Context, evmClient EVMClient, num *big.Int,
) (header *types.Header, err error) {
	for i := 0; i < 3; i++ {
		header, err = evmClient.HeaderByNumber(ctx, num)
		if err == nil {
			return header, nil
		}

		if err.Error() != errStringUnknownBlock {
			return nil, err
		}
	}

	return nil, err
}

func getBlockByNumber(
	ctx context.Context, evmClient EVMClient, num *big.Int,
) (ltypes.Block, error) {
	header, err := getHeaderByNumber(ctx, evmClient, num)
	if err != nil {
		return ltypes.Block{}, err
	}

	logs, err := getLogsByBlockHash(ctx, evmClient, header.Hash())
	if err != nil {
		return ltypes.Block{}, err
	}

	return headerToBlock(header, logs), nil
}

func headerToBlock(header *types.Header, logs []types.Log) ltypes.Block {
	return ltypes.Block{
		Number:     header.Number,
		Hash:       header.Hash(),
		Timestamp:  header.Time,
		ParentHash: header.ParentHash,
		Logs:       logs,
	}
}
