package blockpoller

import (
	"context"
	"math/big"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/evmclient"
	"github.com/KyberNetwork/evmlistener/pkg/types"
)

type pollSubscription struct {
	errCh  chan error
	stopCh chan struct{}
}

func (s *pollSubscription) Err() <-chan error { return s.errCh }
func (s *pollSubscription) Unsubscribe()      { close(s.stopCh) }

// BlockPoller implements evmclient.IClient for chains without websocket support
// Only SubscribeNewHead is implemented for polling new heads; other methods delegate to the underlying httpClient.
type BlockPoller struct {
	evmclient.IClient
	interval time.Duration
}

func New(httpClient evmclient.IClient, interval time.Duration) *BlockPoller {
	return &BlockPoller{IClient: httpClient, interval: interval}
}

func (b *BlockPoller) SubscribeNewHead(ctx context.Context, headerCh chan<- *types.Header) (evmclient.Subscription,
	error) {
	sub := &pollSubscription{
		errCh:  make(chan error, 1),
		stopCh: make(chan struct{}),
	}

	go func() {
		var lastNumber *big.Int
		ticker := time.NewTicker(b.interval)
		defer ticker.Stop()

		for {
			if err := b.fetchNewBlocks(ctx, headerCh, &lastNumber); err != nil {
				sub.errCh <- err
				return
			}

			select {
			case <-ctx.Done():
				sub.errCh <- ctx.Err()
				return
			case <-sub.stopCh:
				close(sub.errCh)
				return
			case <-ticker.C:
			}
		}
	}()

	return sub, nil
}

var biOne = big.NewInt(1)

// fetchNewBlocks fetches all new blocks since lastNumber and updates lastNumber.
func (b *BlockPoller) fetchNewBlocks(ctx context.Context, headerCh chan<- *types.Header, lastNumber **big.Int) error {
	// Get latest block
	header, err := b.IClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return err
	}

	// If first run, just return the latest block
	if *lastNumber == nil {
		*lastNumber = new(big.Int).Set(header.Number)
		headerCh <- header
		return nil
	}

	// If no new blocks, return early
	if (*lastNumber).Cmp(header.Number) >= 0 {
		return nil
	}

	// Fetch all missing blocks in sequence
	for (*lastNumber).Add(*lastNumber, biOne).Cmp(header.Number) < 0 {
		missedHeader, err := b.IClient.HeaderByNumber(ctx, *lastNumber)
		if err != nil {
			return err
		}
		headerCh <- missedHeader
	}
	headerCh <- header

	return nil
}
