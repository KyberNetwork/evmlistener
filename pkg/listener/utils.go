package listener

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"

	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient"
	"github.com/KyberNetwork/evmlistener/pkg/types"
)

const (
	errStringUnknownBlock   = "unknown block"
	errStringResponseTooBig = "Response is too big"

	defaultRetryInterval = 500 * time.Millisecond
)

// filterSpamLogs drops all but the last log for any spam key that appears more than threshold
// times in the slice. The key is (address, topic0, topic1) for events with topics, or (address)
// alone for anonymous events (no topics). This lets through Uniswap V4 / Balancer vault events
// where topic1 carries the pool ID and differs per pool, while catching mass-transfer airdrops
// (same from-address in topic1) and anonymous event spam. Order of surviving logs is preserved.
func filterSpamLogs(logs []types.Log, threshold int) []types.Log {
	type key struct{ addr, topic0, topic1 string }

	keyOf := func(l types.Log) key {
		k := key{addr: l.Address}
		if len(l.Topics) > 0 {
			k.topic0 = l.Topics[0]
		}
		if len(l.Topics) > 1 {
			k.topic1 = l.Topics[1]
		}
		return k
	}

	if len(logs) <= threshold {
		return logs
	}

	counts := make(map[key]int)
	for _, l := range logs {
		counts[keyOf(l)]++
	}

	// For each over-threshold key, record the index of its last occurrence.
	lastIdx := make(map[key]int)
	for i, l := range logs {
		k := keyOf(l)
		if counts[k] > threshold {
			lastIdx[k] = i
		}
	}

	if len(lastIdx) == 0 {
		return logs
	}

	filtered := make([]types.Log, 0, len(logs))
	for i, l := range logs {
		if last, isSpam := lastIdx[keyOf(l)]; isSpam && i != last {
			continue
		}
		filtered = append(filtered, l)
	}
	return filtered
}

// getLogsByBlockHash returns logs by block hash, retry up to 3 times.
func getLogsByBlockHash(ctx context.Context, evmClient evmclient.IClient, hash string,
	contracts []string, topics [][]string,
) (logs []types.Log, err error) {
	for range 3 {
		logs, err = evmClient.FilterLogs(ctx, evmclient.FilterQuery{
			BlockHash: &hash,
			Addresses: contracts,
			Topics:    topics,
		})
		if err == nil {
			if len(logs) == 0 {
				continue
			}

			return logs, nil
		} else if err.Error() == errStringResponseTooBig {
			return nil, nil
		} else if !errors.Is(err, ethereum.NotFound) && err.Error() != errStringUnknownBlock {
			return nil, err
		}

		time.Sleep(defaultRetryInterval)
	}

	return logs, err
}

func GetBlocks(ctx context.Context, evmClient evmclient.IClient, fromBlock uint64, toBlock uint64,
	withLogs bool, contracts []string, topics [][]string,
) ([]types.Block, error) {
	// Get latest block by number.
	b, err := getBlockByNumber(ctx, evmClient, new(big.Int).SetUint64(toBlock), withLogs, contracts, topics)
	if err != nil {
		return nil, err
	}

	// Get block headers and its logs.
	n := int(toBlock - fromBlock + 1)
	blocks := make([]types.Block, n)
	blocks[n-1] = b

	hash := b.ParentHash
	for i := n - 2; i >= 0; i-- {
		b, err = getBlockByHash(ctx, evmClient, hash, withLogs, contracts, topics)
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
	for range 5 {
		header, err = evmClient.HeaderByHash(ctx, hash)
		if err == nil {
			return header, nil
		}

		if !errors.Is(err, ethereum.NotFound) && err.Error() != errStringUnknownBlock {
			return nil, err
		}

		time.Sleep(defaultRetryInterval)
	}

	return nil, err
}

func getBlockByHash(ctx context.Context, evmClient evmclient.IClient, hash string, withLogs bool,
	contracts []string, topics [][]string,
) (types.Block, error) {
	header, err := getHeaderByHash(ctx, evmClient, hash)
	if err != nil {
		return types.Block{}, err
	}
	var logs []types.Log
	if withLogs {
		logs, err = getLogsByBlockHash(ctx, evmClient, hash, contracts, topics)
		if err != nil {
			return types.Block{}, err
		}
	}

	return headerToBlock(header, logs), nil
}

func getHeaderByNumber(
	ctx context.Context, evmClient evmclient.IClient, num *big.Int,
) (header *types.Header, err error) {
	for range 3 {
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

func getBlockByNumber(ctx context.Context, evmClient evmclient.IClient, num *big.Int,
	withLogs bool, contracts []string, topics [][]string,
) (types.Block, error) {
	header, err := getHeaderByNumber(ctx, evmClient, num)
	if err != nil {
		return types.Block{}, err
	}
	var logs []types.Log
	if withLogs {
		logs, err = getLogsByBlockHash(ctx, evmClient, header.Hash, contracts, topics)
		if err != nil {
			return types.Block{}, err
		}
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
