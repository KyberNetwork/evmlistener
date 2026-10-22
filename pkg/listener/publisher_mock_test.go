package listener

import "context"

type PublisherMock struct {
	ch chan any
}

func NewPublisherMock(n int) *PublisherMock {
	return &PublisherMock{ch: make(chan any, n)}
}

func (p *PublisherMock) Publish(ctx context.Context, topic string, msg []byte) error {
	p.ch <- msg

	return nil
}
