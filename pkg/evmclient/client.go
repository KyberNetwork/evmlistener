package evmclient

import (
	"context"
	"errors"
	"math/big"
	"net/http"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/common"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient/avax"
	commonclient "github.com/KyberNetwork/evmlistener/pkg/evmclient/common"
	"github.com/KyberNetwork/evmlistener/pkg/types"
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
	customClient *commonclient.Client
	avaxClient   *avax.Client
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
	case chainIDFantom, chainIDZKSync:
		client.customClient = commonclient.NewClient(rpcClient)
	case chainIDAvalanche:
		client.avaxClient = avax.NewClient(rpcClient)
	default:
		client.ethClient = ethClient
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}

func DialContextWithTimeout(
	ctx context.Context,
	rawurl string,
	httpClient *http.Client,
	timeout time.Duration,
) (*Client, error) {
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
	case chainIDFantom, chainIDZKSync:
		return c.customClient.BlockNumber(ctx)
	case chainIDAvalanche:
		return c.avaxClient.BlockNumber(ctx)
	default:
		return c.ethClient.BlockNumber(ctx)
	}
}

//nolint:cyclop,ireturn,gocognit
func (c *Client) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (Subscription, error) {
	switch c.chainID {
	case chainIDFantom, chainIDAvalanche, chainIDZKSync:
		var (
			err      error
			sub      Subscription
			headerCh = make(chan *commonclient.Header)
		)

		if c.chainID == chainIDAvalanche {
			sub, err = c.avaxClient.SubscribeNewHead(ctx, headerCh)
		} else {
			sub, err = c.customClient.SubscribeNewHead(ctx, headerCh)
		}
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

func (c *Client) FilterLogs(ctx context.Context, q FilterQuery) ([]types.Log, error) {
	switch c.chainID {
	case chainIDFantom, chainIDZKSync:
		return filterLogs(ctx, c.customClient, q)
	case chainIDAvalanche:
		return filterLogs(ctx, c.avaxClient, q)
	default:
		return filterLogs(ctx, c.ethClient, q)
	}
}

func (c *Client) HeaderByHash(ctx context.Context, hash string) (*types.Header, error) {
	switch c.chainID {
	case chainIDFantom, chainIDAvalanche, chainIDZKSync:
		var (
			err    error
			header *commonclient.Header
		)
		if c.chainID == chainIDAvalanche {
			header, err = c.avaxClient.HeaderByHash(ctx, ethcommon.HexToHash(hash))
		} else {
			header, err = c.customClient.HeaderByHash(ctx, ethcommon.HexToHash(hash))
		}
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
	case chainIDFantom, chainIDAvalanche, chainIDZKSync:
		var (
			err    error
			header *commonclient.Header
		)
		if c.chainID == chainIDAvalanche {
			header, err = c.avaxClient.HeaderByNumber(ctx, number)
		} else {
			header, err = c.customClient.HeaderByNumber(ctx, number)
		}
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
