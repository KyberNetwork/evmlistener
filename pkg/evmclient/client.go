package evmclient

import (
	"context"
	"math/big"

	"github.com/KyberNetwork/evmlistener/pkg/types"
	aethclient "github.com/ava-labs/coreth/ethclient"
	"github.com/ethereum/go-ethereum/ethclient"
	fethclient "github.com/fantom-foundation/go-ethereum/ethclient"
)

const (
	chainIDFantom   = 250
	chainIDAvalanch = 43114
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
	ftmClient  *fethclient.Client
	avaxClient *aethclient.Client
}

func Dial(rawurl string) (*Client, error) {
	ethClient, err := ethclient.Dial(rawurl)
	if err != nil {
		return nil, err
	}

	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	client := &Client{
		chainID: chainID.Uint64(),
	}

	switch client.chainID {
	case chainIDFantom:
		client.ftmClient, err = fethclient.Dial(rawurl)
	case chainIDAvalanche:
		client.avaxClient, err = aethclient.Dial(rawurl)
	default:
		client.ethClient = ethClient
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	return c.chainID, nil
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
