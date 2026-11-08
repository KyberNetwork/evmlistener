package evmclient

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/KyberNetwork/evmlistener/pkg/common"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// decodeLogs decodes an eth_getLogs result. Some nodes (e.g. Conduit on elysium) return
// `"topics": null` for topic-less logs, which types.Log rejects. On error, retry per element
// and drop only logs whose topics are absent/null; any other failure is returned unchanged.
func decodeLogs(raw json.RawMessage) ([]ethtypes.Log, error) {
	var logs []ethtypes.Log
	err := json.Unmarshal(raw, &logs)
	if err == nil {
		return logs, nil
	}

	var elems []json.RawMessage
	if json.Unmarshal(raw, &elems) != nil {
		return nil, err
	}

	logs = make([]ethtypes.Log, 0, len(elems))
	for _, elem := range elems {
		var l ethtypes.Log
		if json.Unmarshal(elem, &l) == nil {
			logs = append(logs, l)
			continue
		}

		var probe struct {
			Topics json.RawMessage `json:"topics"`
		}
		if json.Unmarshal(elem, &probe) != nil ||
			(len(probe.Topics) != 0 && !bytes.Equal(probe.Topics, []byte("null"))) {
			return nil, err
		}
	}

	return logs, nil
}

func filterLogs(ctx context.Context, client *rpc.Client, q FilterQuery) ([]types.Log, error) {
	fq := toEthereumFilterQuery(q)
	arg := map[string]any{}
	if len(fq.Addresses) > 0 {
		arg["address"] = fq.Addresses
	}
	if len(fq.Topics) > 0 {
		arg["topics"] = fq.Topics
	}
	if fq.BlockHash != nil {
		arg["blockHash"] = fq.BlockHash
	} else {
		arg["fromBlock"], arg["toBlock"] = "latest", "latest"
		if fq.FromBlock != nil {
			arg["fromBlock"] = hexutil.EncodeBig(fq.FromBlock)
		}
		if fq.ToBlock != nil {
			arg["toBlock"] = hexutil.EncodeBig(fq.ToBlock)
		}
	}

	var raw json.RawMessage
	if err := client.CallContext(ctx, &raw, "eth_getLogs", arg); err != nil {
		return nil, err
	}

	logs, err := decodeLogs(raw)
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
