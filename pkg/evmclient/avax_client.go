package evmclient

import (
	"context"
	"net/http"

	"github.com/ava-labs/coreth/ethclient"
	"github.com/ava-labs/coreth/rpc"
)

func AvaxDialContext(ctx context.Context, rawurl string) (ethclient.Client, error) {
	httpClient := rpc.WithHTTPClient(&http.Client{
		Timeout: defaultRequestTimeout,
	})

	rpcClient, err := rpc.DialOptions(ctx, rawurl, httpClient)
	if err != nil {
		return nil, err
	}

	return ethclient.NewClient(rpcClient), nil
}
