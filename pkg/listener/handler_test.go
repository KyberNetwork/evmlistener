package listener

import (
	"testing"

	"github.com/KyberNetwork/evmlistener/pkg/block"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type HandlerTestSuite struct {
	suite.Suite
	handler     *Handler
	evmClient   *EVMClientMock
	blockKeeper block.Keeper
	publisher   *PublisherMock
}

func (ts *HandlerTestSuite) SetupTest() {
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
	ts.blockKeeper = block.NewBaseBlockKeeper(32)
	ts.publisher = NewPublisherMock(1000)
	ts.handler = NewHandler(zap.S(), "test-topic", ts.evmClient, ts.blockKeeper, ts.publisher)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
