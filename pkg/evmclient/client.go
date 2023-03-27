package evmclient

import (
	"context"
	"math/big"

	"github.com/KyberNetwork/evmlistener/pkg/common"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient/ftmclient"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	avaxtypes "github.com/ava-labs/coreth/core/types"
	avaxclient "github.com/ava-labs/coreth/ethclient"
	"github.com/ava-labs/coreth/interfaces"
	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	chainIDFantom    = 250
	chainIDAvalanche = 43114
)

type FilterQuery struct {
	BlockHash *string
	FromBlock *big.Int
	ToBlock   *big.Int
	Addresses []string
	Topics    [][]string
}

type Subscription interface {
	// Unsubscribe cancels the sending of events to the data channel
	// and closes the error channel.
	Unsubscribe()
	// Err returns the subscription error channel. The error channel receives
	// a value if there is an issue with the subscription (e.g. the network connection
	// delivering the events has been closed). Only one value will ever be sent.
	// The error channel is closed by Unsubscribe.
	Err() <-chan error
}

// IClient is an interface for EVM client.
type IClient interface {
	BlockNumber(context.Context) (uint64, error)
	SubscribeNewHead(context.Context, chan<- *types.Header) (Subscription, error)
	FilterLogs(context.Context, FilterQuery) ([]types.Log, error)
	HeaderByHash(context.Context, string) (*types.Header, error)
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
}

type Client struct {
	chainID    uint64
	ethClient  *ethclient.Client
	ftmClient  *ftmclient.Client
	avaxClient avaxclient.Client
}

func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	ethClient, err := ethclient.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	client := &Client{
		chainID: chainID.Uint64(),
	}

	switch client.chainID {
	case chainIDFantom:
		client.ftmClient, err = ftmclient.DialContext(ctx, rawurl)
	case chainIDAvalanche:
		client.avaxClient, err = avaxclient.DialContext(ctx, rawurl)
	default:
		client.ethClient = ethClient
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	return new(big.Int).SetUint64(c.chainID), nil
}

func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	switch c.chainID {
	case chainIDFantom:
		return c.ftmClient.BlockNumber(ctx)
	case chainIDAvalanche:
		return c.avaxClient.BlockNumber(ctx)
	default:
		return c.ethClient.BlockNumber(ctx)
	}
}

//nolint:cyclop,ireturn
func (c *Client) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (Subscription, error) {
	switch c.chainID {
	case chainIDFantom:
		headerCh := make(chan *ftmclient.Header)
		sub, err := c.ftmClient.SubscribeNewHead(ctx, headerCh)
		if err != nil {
			return nil, err
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case header := <-headerCh:
					ch <- &types.Header{
						Hash:       common.ToHex(header.Hash),
						ParentHash: common.ToHex(header.ParentHash),
						Number:     header.Number,
						Time:       header.Time,
					}
				}
			}
		}()

		return sub, nil
	case chainIDAvalanche:
		headerCh := make(chan *avaxtypes.Header)
		sub, err := c.avaxClient.SubscribeNewHead(ctx, headerCh)
		if err != nil {
			return nil, err
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case header := <-headerCh:
					ch <- &types.Header{
						Hash:       common.ToHex(header.Hash()),
						ParentHash: common.ToHex(header.ParentHash),
						Number:     header.Number,
						Time:       header.Time,
					}
				}
			}
		}()

		return sub, nil
	default:
		headerCh := make(chan *ethtypes.Header)
		sub, err := c.ethClient.SubscribeNewHead(ctx, headerCh)
		if err != nil {
			return nil, err
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case header := <-headerCh:
					ch <- &types.Header{
						Hash:       common.ToHex(header.Hash()),
						ParentHash: common.ToHex(header.ParentHash),
						Number:     header.Number,
						Time:       header.Time,
					}
				}
			}
		}()

		return sub, nil
	}
}

//nolint:dupl
func (c *Client) ftmFilterLogs(ctx context.Context, q FilterQuery) ([]types.Log, error) {
	var blockHash *ethcommon.Hash
	if q.BlockHash != nil {
		hash := ethcommon.HexToHash(*q.BlockHash)
		blockHash = &hash
	}

	addresses := make([]ethcommon.Address, 0, len(q.Addresses))
	for _, address := range q.Addresses {
		addresses = append(addresses, ethcommon.HexToAddress(address))
	}

	topics := make([][]ethcommon.Hash, 0, len(q.Topics))
	for i, ts := range q.Topics {
		topics[i] = make([]ethcommon.Hash, 0, len(ts))
		for _, t := range ts {
			topics[i] = append(topics[i], ethcommon.HexToHash(t))
		}
	}

	logs, err := c.ftmClient.FilterLogs(ctx, ethereum.FilterQuery{
		BlockHash: blockHash,
		FromBlock: q.FromBlock,
		ToBlock:   q.ToBlock,
		Addresses: addresses,
		Topics:    topics,
	})
	if err != nil {
		return nil, err
	}

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

	return res, nil
}

//nolint:dupl
func (c *Client) avaxFilterLogs(ctx context.Context, q FilterQuery) ([]types.Log, error) {
	var blockHash *ethcommon.Hash
	if q.BlockHash != nil {
		hash := ethcommon.HexToHash(*q.BlockHash)
		blockHash = &hash
	}

	addresses := make([]ethcommon.Address, 0, len(q.Addresses))
	for _, address := range q.Addresses {
		addresses = append(addresses, ethcommon.HexToAddress(address))
	}

	topics := make([][]ethcommon.Hash, 0, len(q.Topics))
	for i, ts := range q.Topics {
		topics[i] = make([]ethcommon.Hash, 0, len(ts))
		for _, t := range ts {
			topics[i] = append(topics[i], ethcommon.HexToHash(t))
		}
	}

	logs, err := c.avaxClient.FilterLogs(ctx, interfaces.FilterQuery{
		BlockHash: blockHash,
		FromBlock: q.FromBlock,
		ToBlock:   q.ToBlock,
		Addresses: addresses,
		Topics:    topics,
	})
	if err != nil {
		return nil, err
	}

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

	return res, nil
}

//nolint:dupl
func (c *Client) ethFilterLogs(ctx context.Context, q FilterQuery) ([]types.Log, error) {
	var blockHash *ethcommon.Hash
	if q.BlockHash != nil {
		hash := ethcommon.HexToHash(*q.BlockHash)
		blockHash = &hash
	}

	addresses := make([]ethcommon.Address, 0, len(q.Addresses))
	for _, address := range q.Addresses {
		addresses = append(addresses, ethcommon.HexToAddress(address))
	}

	topics := make([][]ethcommon.Hash, 0, len(q.Topics))
	for i, ts := range q.Topics {
		topics[i] = make([]ethcommon.Hash, 0, len(ts))
		for _, t := range ts {
			topics[i] = append(topics[i], ethcommon.HexToHash(t))
		}
	}

	logs, err := c.ethClient.FilterLogs(ctx, ethereum.FilterQuery{
		BlockHash: blockHash,
		FromBlock: q.FromBlock,
		ToBlock:   q.ToBlock,
		Addresses: addresses,
		Topics:    topics,
	})
	if err != nil {
		return nil, err
	}

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

	return res, nil
}

func (c *Client) FilterLogs(ctx context.Context, q FilterQuery) ([]types.Log, error) {
	switch c.chainID {
	case chainIDFantom:
		return c.ftmFilterLogs(ctx, q)
	case chainIDAvalanche:
		return c.avaxFilterLogs(ctx, q)
	default:
		return c.ethFilterLogs(ctx, q)
	}
}

func (c *Client) HeaderByHash(ctx context.Context, hash string) (*types.Header, error) {
	switch c.chainID {
	case chainIDFantom:
		header, err := c.ftmClient.HeaderByHash(ctx, ethcommon.HexToHash(hash))
		if err != nil {
			return nil, err
		}

		return &types.Header{
			Hash:       common.ToHex(header.Hash),
			ParentHash: common.ToHex(header.ParentHash),
			Number:     header.Number,
			Time:       header.Time,
		}, nil
	case chainIDAvalanche:
		header, err := c.avaxClient.HeaderByHash(ctx, ethcommon.HexToHash(hash))
		if err != nil {
			return nil, err
		}

		return &types.Header{
			Hash:       common.ToHex(header.Hash()),
			ParentHash: common.ToHex(header.ParentHash),
			Number:     header.Number,
			Time:       header.Time,
		}, nil
	default:
		header, err := c.ethClient.HeaderByHash(ctx, ethcommon.HexToHash(hash))
		if err != nil {
			return nil, err
		}

		return &types.Header{
			Hash:       common.ToHex(header.Hash()),
			ParentHash: common.ToHex(header.ParentHash),
			Number:     header.Number,
			Time:       header.Time,
		}, nil
	}
}

func (c *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	switch c.chainID {
	case chainIDFantom:
		header, err := c.ftmClient.HeaderByNumber(ctx, number)
		if err != nil {
			return nil, err
		}

		return &types.Header{
			Hash:       common.ToHex(header.Hash),
			ParentHash: common.ToHex(header.ParentHash),
			Number:     header.Number,
			Time:       header.Time,
		}, nil
	case chainIDAvalanche:
		header, err := c.avaxClient.HeaderByNumber(ctx, number)
		if err != nil {
			return nil, err
		}

		return &types.Header{
			Hash:       common.ToHex(header.Hash()),
			ParentHash: common.ToHex(header.ParentHash),
			Number:     header.Number,
			Time:       header.Time,
		}, nil
	default:
		header, err := c.ethClient.HeaderByNumber(ctx, number)
		if err != nil {
			return nil, err
		}

		return &types.Header{
			Hash:       common.ToHex(header.Hash()),
			ParentHash: common.ToHex(header.ParentHash),
			Number:     header.Number,
			Time:       header.Time,
		}, nil
	}
}
