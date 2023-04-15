package listener

import (
	"context"
	"testing"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/block"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type ListenerTestSuite struct {
	suite.Suite
	evmClient *EVMClientMock
	publisher *PublisherMock
	listener  *Listener
}

func (ts *ListenerTestSuite) SetupTest() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)

	evmClient, err := NewEVMClientMock("data.json")
	if err != nil {
		panic(err)
	}

	ts.evmClient = evmClient
	ts.publisher = NewPublisherMock(1000)
	blockKeeper := block.NewBaseBlockKeeper(32)
	handler := NewHandler(zap.S(), "test-topic", ts.evmClient, blockKeeper, ts.publisher)
	ts.listener = New(zap.S(), evmClient, handler, nil, 0, 0)
}

func (ts *ListenerTestSuite) TestRun() {
	ts.evmClient.SetHead(43)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)

	go func() {
		err := ts.listener.Run(ctx)
		errCh <- err
	}()

	// Sent new head to listeneer.
	for i := 0; i < 11; i++ {
		time.Sleep(100 * time.Millisecond)
		ts.evmClient.Next()
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	err := <-errCh

	if ts.Assert().NoError(err) {
		ts.Assert().Equal(11, len(ts.publisher.ch))
	}

	// Test for re-subscribe.
	ctx, cancel = context.WithCancel(context.Background())

	go func() {
		err := ts.listener.Run(ctx)
		errCh <- err
	}()

	time.Sleep(100 * time.Millisecond)
	err = &websocket.CloseError{
		Code: websocket.CloseServiceRestart,
		Text: "service was restarted",
	}
	ts.evmClient.NotifyDisconnect(err)

	time.Sleep(100 * time.Millisecond)
	cancel()
	err = <-errCh
	ts.Assert().NoError(err)
}

func TestListenerTestSuite(t *testing.T) {
	suite.Run(t, new(ListenerTestSuite))
}
