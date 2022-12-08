package listener

import (
	"context"
	"testing"

	"github.com/KyberNetwork/evmlistener/pkg/block"
	ltypes "github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum/common"
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
	ts.handler = NewHandler("test-topic", ts.evmClient, ts.blockKeeper, ts.publisher)
}

func (ts *HandlerTestSuite) TestInit() {
	ts.evmClient.SetHead(34)
	blockKeeper := NewBlockKeeperMock(32)
	handler := NewHandler("test-topic", ts.evmClient, blockKeeper, ts.publisher)

	// Init handler without saved data.
	err := handler.Init(context.Background())
	if ts.Assert().NoError(err) {
		ts.Assert().Equal(32, blockKeeper.Len())
	}

	// Init handler with saved data.
	blocks, err := blockKeeper.GetRecentBlocks(16)
	ts.Require().NoError(err)
	blockKeeper.SetInitData(blocks)

	err = handler.Init(context.Background())
	if ts.Assert().NoError(err) {
		ts.Assert().Equal(16, blockKeeper.Len())
	}
}

//nolint
func (ts *HandlerTestSuite) TestHandle() {
	ts.evmClient.SetHead(44)

	err := ts.handler.Init(context.Background())
	ts.Require().NoError(err)

	// Handle for handled block.
	b, err := ts.blockKeeper.Head()
	ts.Require().NoError(err)

	err = ts.handler.Handle(context.Background(), b)
	ts.Require().NoError(err)
	ts.Assert().Equal(0, len(ts.publisher.ch))

	// Handle for normal block (chain was not re-organized).
	ts.evmClient.Next()
	hash := common.HexToHash("0xc0c29448be86bca9d0db94b79cd1a6bd1361aed1e394d3a2a218fb98b159ab74")
	b, err = getBlockByHash(context.Background(), ts.evmClient, hash)
	ts.Require().NoError(err)

	err = ts.handler.Handle(context.Background(), b)
	ts.Require().NoError(err)
	ts.Require().Equal(1, len(ts.publisher.ch))
	data := <-ts.publisher.ch
	msg, ok := data.(ltypes.Message)
	ts.Require().True(ok)
	ts.Require().Equal(0, len(msg.RevertedBlocks))
	ts.Require().Equal(1, len(msg.NewBlocks))
	ts.Assert().Equal(b.Hash.String(), msg.NewBlocks[0].Hash.String())
	ts.Assert().Equal(len(b.Logs), len(msg.NewBlocks[0].Logs))

	// Handle for far away block (lost connection).
	ts.evmClient.SetHead(52)
	hash = common.HexToHash("0x132c1eb1799a5219b055674177ba95e946feb5f011c7c1409630d42c0581ee52")
	b, err = getBlockByHash(context.Background(), ts.evmClient, hash)
	ts.Require().NoError(err)

	err = ts.handler.Handle(context.Background(), b)
	ts.Require().NoError(err)
	ts.Require().Equal(1, len(ts.publisher.ch))
	data = <-ts.publisher.ch
	msg, ok = data.(ltypes.Message)
	ts.Require().True(ok)
	ts.Require().Equal(0, len(msg.RevertedBlocks))
	ts.Require().Equal(7, len(msg.NewBlocks))
	ts.Assert().Equal(b.Hash.String(), msg.NewBlocks[6].Hash.String())
	ts.Assert().Equal(len(b.Logs), len(msg.NewBlocks[6].Logs))

	// Handle for re-org block.
	ts.evmClient.Next()
	head, err := ts.blockKeeper.Head()
	ts.Require().NoError(err)

	hash = common.HexToHash("0xfe5db0e13993eb721f8174edc783e92dcee70e5a2eb3cd87e8b6c7ba5ab24986")
	b, err = getBlockByHash(context.Background(), ts.evmClient, hash)
	ts.Require().NoError(err)

	err = ts.handler.Handle(context.Background(), b)
	ts.Require().NoError(err)
	ts.Require().Equal(1, len(ts.publisher.ch))
	data = <-ts.publisher.ch
	msg, ok = data.(ltypes.Message)
	ts.Require().True(ok)
	ts.Require().Equal(1, len(msg.RevertedBlocks))
	ts.Require().Equal(1, len(msg.NewBlocks))
	ts.Assert().Equal(head.Hash.String(), msg.RevertedBlocks[0].Hash.String())
	ts.Assert().Equal(len(head.Logs), len(msg.RevertedBlocks[0].Logs))
	ts.Assert().Equal(b.Hash.String(), msg.NewBlocks[0].Hash.String())
	ts.Assert().Equal(len(b.Logs), len(msg.NewBlocks[0].Logs))

	// Handle for re-org block plus missing block.
	ts.evmClient.SetHead(54)
	err = ts.handler.Init(context.Background())
	ts.Require().NoError(err)

	head, err = ts.blockKeeper.Head()
	ts.Require().NoError(err)

	ts.evmClient.Next()
	hash = common.HexToHash("0x2394b0b03959156ec90096deadd34f68195a8d8f5f1e5438ea237be7675178c2")
	b, err = getBlockByHash(context.Background(), ts.evmClient, hash)
	ts.Require().NoError(err)

	err = ts.handler.Handle(context.Background(), b)
	ts.Require().NoError(err)
	ts.Require().Equal(1, len(ts.publisher.ch))
	data = <-ts.publisher.ch
	msg, ok = data.(ltypes.Message)
	ts.Require().True(ok)
	ts.Require().Equal(2, len(msg.RevertedBlocks))
	ts.Require().Equal(2, len(msg.NewBlocks))
	ts.Assert().Equal(head.Hash.String(), msg.RevertedBlocks[0].Hash.String())
	ts.Assert().Equal(len(head.Logs), len(msg.RevertedBlocks[0].Logs))
	ts.Assert().Equal(b.Hash.String(), msg.NewBlocks[1].Hash.String())
	ts.Assert().Equal(len(b.Logs), len(msg.NewBlocks[1].Logs))
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
