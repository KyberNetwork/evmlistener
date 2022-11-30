package listener

import (
	"context"
	"math/big"
	"sort"

	ltypes "github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func getLogs(
	ctx context.Context, evmClient EVMClient, fromBlock uint64, toBlock uint64,
) (logs []types.Log, err error) {
	const step = 32
	for i := fromBlock; i <= toBlock; i += step {
		e := i + step - 1
		if e > toBlock {
			e = toBlock
		}

		newLogs, err := evmClient.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: new(big.Int).SetUint64(i),
			ToBlock:   new(big.Int).SetUint64(e),
		})
		if err != nil {
			return nil, err
		}

		logs = append(logs, newLogs...)
	}

	return logs, nil
}

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
	// Get events for blocks.
	logs, err := getLogs(ctx, evmClient, fromBlock, toBlock)
	if err != nil {
		return nil, err
	}

	blockLogsMap := make(map[common.Hash][]types.Log)
	for _, log := range logs {
		blockLogs := blockLogsMap[log.BlockHash]
		blockLogs = append(blockLogs, log)
		blockLogsMap[log.BlockHash] = blockLogs
	}

	// Get block headers.
	blocks := make([]ltypes.Block, 0, int(toBlock-fromBlock+1))
	for hash, logs := range blockLogsMap {
		header, err := evmClient.HeaderByHash(ctx, hash)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, headerToBlock(header, logs))
	}

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Number.Cmp(blocks[j].Number) < 0
	})

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

func headerToBlock(header *types.Header, logs []types.Log) ltypes.Block {
	return ltypes.Block{
		Number:     header.Number,
		Hash:       header.Hash(),
		ParentHash: header.ParentHash,
		Logs:       logs,
	}
}
