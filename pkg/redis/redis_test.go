package redis

import (
	"context"
	"testing"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite

	c *Client
}

func (ts *ClientTestSuite) SetupTest() {
	client, err := New(Config{
		Addrs:     []string{":6379"},
		KeyPrefix: "test:",
	})
	if err != nil {
		panic(err)
	}

	ts.c = client
}

func (ts *ClientTestSuite) TestSet() {
	err := ts.c.Set(context.Background(), "set-1", "test set data", time.Second)
	ts.Assert().NoError(err)
}

func (ts *ClientTestSuite) TestExists() {
	// Test for key non exists.
	exists, err := ts.c.Exists(context.Background(), "non-exists")
	ts.Require().NoError(err)
	ts.Assert().False(exists)

	// Test for key exists.
	err = ts.c.Set(context.Background(), "exists-1", "test key exists", time.Second)
	ts.Require().NoError(err)
	exists, err = ts.c.Exists(context.Background(), "exists-1")
	ts.Require().NoError(err)
	ts.Assert().True(exists)
}

func (ts *ClientTestSuite) TestGet() {
	var s string

	// Test get for non-exists key.
	err := ts.c.Get(context.Background(), "non-exists", &s)
	ts.Require().ErrorIs(err, errors.ErrNotFound)

	// Test get for exists key.
	err = ts.c.Set(context.Background(), "get-1", "test get data", time.Second)
	ts.Require().NoError(err)
	err = ts.c.Get(context.Background(), "get-1", &s)
	ts.Require().NoError(err)
	ts.Assert().Equal("test get data", s)

	// Test get for exists key but wrong output data type.
	var i int
	err = ts.c.Get(context.Background(), "get-1", &i)
	ts.Require().Error(err)
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
