package evmclient

import (
	"context"

	"github.com/KyberNetwork/evmlistener/pkg/common"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func filterLogs(ctx context.Context, client ethereum.LogFilterer, q FilterQuery) ([]types.Log, error) {
	logs, err := client.FilterLogs(ctx, toEthereumFilterQuery(q))
	if err != nil {
		return nil, err
	}

	return fromEthereumLogs(logs), nil
}

func toEthereumFilterQuery(q FilterQuery) ethereum.FilterQuery {
	var blockHash *ethcommon.Hash
	if q.BlockHash != nil {
		hash := ethcommon.HexToHash(*q.BlockHash)
		blockHash = &hash
	}

	addresses := make([]ethcommon.Address, 0, len(q.Addresses))
	for _, address := range q.Addresses {
		addresses = append(addresses, ethcommon.HexToAddress(address))
	}

	var topics [][]ethcommon.Hash
	if len(q.Topics) > 0 {
		topics = make([][]ethcommon.Hash, 0, len(q.Topics))
		for _, ts := range q.Topics {
			tps := make([]ethcommon.Hash, 0, len(ts))
			for _, t := range ts {
				tps = append(tps, ethcommon.HexToHash(t))
			}
			topics = append(topics, tps)
		}
	}

	return ethereum.FilterQuery{
		BlockHash: blockHash,
		FromBlock: q.FromBlock,
		ToBlock:   q.ToBlock,
		Addresses: addresses,
		Topics:    topics,
	}
}

func fromEthereumLogs(logs []ethtypes.Log) []types.Log {
	res := make([]types.Log, 0, len(logs))
	for _, log := range logs {
		topics := make([]string, 0, len(log.Topics))
		for _, topic := range log.Topics {
			topics = append(topics, common.ToHex(topic))
		}

		res = append(res, types.Log{
			Address:     common.ToHex(log.Address),
			Topics:      topics,
			Data:        log.Data,
			BlockNumber: log.BlockNumber,
			TxHash:      common.ToHex(log.TxHash),
			TxIndex:     log.TxIndex,
			BlockHash:   common.ToHex(log.BlockHash),
			Index:       log.Index,
			Removed:     log.Removed,
		})
	}

	return res
}
