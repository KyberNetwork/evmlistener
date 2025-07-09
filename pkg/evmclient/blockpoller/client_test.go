package blockpoller

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/evmclient"
	"github.com/KyberNetwork/evmlistener/pkg/types"
)

func TestBlockPoller_SubscribeNewHead(t *testing.T) {
	t.Skip()
	evmclient.UseCustomClient = true
	ctx := context.Background()
	httpCli, _ := evmclient.DialContext(ctx, "https://rpc.hyperliquid.xyz/evm", &http.Client{
		Timeout: 5 * time.Second,
	})
	blockPoller := New(httpCli, 2500*time.Millisecond)
	headerCh := make(chan *types.Header, 5)
	sub, err := blockPoller.SubscribeNewHead(ctx, headerCh)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()
	defer close(headerCh)
	after := time.After(10 * time.Second)
	for header := range headerCh {
		fmt.Println(header)
		select {
		case <-after:
			return
		default:
		}
	}
}
