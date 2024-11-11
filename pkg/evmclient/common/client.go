package common

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// Client extends Ethereum API client with typed wrappers for the FTM API.
type Client struct {
	ethclient.Client
	c *rpc.Client
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}

	return NewClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.Client) *Client {
	return &Client{
		Client: *ethclient.NewClient(c),
		c:      c,
	}
}

type Header struct {
	Hash common.Hash `json:"hash"`
	types.Header
}

//nolint:funlen,cyclop
func (h *Header) UnmarshalJSON(data []byte) error {
	type Header struct {
		Hash            *common.Hash      `json:"hash"`
		ParentHash      *common.Hash      `json:"parentHash"`
		UncleHash       *common.Hash      `json:"sha3Uncles"`
		Coinbase        *common.Address   `json:"miner"`
		Root            *common.Hash      `json:"stateRoot"`
		TxHash          *common.Hash      `json:"transactionsRoot"`
		ReceiptHash     *common.Hash      `json:"receiptsRoot"`
		Bloom           *types.Bloom      `json:"logsBloom"`
		Difficulty      *hexutil.Big      `json:"difficulty"`
		Number          *hexutil.Big      `json:"number"`
		GasLimit        *hexutil.Uint64   `json:"gasLimit"`
		GasUsed         *hexutil.Uint64   `json:"gasUsed"`
		Time            *hexutil.Uint64   `json:"timestamp"`
		Extra           *hexutil.Bytes    `json:"extraData"`
		MixDigest       *common.Hash      `json:"mixHash"`
		Nonce           *types.BlockNonce `json:"nonce"`
		BaseFee         *hexutil.Big      `json:"baseFeePerGas"`
		WithdrawalsHash *common.Hash      `json:"withdrawalsRoot"`
	}

	var dec Header
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	if dec.Hash == nil {
		return errors.New("missing required field 'hash' for Header")
	}
	h.Hash = *dec.Hash

	if dec.ParentHash == nil {
		return errors.New("missing required field 'parentHash' for Header")
	}
	h.ParentHash = *dec.ParentHash

	if dec.UncleHash == nil {
		return errors.New("missing required field 'sha3Uncles' for Header")
	}
	h.UncleHash = *dec.UncleHash

	if dec.Coinbase != nil {
		h.Coinbase = *dec.Coinbase
	}
	if dec.Root == nil {
		return errors.New("missing required field 'stateRoot' for Header")
	}
	h.Root = *dec.Root

	if dec.TxHash == nil {
		return errors.New("missing required field 'transactionsRoot' for Header")
	}
	h.TxHash = *dec.TxHash

	if dec.ReceiptHash == nil {
		return errors.New("missing required field 'receiptsRoot' for Header")
	}
	h.ReceiptHash = *dec.ReceiptHash

	if dec.Bloom == nil {
		return errors.New("missing required field 'logsBloom' for Header")
	}
	h.Bloom = *dec.Bloom

	if dec.Difficulty == nil {
		return errors.New("missing required field 'difficulty' for Header")
	}
	h.Difficulty = (*big.Int)(dec.Difficulty)

	if dec.Number == nil {
		return errors.New("missing required field 'number' for Header")
	}
	h.Number = (*big.Int)(dec.Number)

	if dec.GasLimit == nil {
		return errors.New("missing required field 'gasLimit' for Header")
	}
	h.GasLimit = uint64(*dec.GasLimit)

	if dec.GasUsed == nil {
		return errors.New("missing required field 'gasUsed' for Header")
	}
	h.GasUsed = uint64(*dec.GasUsed)

	if dec.Time == nil {
		return errors.New("missing required field 'timestamp' for Header")
	}
	h.Time = uint64(*dec.Time)

	if dec.Extra == nil {
		return errors.New("missing required field 'extraData' for Header")
	}
	h.Extra = *dec.Extra

	if dec.MixDigest != nil {
		h.MixDigest = *dec.MixDigest
	}
	if dec.Nonce != nil {
		h.Nonce = *dec.Nonce
	}
	if dec.BaseFee != nil {
		h.BaseFee = (*big.Int)(dec.BaseFee)
	}
	if dec.WithdrawalsHash != nil {
		h.WithdrawalsHash = dec.WithdrawalsHash
	}

	return nil
}

func (c *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*Header, error) {
	var head *Header
	err := c.c.CallContext(ctx, &head, "eth_getBlockByNumber", toBlockNumArg(number), false)
	if err == nil && head == nil {
		err = ethereum.NotFound
	}

	return head, err
}

func (c *Client) HeaderByHash(ctx context.Context, hash common.Hash) (*Header, error) {
	var head *Header
	err := c.c.CallContext(ctx, &head, "eth_getBlockByHash", hash, false)
	if err == nil && head == nil {
		err = ethereum.NotFound
	}

	return head, err
}

//nolint:ireturn
func (c *Client) SubscribeNewHead(
	ctx context.Context, ch chan<- *Header,
) (ethereum.Subscription, error) {
	return c.c.EthSubscribe(ctx, ch, "newHeads")
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}

	pending := big.NewInt(-1)
	if number.Cmp(pending) == 0 {
		return "pending"
	}

	finalized := big.NewInt(int64(rpc.FinalizedBlockNumber))
	if number.Cmp(finalized) == 0 {
		return "finalized"
	}

	safe := big.NewInt(int64(rpc.SafeBlockNumber))
	if number.Cmp(safe) == 0 {
		return "safe"
	}

	return hexutil.EncodeBig(number)
}
