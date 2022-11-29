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

func getBlocks(
	ctx context.Context, evmClient EVMClient, fromBlock uint64, toBlock uint64,
) ([]ltypes.Block, error) {
	// Get events for blocks.
	logs, err := getLogs(ctx, evmClient, fromBlock, toBlock)
	if err != nil {
		return nil, err
	}

	if len(logs) > 1 {
		sort.Slice(logs, func(i, j int) bool {
			if logs[i].BlockNumber == logs[j].BlockNumber {
				return logs[i].Index < logs[j].Index
			}

			return logs[i].BlockNumber < logs[j].BlockNumber
		})
	}

	blockLogsMap := make(map[common.Hash][]types.Log)
	for _, log := range logs {
		blockLogs := blockLogsMap[log.BlockHash]
		blockLogs = append(blockLogs, log)
		blockLogsMap[log.BlockHash] = blockLogs
	}

	// Get block headers.
	blocks := make([]ltypes.Block, 0, int(toBlock-fromBlock+1))
	for i := fromBlock; i <= toBlock; i++ {
		header, err := evmClient.HeaderByNumber(ctx, new(big.Int).SetUint64(i))
		if err != nil {
			return nil, err
		}

		hash := header.Hash()
		blocks = append(blocks, ltypes.Block{
			Number:     header.Number,
			Hash:       hash,
			ParentHash: header.ParentHash,
			Logs:       blockLogsMap[hash],
		})
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

	logs, err := evmClient.FilterLogs(ctx, ethereum.FilterQuery{BlockHash: &hash})
	if err != nil {
		return ltypes.Block{}, err
	}

	return ltypes.Block{
		Number:     header.Number,
		Hash:       hash,
		ParentHash: header.ParentHash,
		Logs:       logs,
	}, nil
}
