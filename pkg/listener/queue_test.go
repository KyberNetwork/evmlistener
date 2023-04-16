package listener

import (
	"math/big"
	"testing"

	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/stretchr/testify/suite"
)

type QueueTestSuite struct {
	suite.Suite
	queue *Queue
}

func (ts *QueueTestSuite) SetupTest() {
	ts.queue = NewQueue(10)
}

func (ts *QueueTestSuite) TestInsertDequeue() {
	ts.queue.Clear()

	blocks := []*types.Block{
		{
			Number: big.NewInt(10),
			Hash:   "0x0a",
		},
		{
			Number: big.NewInt(11),
			Hash:   "0x0b",
		},
		{
			Number: big.NewInt(9),
			Hash:   "0x09",
		},
		{
			Number: big.NewInt(15),
			Hash:   "0x0f",
		},
	}

	for _, block := range blocks {
		ts.queue.Insert(block)
	}

	ts.Require().Equal(uint64(10), ts.queue.BlockNumber())
	ts.Require().Equal(3, ts.queue.Size())

	values := ts.queue.Values()
	ts.Require().Equal(3, len(values))

	ts.Assert().Equal(blocks[0], values[0])
	ts.Assert().Equal(blocks[1], values[1])
	ts.Assert().Equal(blocks[3], values[2])

	value, ok := ts.queue.Dequeue()
	ts.Assert().True(ok)
	if ts.Assert().NotNil(value) {
		ts.Assert().Equal(blocks[0], value)
	}
	ts.Assert().Equal(uint64(11), ts.queue.BlockNumber())
	ts.Assert().Equal(2, ts.queue.Size())

	value, ok = ts.queue.Dequeue()
	ts.Assert().True(ok)
	if ts.Assert().NotNil(value) {
		ts.Assert().Equal(blocks[1], value)
	}
	ts.Assert().Equal(uint64(12), ts.queue.BlockNumber())
	ts.Assert().Equal(1, ts.queue.Size())

	value, ok = ts.queue.Dequeue()
	ts.Assert().False(ok)
	ts.Assert().Nil(value)
	ts.Assert().Equal(uint64(13), ts.queue.BlockNumber())
	ts.Assert().Equal(1, ts.queue.Size())

	value, ok = ts.queue.Dequeue()
	ts.Assert().False(ok)
	ts.Assert().Nil(value)
	ts.Assert().Equal(uint64(14), ts.queue.BlockNumber())
	ts.Assert().Equal(1, ts.queue.Size())

	value, ok = ts.queue.Dequeue()
	ts.Assert().False(ok)
	ts.Assert().Nil(value)
	ts.Assert().Equal(uint64(15), ts.queue.BlockNumber())
	ts.Assert().Equal(1, ts.queue.Size())

	value, ok = ts.queue.Dequeue()
	ts.Assert().True(ok)
	if ts.Assert().NotNil(value) {
		ts.Assert().Equal(blocks[3], value)
	}
	ts.Assert().Equal(uint64(16), ts.queue.BlockNumber())
	ts.Assert().Equal(0, ts.queue.Size())
}

func TestQueueTestSuite(t *testing.T) {
	suite.Run(t, new(QueueTestSuite))
}
