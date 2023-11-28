package publisher

import (
	"context"
	"math/big"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/KyberNetwork/evmlistener/pkg/common"
	evmPubsub "github.com/KyberNetwork/evmlistener/pkg/pubsub"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/KyberNetwork/evmlistener/protobuf/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

// nolint
func TestDataCentralPublisher_Publish(t *testing.T) {
	ctx := context.TODO()
	c, srv, topic, sub := initFakePubsub(t, ctx)
	defer topic.Stop()
	defer c.Close()
	defer srv.Close()

	client, err := evmPubsub.InitPubsub(ctx, c.Project(),
		option.WithEndpoint(srv.Addr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	if err != nil {
		t.Fatal(err)
	}

	publisher := NewDataCentralPublisher(client, Config{
		Topic:       topic.ID(),
		OrderingKey: "k",
	})

	err = publisher.Publish(ctx, types.Message{
		RevertedBlocks: []types.Block{
			{
				Hash: "block1",
				Header: types.Header{
					Hash:       "block1",
					ParentHash: "block1_parent",
					Number:     big.NewInt(1),
					Time:       123,
					UncleHash:  "block1_uncle",
					Nonce:      1,
				},
			},
		},
		NewBlocks: []types.Block{
			{
				Hash: "block2",
				Header: types.Header{
					Hash:          "block2",
					ParentHash:    "block2_parent",
					Number:        big.NewInt(2),
					Time:          124,
					UncleHash:     "block2_uncle",
					Nonce:         2,
					Difficulty:    big.NewInt(22),
					BaseFeePerGas: big.NewInt(222),
				},
			},
			{
				Hash: "block3",
				Header: types.Header{
					Hash:          "block3",
					ParentHash:    "block3_parent",
					Number:        big.NewInt(3),
					Time:          125,
					UncleHash:     "block3_uncle",
					Nonce:         3,
					Difficulty:    big.NewInt(33),
					BaseFeePerGas: big.NewInt(333),
				},
			},
		},
	})
	assert.NoError(t, err)

	sender := make(chan *pb.Block)
	childCtx, done := context.WithCancel(ctx)
	go getBlockFromSub(t, childCtx, sub, sender)

	var blocks []*pb.Block
	counter := 3
	for {
		select {
		case b := <-sender:
			blocks = append(blocks, b)
		case <-time.After(1 * time.Second):
			counter -= 1
			if counter == 0 {
				assert.Len(t, blocks, 2)
				assert.Equal(t, uint64(2), blocks[0].Number)
				assert.Equal(t, uint64(3), blocks[1].Number)

				t.Log("Done counter")
				done()
			}
		case <-childCtx.Done():
			t.Log("Done context")
			done()

			return
		}
	}
}

func getBlockFromSub(t *testing.T, ctx context.Context, sub *pubsub.Subscription, sender chan<- *pb.Block) {
	t.Helper()
	sub.ReceiveSettings.NumGoroutines = 1

	err := sub.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
		assert.Len(t, message.Attributes, 4, "must contain extra info")
		_, ok := message.Attributes["block_number"]
		assert.True(t, ok)
		_, ok = message.Attributes["block_hash"]
		assert.True(t, ok)

		data, err := common.DecompressWithSizePrepended(message.Data)
		assert.NoError(t, err)

		var block pb.Block
		err = proto.Unmarshal(data, &block)
		assert.NoError(t, err)

		sender <- &block
		t.Logf("got block %d", block.Number)
	})

	assert.NoError(t, err)
}

func initFakePubsub(t *testing.T, ctx context.Context) (
	*pubsub.Client, *pstest.Server, *pubsub.Topic, *pubsub.Subscription,
) {
	t.Helper()
	srv := pstest.NewServer()

	c, err := pubsub.NewClient(ctx, "P",
		option.WithEndpoint(srv.Addr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	if err != nil {
		t.Fatal(err)
	}

	topic, err := c.CreateTopic(ctx, "t")
	if err != nil {
		t.Fatal(err)
	}
	topic.EnableMessageOrdering = true

	sub, err := c.CreateSubscription(ctx, "s", pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if err != nil {
		t.Fatal(err)
	}

	return c, srv, topic, sub
}
