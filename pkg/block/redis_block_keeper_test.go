package block

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/KyberNetwork/evmlistener/pkg/redis"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
)

type RedisBlockKeeperTestSuite struct {
	suite.Suite

	redisClient *redis.Client
	keeper      *RedisBlockKeeper
}

func (ts *RedisBlockKeeperTestSuite) SetupTest() {
	rand.Seed(time.Now().UnixNano())
	prefix := fmt.Sprintf("test-redis-block-keeper-%d:", rand.Int()) // nolint

	redisClient, err := redis.New(redis.Config{
		Addrs:     []string{":6379"},
		KeyPrefix: prefix,
	})
	if err != nil {
		panic(err)
	}

	keeper := NewRedisBlockKeeper(redisClient, 4, time.Second)

	ts.redisClient = redisClient
	ts.keeper = keeper

	var _ Keeper = keeper
}

func (ts *RedisBlockKeeperTestSuite) TestInit() {
	// First time initialization.
	err := ts.keeper.Init()
	if ts.Assert().NoError(err) {
		ts.Assert().Equal(0, ts.keeper.Len())
	}

	// Second time initialization.
	for _, block := range sampleBlocks {
		err = ts.redisClient.Set(context.Background(), block.Hash.String(), block, time.Second)
		ts.Require().NoError(err)
	}
	err = ts.redisClient.Set(context.Background(), blockHeadKey, "0x9a24538f47e0c6faa56732a0c3f1f036bea5372a57369c3ecef1423972957c6a", time.Second) // nolint
	ts.Require().NoError(err)

	err = ts.keeper.Init()
	if ts.Assert().NoError(err) {
		ts.Assert().Equal(len(sampleBlocks), ts.keeper.Len())
	}
}

func (ts *RedisBlockKeeperTestSuite) TestAdd() {
	block := types.Block{
		Number:     big.NewInt(35338115),
		Hash:       common.HexToHash("0xf11b9c19c31319321e6730754f4fe1746f24d1b6ca925d30622059e6a5d79450"),
		ParentHash: common.HexToHash("0x9a24538f47e0c6faa56732a0c3f1f036bea5372a57369c3ecef1423972957c6a"),
	}

	// Test for adding new block.
	err := ts.keeper.Add(block)
	ts.Assert().NoError(err)

	// Test for adding existing block.
	err = ts.keeper.Add(block)
	ts.Assert().ErrorIs(err, errors.ErrAlreadyExists)
}

func TestRedisBlockKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(RedisBlockKeeperTestSuite))
}
