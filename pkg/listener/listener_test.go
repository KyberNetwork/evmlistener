package listener

import (
	"testing"

	"github.com/KyberNetwork/evmlistener/pkg/block"
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
	ts.listener = New(zap.S(), evmClient, handler)
}

func TestListenerTestSuite(t *testing.T) {
	suite.Run(t, new(ListenerTestSuite))
}
