package listener

import (
	"context"
	"math/big"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/fantom-foundation/go-ethereum"
	"go.uber.org/zap"
)

const (
	errStringUnknownBlock = "unknown block"

	defaultRetryInterval = 1000 * time.Millisecond
)

// getLogsByBlockHash returns logs by block hash, retry up to 3 times.
func getLogsByBlockHash(
	ctx context.Context, evmClient EVMClient, hash string,
) (logs []types.Log, err error) {
	l := zap.S().With("hash", hash)

	for i := 0; i < 3; i++ {
		logs, err = evmClient.FilterLogs(ctx, ethereum.FilterQuery{BlockHash: &hash})
		if err == nil {
			return logs, nil
		}

		if err.Error() != errStringUnknownBlock {
			l.Errorw("Fail to get logs by block hash", "error", err)

			return nil, err
		}

		time.Sleep(defaultRetryInterval)
		zap.S().Infow("Retry get logs")
	}

	l.Errorw("Fail to get logs by block hash", "error", err)

	return nil, err
}

func getBlocks(
	ctx context.Context, evmClient EVMClient, fromBlock uint64, toBlock uint64,
) ([]types.Block, error) {
	// Get latest block by number.
	b, err := getBlockByNumber(ctx, evmClient, new(big.Int).SetUint64(toBlock))
	if err != nil {
		return nil, err
	}

	// Get block headers and its logs.
	n := int(toBlock - fromBlock + 1)
	blocks := make([]types.Block, n)
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
	ctx context.Context, evmClient EVMClient, hash string,
) (header *types.Header, err error) {
	for i := 0; i < 3; i++ {
		header, err = evmClient.HeaderByHash(ctx, hash)
		if err == nil {
			return header, nil
		}

		if err.Error() != errStringUnknownBlock {
			return nil, err
		}

		time.Sleep(defaultRetryInterval)
		zap.S().Infow("Retry get header by hash")
	}

	return nil, err
}

func getBlockByHash(
	ctx context.Context, evmClient EVMClient, hash string,
) (types.Block, error) {
	header, err := getHeaderByHash(ctx, evmClient, hash)
	if err != nil {
		return types.Block{}, err
	}

	logs, err := getLogsByBlockHash(ctx, evmClient, hash)
	if err != nil {
		return types.Block{}, err
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

		time.Sleep(defaultRetryInterval)
		zap.S().Infow("Retry get header by number")
	}

	return nil, err
}

func getBlockByNumber(
	ctx context.Context, evmClient EVMClient, num *big.Int,
) (types.Block, error) {
	header, err := getHeaderByNumber(ctx, evmClient, num)
	if err != nil {
		return types.Block{}, err
	}

	zap.S().Infow("Block information", "header", header)

	logs, err := getLogsByBlockHash(ctx, evmClient, header.Hash())
	if err != nil {
		return types.Block{}, err
	}

	return headerToBlock(header, logs), nil
}

func headerToBlock(header *types.Header, logs []Log) types.Block {
	return types.Block{
		Number:     header.Number,
		Hash:       header.Hash,
		Timestamp:  header.Time,
		ParentHash: header.ParentHash,
		Logs:       logs,
	}
}
