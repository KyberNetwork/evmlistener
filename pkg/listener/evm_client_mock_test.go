package listener

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"os"
	"sync"

	"github.com/KyberNetwork/evmlistener/pkg/common"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/KyberNetwork/evmlistener/protobuf/pb"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type ClientSubscription struct {
	errCh chan error
}

func (s *ClientSubscription) Err() <-chan error {
	return s.errCh
}

func (s *ClientSubscription) Unsubscribe() {}

type EVMClientMock struct {
	mu         sync.Mutex
	head       int
	sequence   []ethcommon.Hash
	headerMap  map[ethcommon.Hash]*ethtypes.Header
	logsMap    map[ethcommon.Hash][]ethtypes.Log
	subHeadChs []chan<- *types.Header
	subs       []*ClientSubscription
}

func (c *EVMClientMock) GetFullBlockByHash(ctx context.Context, s string) (*pb.Block, error) {
	//TODO implement me
	panic("implement me")
}

func NewEVMClientMock(dataFile string) (*EVMClientMock, error) {
	data := struct {
		HeadSequence []ethcommon.Hash                    `json:"headSequence"`
		HeaderMap    map[ethcommon.Hash]*ethtypes.Header `json:"headerMap"`
		LogsMap      map[ethcommon.Hash][]ethtypes.Log   `json:"logsMap"`
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
	c.mu.Lock()
	defer c.mu.Unlock()

	c.head++

	// Publish new head for subscriptions.
	for _, ch := range c.subHeadChs {
		go func(ch chan<- *types.Header) {
			header := c.headerMap[c.sequence[c.head]]
			ch <- &types.Header{
				Hash:       common.ToHex(header.Hash()),
				ParentHash: common.ToHex(header.ParentHash),
				Number:     header.Number,
				Time:       header.Time,
			}
		}(ch)
	}
}

func (c *EVMClientMock) SetHead(index int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.head = index
}

func (c *EVMClientMock) BlockNumber(ctx context.Context) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	header := c.headerMap[c.sequence[c.head]]

	return header.Number.Uint64(), nil
}

// nolint
func (c *EVMClientMock) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (evmclient.Subscription, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.subHeadChs = append(c.subHeadChs, ch)

	// Publish current head to the channel.
	go func() {
		header := c.headerMap[c.sequence[c.head]]
		ch <- &types.Header{
			Hash:       common.ToHex(header.Hash()),
			ParentHash: common.ToHex(header.ParentHash),
			Number:     header.Number,
			Time:       header.Time,
		}
	}()

	sub := &ClientSubscription{errCh: make(chan error)}
	c.subs = append(c.subs, sub)

	return sub, nil
}

// nolint
func (c *EVMClientMock) FilterLogs(ctx context.Context, filter evmclient.FilterQuery) ([]types.Log, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var logs []ethtypes.Log
	if filter.BlockHash != nil {
		logs = c.logsMap[ethcommon.HexToHash(*filter.BlockHash)]
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

	res := make([]types.Log, 0, len(logs))
	for _, log := range logs {
		topics := make([]string, 0, len(log.Topics))
		for _, topic := range log.Topics {
			topics = append(topics, common.ToHex(topic))
		}

		res = append(res, types.Log{
			Address:     common.ToHex(log.Address),
			Topics:      topics,
			Data:        log.Data,
			BlockNumber: log.BlockNumber,
			TxHash:      common.ToHex(log.TxHash),
			BlockHash:   common.ToHex(log.BlockHash),
			Index:       log.Index,
			Removed:     log.Removed,
		})
	}

	// Filter logs by addresses
	if len(filter.Addresses) > 0 {
		var filterLogs []types.Log
		for _, log := range res {
			for _, address := range filter.Addresses {
				if address == log.Address {
					filterLogs = append(filterLogs, log)
				}
			}
		}

		res = filterLogs
	}

	// Filter logs by topics
	if len(filter.Topics) > 0 {
		var filterLogs []types.Log
		for _, log := range res {
			match := true
			for i, topic := range filter.Topics {
				for _, v := range topic {
					if log.Topics[i] == v {
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
		res = filterLogs
	}

	return res, nil
}

func (c *EVMClientMock) HeaderByNumber(ctx context.Context, num *big.Int) (*types.Header, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var ok bool

	header := c.headerMap[c.sequence[c.head]]
	for {
		cmp := num.Cmp(header.Number)
		if cmp == 0 {
			return &types.Header{
				Hash:       common.ToHex(header.Hash()),
				ParentHash: common.ToHex(header.ParentHash),
				Number:     header.Number,
				Time:       header.Time,
			}, nil
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

func (c *EVMClientMock) HeaderByHash(ctx context.Context, hash string) (*types.Header, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	header, ok := c.headerMap[ethcommon.HexToHash(hash)]
	if !ok {
		return nil, errors.New("header not found") //nolint
	}

	return &types.Header{
		Hash:       common.ToHex(header.Hash()),
		ParentHash: common.ToHex(header.ParentHash),
		Number:     header.Number,
		Time:       header.Time,
	}, nil
}

func (c *EVMClientMock) NotifyDisconnect(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, sub := range c.subs {
		go func(sub *ClientSubscription) {
			sub.errCh <- err
		}(sub)
	}
}
