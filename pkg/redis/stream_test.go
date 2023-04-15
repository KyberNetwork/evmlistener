package redis

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
)

type StreamTestSuite struct {
	suite.Suite

	s *Stream
}

func (ts *StreamTestSuite) SetupTest() {
	client, err := New(Config{
		Addrs:     []string{":6379"},
		KeyPrefix: "test:",
	})
	if err != nil {
		panic(err)
	}

	ts.s = NewStream(client, 3)
}

func (ts *StreamTestSuite) TestPublish() {
	topic := fmt.Sprintf("test-redis-stream-%d\n", rand.Int()) // nolint

	tests := []struct {
		msg    string
		expect string
	}{
		{
			msg:    "Message 1",
			expect: "\"Message 1\"",
		},
		{
			msg:    "Message 2",
			expect: "\"Message 2\"",
		},
		{
			msg:    "Message 3",
			expect: "\"Message 3\"",
		},
		{
			msg:    "Message 4",
			expect: "\"Message 4\"",
		},
	}

	for _, test := range tests {
		err := ts.s.Publish(context.Background(), topic, test.msg)
		ts.Require().NoError(err)

		res, err := ts.s.client.XRevRangeN(context.Background(), topic, "+", "-", 1).Result()
		ts.Require().NoError(err)
		ts.Require().Len(res, 1)

		msg, ok := res[0].Values[MessageKey].(string)
		ts.Require().True(ok)
		ts.Assert().Equal(test.expect, msg)
	}
}

func TestStreamTestSuite(t *testing.T) {
	suite.Run(t, new(StreamTestSuite))
}
