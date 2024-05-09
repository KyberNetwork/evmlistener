package evmclient

import (
	"context"
	"errors"
	"math/big"
	"net/http"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/common"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient/ftmclient"
	zksyncclient "github.com/KyberNetwork/evmlistener/pkg/evmclient/zksync-client"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	avaxtypes "github.com/ava-labs/coreth/core/types"
	aethclient "github.com/ava-labs/coreth/ethclient"
	"github.com/ava-labs/coreth/interfaces"
	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	chainIDFantom    = 250
	chainIDAvalanche = 43114
	chainIDZKSync    = 324
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
	chainID      uint64
	ethClient    *ethclient.Client
	ftmClient    *ftmclient.Client
	avaxClient   aethclient.Client
	zksyncClient *zksyncclient.Client
}

func Dial(rawurl string, httpClient *http.Client) (*Client, error) {
	return DialContext(context.Background(), rawurl, httpClient)
}

func DialContext(ctx context.Context, rawurl string, httpClient *http.Client) (*Client, error) {
	rpcClient, err := rpc.DialOptions(ctx, rawurl, rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	ethClient := ethclient.NewClient(rpcClient)
	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	client := &Client{
		chainID: chainID.Uint64(),
	}

	switch client.chainID {
	case chainIDFantom:
		client.ftmClient = ftmclient.NewClient(rpcClient)
	case chainIDAvalanche:
		client.avaxClient, err = AvaxDialContext(ctx, rawurl, httpClient)
	case chainIDZKSync:
		client.zksyncClient = zksyncclient.NewClient(rpcClient)
	default:
		client.ethClient = ethClient
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}

func DialContextWithTimeout(ctx context.Context, rawurl string, httpClient *http.Client, timeout time.Duration) (*Client, error) {
	type dialContextResponse struct {
		client *Client
		err    error
	}

	ch := make(chan dialContextResponse, 1)
	go func() {
		client, err := DialContext(ctx, rawurl, httpClient)
		ch <- dialContextResponse{
			client: client,
			err:    err,
		}
	}()

	select {
	case res := <-ch:
		return res.client, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, errors.New("timeout when dial RPC")
	}
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
	case chainIDZKSync:
		return c.zksyncClient.BlockNumber(ctx)
	default:
		return c.ethClient.BlockNumber(ctx)
	}
}

//nolint:cyclop,ireturn,funlen,gocognit
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
	case chainIDZKSync:
		headerCh := make(chan *zksyncclient.Header)
		sub, err := c.zksyncClient.SubscribeNewHead(ctx, headerCh)
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
	logs, err := c.ftmClient.FilterLogs(ctx, toEthereumFilterQuery(q))
	if err != nil {
		return nil, err
	}

	return fromEthereumLogs(logs), nil
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

	return fromAvalancheLogs(logs), nil
}

//nolint:dupl
func (c *Client) ethFilterLogs(ctx context.Context, q FilterQuery) ([]types.Log, error) {
	logs, err := c.ethClient.FilterLogs(ctx, toEthereumFilterQuery(q))
	if err != nil {
		return nil, err
	}

	return fromEthereumLogs(logs), nil
}

//nolint:dupl
func (c *Client) zksyncFilterLogs(ctx context.Context, q FilterQuery) ([]types.Log, error) {
	logs, err := c.zksyncClient.FilterLogs(ctx, toEthereumFilterQuery(q))
	if err != nil {
		return nil, err
	}

	return fromEthereumLogs(logs), nil
}

func (c *Client) FilterLogs(ctx context.Context, q FilterQuery) ([]types.Log, error) {
	switch c.chainID {
	case chainIDFantom:
		return c.ftmFilterLogs(ctx, q)
	case chainIDAvalanche:
		return c.avaxFilterLogs(ctx, q)
	case chainIDZKSync:
		return c.zksyncFilterLogs(ctx, q)
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
	case chainIDZKSync:
		header, err := c.zksyncClient.HeaderByHash(ctx, ethcommon.HexToHash(hash))
		if err != nil {
			return nil, err
		}

		return &types.Header{
			Hash:       common.ToHex(header.Hash),
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
	case chainIDZKSync:
		header, err := c.zksyncClient.HeaderByNumber(ctx, number)
		if err != nil {
			return nil, err
		}

		return &types.Header{
			Hash:       common.ToHex(header.Hash),
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

func fromAvalancheLogs(logs []avaxtypes.Log) []types.Log {
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
