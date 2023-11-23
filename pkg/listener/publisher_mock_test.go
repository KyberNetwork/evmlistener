package listener

import (
	"context"

	"github.com/KyberNetwork/evmlistener/pkg/types"
)

type PublisherMock struct {
	ch chan interface{}
}

func NewPublisherMock(n int) *PublisherMock {
	return &PublisherMock{ch: make(chan interface{}, n)}
}

func (p *PublisherMock) Publish(ctx context.Context, msg types.Message) error {
	p.ch <- msg

	return nil
}
