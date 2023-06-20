package listener

import (
	"context"
	"math/big"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum"
)

const (
	errStringUnknownBlock = "unknown block"

	defaultRetryInterval = 500 * time.Millisecond
)

// getLogsByBlockHash returns logs by block hash, retry up to 3 times.
func getLogsByBlockHash(
	ctx context.Context, evmClient evmclient.IClient, hash string,
) (logs []types.Log, err error) {
	for i := 0; i < 3; i++ {
		logs, err = evmClient.FilterLogs(ctx, evmclient.FilterQuery{BlockHash: &hash})
		if err == nil {
			return logs, nil
		}

		if errors.Is(err, ethereum.NotFound) && err.Error() != errStringUnknownBlock {
			return nil, err
		}

		time.Sleep(defaultRetryInterval)
	}

	return nil, err
}

func getBlocks(
	ctx context.Context, evmClient evmclient.IClient, fromBlock uint64, toBlock uint64,
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
	ctx context.Context, evmClient evmclient.IClient, hash string,
) (header *types.Header, err error) {
	for i := 0; i < 3; i++ {
		header, err = evmClient.HeaderByHash(ctx, hash)
		if err == nil {
			return header, nil
		}

		if errors.Is(err, ethereum.NotFound) && err.Error() != errStringUnknownBlock {
			return nil, err
		}

		time.Sleep(defaultRetryInterval)
	}

	return nil, err
}

func getBlockByHash(
	ctx context.Context, evmClient evmclient.IClient, hash string,
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
	ctx context.Context, evmClient evmclient.IClient, num *big.Int,
) (header *types.Header, err error) {
	for i := 0; i < 3; i++ {
		header, err = evmClient.HeaderByNumber(ctx, num)
		if err == nil {
			return header, nil
		}

		if errors.Is(err, ethereum.NotFound) && err.Error() != errStringUnknownBlock {
			return nil, err
		}

		time.Sleep(defaultRetryInterval)
	}

	return nil, err
}

func getBlockByNumber(
	ctx context.Context, evmClient evmclient.IClient, num *big.Int,
) (types.Block, error) {
	header, err := getHeaderByNumber(ctx, evmClient, num)
	if err != nil {
		return types.Block{}, err
	}

	logs, err := getLogsByBlockHash(ctx, evmClient, header.Hash)
	if err != nil {
		return types.Block{}, err
	}

	return headerToBlock(header, logs), nil
}

func headerToBlock(header *types.Header, logs []types.Log) types.Block {
	return types.Block{
		Hash:       header.Hash,
		Number:     header.Number,
		Timestamp:  header.Time,
		ParentHash: header.ParentHash,
		Logs:       logs,
	}
}
