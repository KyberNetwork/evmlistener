package publisher

import (
	"context"
	"encoding/json"

	"github.com/KyberNetwork/evmlistener/pkg/redis"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"go.uber.org/zap"
)

type RedisStreamPublisher struct {
	client *redis.Stream
	config Config
	logger *zap.SugaredLogger
}

func NewRedisStreamPublisher(client *redis.Stream, cfg Config) *RedisStreamPublisher {
	return &RedisStreamPublisher{
		client: client,
		config: cfg,
	}
}

func (p *RedisStreamPublisher) Publish(ctx context.Context, msg types.Message) error {
	p.logger.Infow("Publish message to queue",
		"topic", p.config.Topic,
		"numRevertedBlocks", len(msg.RevertedBlocks),
		"numNewBlocks", len(msg.NewBlocks))

	data, err := p.packMsgData(msg)
	if err != nil {
		p.logger.Errorf("error on packing message to publish: %v", err)

		return err
	}

	err = p.client.Publish(ctx, p.config.Topic, data)
	if err != nil {
		p.logger.Errorf("error publish to redis stream: %v", err)

		return err
	}

	return nil
}

func (p *RedisStreamPublisher) packMsgData(msg types.Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		p.logger.Errorf("error marshalling message: %v", err)

		return nil, err
	}

	return data, nil
}
