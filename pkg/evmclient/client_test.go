package evmclient

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

const hash = "0x" + "11111111111111111111111111111111111111111111111111111111111111aa"

// Nodes like Conduit return `"topics": null` for LOG0; FilterLogs must still decode them.
func TestFilterLogsNullTopics(t *testing.T) {
	var gotParams json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			ID     json.RawMessage   `json:"id"`
			Params []json.RawMessage `json:"params"`
		}
		_ = json.Unmarshal(body, &req)
		gotParams = req.Params[0]
		resp := `{"jsonrpc":"2.0","id":` + string(req.ID) + `,"result":[` +
			`{"address":"0x524f17a90c20a6de53bc0b82899fbcfef6bb7a39","topics":null,"data":"0x01",` +
			`"blockNumber":"0x19d8da","transactionHash":"` + hash + `","transactionIndex":"0x0",` +
			`"blockHash":"` + hash + `","logIndex":"0x0","removed":false},` +
			`{"address":"0x524f17a90c20a6de53bc0b82899fbcfef6bb7a39","topics":["` + hash + `"],"data":"0x",` +
			`"blockNumber":"0x19d8da","transactionHash":"` + hash + `","transactionIndex":"0x0",` +
			`"blockHash":"` + hash + `","logIndex":"0x1","removed":false}]}`
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c, err := DialContext(context.Background(), srv.URL, srv.Client())
	require.NoError(t, err)

	logs, err := c.FilterLogs(context.Background(), FilterQuery{
		FromBlock: big.NewInt(1693914), ToBlock: big.NewInt(1693914),
		Addresses: []string{"0x524f17a90c20a6de53bc0b82899fbcfef6bb7a39"},
	})
	require.NoError(t, err)
	require.Len(t, logs, 2)
	require.Empty(t, logs[0].Topics)
	require.NotNil(t, logs[0].Topics)
	require.Equal(t, []string{hash}, logs[1].Topics)
	require.Equal(t, uint64(1693914), logs[0].BlockNumber)
	require.JSONEq(t, `{"fromBlock":"0x19d8da","toBlock":"0x19d8da",`+
		`"address":["0x524f17a90c20a6de53bc0b82899fbcfef6bb7a39"]}`, string(gotParams))
}
