package listener

import (
	"fmt"
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
		ts.queue.Insert(block.Number.Uint64()-8, block)
		fmt.Println(ts.queue.String())
	}

	ts.Require().Equal(uint64(1), ts.queue.SequenceNumber())
	ts.Require().Equal(4, ts.queue.Size())

	values := ts.queue.Values()
	ts.Require().Equal(4, len(values))

	ts.Assert().Equal(blocks[2], values[0])
	ts.Assert().Equal(blocks[0], values[1])
	ts.Assert().Equal(blocks[1], values[2])
	ts.Assert().Equal(blocks[3], values[3])

	value, ok := ts.queue.Dequeue()
	ts.Assert().True(ok)
	if ts.Assert().NotNil(value) {
		ts.Assert().Equal(blocks[2], value)
	}
	ts.Assert().Equal(uint64(2), ts.queue.SequenceNumber())
	ts.Assert().Equal(3, ts.queue.Size())

	value, ok = ts.queue.Dequeue()
	ts.Assert().True(ok)
	if ts.Assert().NotNil(value) {
		ts.Assert().Equal(blocks[0], value)
	}
	ts.Assert().Equal(uint64(3), ts.queue.SequenceNumber())
	ts.Assert().Equal(2, ts.queue.Size())

	value, ok = ts.queue.Dequeue()
	ts.Assert().True(ok)
	if ts.Assert().NotNil(value) {
		ts.Assert().Equal(blocks[1], value)
	}
	ts.Assert().Equal(uint64(4), ts.queue.SequenceNumber())
	ts.Assert().Equal(1, ts.queue.Size())

	value, ok = ts.queue.Dequeue()
	ts.Assert().False(ok)
	ts.Assert().Nil(value)
	ts.Assert().Equal(uint64(5), ts.queue.SequenceNumber())
	ts.Assert().Equal(1, ts.queue.Size())

	value, ok = ts.queue.Dequeue()
	ts.Assert().False(ok)
	ts.Assert().Nil(value)
	ts.Assert().Equal(uint64(6), ts.queue.SequenceNumber())
	ts.Assert().Equal(1, ts.queue.Size())

	value, ok = ts.queue.Dequeue()
	ts.Assert().False(ok)
	ts.Assert().Nil(value)
	ts.Assert().Equal(uint64(7), ts.queue.SequenceNumber())
	ts.Assert().Equal(1, ts.queue.Size())

	value, ok = ts.queue.Dequeue()
	ts.Assert().True(ok)
	if ts.Assert().NotNil(value) {
		ts.Assert().Equal(blocks[3], value)
	}
	ts.Assert().Equal(uint64(8), ts.queue.SequenceNumber())
	ts.Assert().Equal(0, ts.queue.Size())
}

func TestQueueTestSuite(t *testing.T) {
	suite.Run(t, new(QueueTestSuite))
}
