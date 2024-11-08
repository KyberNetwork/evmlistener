package avax

import (
	"context"

	"github.com/KyberNetwork/evmlistener/pkg/evmclient/common"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// Client extends Ethereum API client with typed wrappers for the FTM API.
type Client struct {
	*common.Client
	c *rpc.Client
}

func NewClient(c *rpc.Client) *Client {
	return &Client{
		Client: common.NewClient(c),
		c:      c,
	}
}

func (c *Client) SubscribeNewHead(
	ctx context.Context, ch chan<- *common.Header,
) (ethereum.Subscription, error) {
	// Subscribe headers to this intermediate channel, calculate hash and forward for original channel.
	ch2 := make(chan *Header, 1000)

	sub, err := c.c.EthSubscribe(ctx, ch2, "newHeads")
	if err != nil {
		return nil, err
	}

	go func() {
		select {
		case <-ctx.Done():
			return
		case h := <-ch2:
			ch <- &common.Header{
				Hash: h.Hash(),
				Header: types.Header{
					ParentHash:  h.ParentHash,
					UncleHash:   h.UncleHash,
					Coinbase:    h.Coinbase,
					Root:        h.Root,
					TxHash:      h.TxHash,
					ReceiptHash: h.ReceiptHash,
					Bloom:       h.Bloom,
					Difficulty:  h.Difficulty,
					Number:      h.Number,
					GasLimit:    h.GasLimit,
					GasUsed:     h.GasUsed,
					Time:        h.Time,
					Extra:       h.Extra,
					MixDigest:   h.MixDigest,
					Nonce:       h.Nonce,
					BaseFee:     h.BaseFee,
				},
			}
		}
	}()

	return sub, nil
}
