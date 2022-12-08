package listener

import (
	"context"
	"math/big"

	ltypes "github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func getLogsByBlockHash(
	ctx context.Context, evmClient EVMClient, hash common.Hash,
) ([]types.Log, error) {
	return evmClient.FilterLogs(ctx, ethereum.FilterQuery{
		BlockHash: &hash,
	})
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

func getBlockByHash(
	ctx context.Context, evmClient EVMClient, hash common.Hash,
) (ltypes.Block, error) {
	header, err := evmClient.HeaderByHash(ctx, hash)
	if err != nil {
		return ltypes.Block{}, err
	}

	logs, err := getLogsByBlockHash(ctx, evmClient, hash)
	if err != nil {
		return ltypes.Block{}, err
	}

	return headerToBlock(header, logs), nil
}

func getBlockByNumber(
	ctx context.Context, evmClient EVMClient, num *big.Int,
) (ltypes.Block, error) {
	header, err := evmClient.HeaderByNumber(ctx, num)
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
		ParentHash: header.ParentHash,
		Logs:       logs,
	}
}
