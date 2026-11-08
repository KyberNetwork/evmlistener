package evmclient

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	hash = "0x" + "11111111111111111111111111111111111111111111111111111111111111aa"
	addr = "0x524f17a90c20a6de53bc0b82899fbcfef6bb7a39"
)

// testLog builds a log JSON; fields is spliced in for address/topics so tests can omit or null them.
func testLog(idx, fields string) string {
	return `{` + fields + `"data":"0x01","blockNumber":"0x19d8da","transactionHash":"` + hash +
		`","transactionIndex":"0x0","blockHash":"` + hash + `","logIndex":"` + idx + `","removed":false}`
}

const (
	okFields   = `"address":"` + addr + `","topics":["` + hash + `"],`
	nullFields = `"address":"` + addr + `","topics":null,`
	noTopics   = `"address":"` + addr + `",`
	noAddr     = `"topics":["` + hash + `"],`
)

func filterVia(t *testing.T, result string) ([]uint, error) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			ID json.RawMessage `json:"id"`
		}
		_ = json.Unmarshal(body, &req)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":` + string(req.ID) + `,"result":` + result + `}`))
	}))
	defer srv.Close()

	c, err := DialContext(context.Background(), srv.URL, srv.Client())
	require.NoError(t, err)

	logs, err := c.FilterLogs(context.Background(), FilterQuery{
		FromBlock: big.NewInt(1693914), ToBlock: big.NewInt(1693914),
	})
	var idx []uint
	for _, l := range logs {
		idx = append(idx, l.Index)
	}

	return idx, err
}

// Nodes like Conduit return `"topics": null` for LOG0. Those logs are dropped, nothing else is.
func TestFilterLogs(t *testing.T) {
	join := func(ls ...string) string { return "[" + strings.Join(ls, ",") + "]" }

	idx, err := filterVia(t, join(testLog("0x0", okFields), testLog("0x1", okFields)))
	require.NoError(t, err)
	require.Equal(t, []uint{0, 1}, idx)

	idx, err = filterVia(t, join(testLog("0x0", nullFields)))
	require.NoError(t, err)
	require.Empty(t, idx)

	idx, err = filterVia(t, join(testLog("0x0", noTopics)))
	require.NoError(t, err)
	require.Empty(t, idx)

	idx, err = filterVia(t, join(testLog("0x0", okFields), testLog("0x1", nullFields),
		testLog("0x2", okFields), testLog("0x3", noTopics), testLog("0x4", okFields)))
	require.NoError(t, err)
	require.Equal(t, []uint{0, 2, 4}, idx)

	_, err = filterVia(t, join(testLog("0x0", okFields), testLog("0x1", noAddr)))
	require.ErrorContains(t, err, "missing required field 'address'")

	_, err = filterVia(t, `{"not":"an array"}`)
	require.Error(t, err)
}

func TestDecodeLogsKeepsTopics(t *testing.T) {
	logs, err := decodeLogs(json.RawMessage(`[` + testLog("0x7", okFields) + `]`))
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, uint(7), logs[0].Index)
	require.Len(t, logs[0].Topics, 1)
	require.Equal(t, uint64(1693914), logs[0].BlockNumber)
}
