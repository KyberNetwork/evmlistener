package evmclient

import (
	"context"
	"net/http"

	"github.com/ava-labs/coreth/ethclient"
	"github.com/ava-labs/coreth/rpc"
)

func AvaxDialContext(ctx context.Context, rawurl string, httpClient *http.Client) (ethclient.Client, error) {
	rpcClient, err := rpc.DialOptions(ctx, rawurl, rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	return ethclient.NewClient(rpcClient), nil
}
