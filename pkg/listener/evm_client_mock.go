package listener

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type EVMClientMock struct {
	head       int
	sequence   []common.Hash
	headerMap  map[common.Hash]*types.Header
	logsMap    map[common.Hash][]types.Log
	subHeadChs []chan<- *types.Header
}

func NewEVMClientMock(dataFile string) (*EVMClientMock, error) {
	data := struct {
		HeadSequence []common.Hash                 `json:"headSequence"`
		HeaderMap    map[common.Hash]*types.Header `json:"headerMap"`
		LogsMap      map[common.Hash][]types.Log   `json:"logsMap"`
	}{}

	f, err := os.Open(dataFile)
	if err != nil {
		return nil, err
	}

	bs, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bs, &data)
	if err != nil {
		return nil, err
	}

	return &EVMClientMock{
		head:      0, //nolint
		sequence:  data.HeadSequence,
		headerMap: data.HeaderMap,
		logsMap:   data.LogsMap,
	}, nil
}

func (c *EVMClientMock) Next() {
	c.head++

	// Publish new head for subscriptions.
	for _, ch := range c.subHeadChs {
		go func(ch chan<- *types.Header) {
			header := c.headerMap[c.sequence[c.head]]
			ch <- header
		}(ch)
	}
}

func (c *EVMClientMock) SetHead(index int) {
	c.head = index
}

func (c *EVMClientMock) BlockNumber(ctx context.Context) (uint64, error) {
	header := c.headerMap[c.sequence[c.head]]

	return header.Number.Uint64(), nil
}

type ClientSubscription struct {
	errCh chan error
}

func (s *ClientSubscription) Err() <-chan error {
	return s.errCh
}

func (s *ClientSubscription) Unsubscribe() {}

//nolint
func (c *EVMClientMock) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	c.subHeadChs = append(c.subHeadChs, ch)

	// Publish current head to the channel.
	go func() {
		header := c.headerMap[c.sequence[c.head]]
		ch <- header
	}()

	return &ClientSubscription{errCh: make(chan error)}, nil
}

//nolint
func (c *EVMClientMock) FilterLogs(ctx context.Context, filter ethereum.FilterQuery) (logs []types.Log, err error) {
	if filter.BlockHash != nil {
		logs = c.logsMap[*filter.BlockHash]
	} else {
		var ok bool
		header := c.headerMap[c.sequence[c.head]]
		if filter.ToBlock != nil {
			for {
				cmp := filter.ToBlock.Cmp(header.Number)
				if cmp >= 0 {
					break
				}

				header, ok = c.headerMap[header.ParentHash]
				if !ok {
					return nil, nil
				}
			}
		}

		for {
			if filter.FromBlock != nil && filter.FromBlock.Cmp(header.Number) > 0 {
				break
			}

			logs = append(logs, c.logsMap[header.Hash()]...)
			header, ok = c.headerMap[header.ParentHash]
			if !ok {
				break
			}
		}
	}

	// Filter logs by addresses
	if len(filter.Addresses) > 0 {
		var filterLogs []types.Log
		for _, log := range logs {
			for _, address := range filter.Addresses {
				if address.String() == log.Address.String() {
					filterLogs = append(filterLogs, log)
				}
			}
		}

		logs = filterLogs
	}

	// Filter logs by topics
	if len(filter.Topics) > 0 {
		var filterLogs []types.Log
		for _, log := range logs {
			match := true
			for i, topic := range filter.Topics {
				for _, v := range topic {
					if log.Topics[i].String() == v.String() {
						match = true
						break
					}
					match = false
				}
				if !match {
					break
				}
			}
			if match {
				filterLogs = append(filterLogs, log)
			}
		}
		logs = filterLogs
	}

	return logs, nil
}

func (c *EVMClientMock) HeaderByNumber(ctx context.Context, num *big.Int) (*types.Header, error) {
	var ok bool

	header := c.headerMap[c.sequence[c.head]]
	for {
		cmp := num.Cmp(header.Number)
		if cmp == 0 {
			return header, nil
		}

		if cmp > 0 {
			return nil, errors.New("header not found") //nolint
		}

		header, ok = c.headerMap[header.ParentHash]
		if !ok {
			return nil, errors.New("header not found") //nolint
		}
	}
}

func (c *EVMClientMock) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	header, ok := c.headerMap[hash]
	if !ok {
		return nil, errors.New("header not found") //nolint
	}

	return header, nil
}
