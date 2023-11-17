package pubsub

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
	"go.uber.org/zap"
)

// Publisher ...
type Publisher interface {
	Publish(ctx context.Context, topic string, data interface{}) error
}

type pubSubPublisher struct {
	client      *pubsub.Client
	orderingKey string
	logger      *zap.SugaredLogger
}

func NewPublisher(ctx context.Context, projectID string, orderingKey string) (Publisher, error) {
	l := zap.S()
	l.With("orderingKey", orderingKey)

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		l.Errorf("error create new publisher: %v", err)
		return nil, err
	}

	return &pubSubPublisher{
		client: client,
		logger: l,
	}, nil
}

func (p *pubSubPublisher) Publish(ctx context.Context, topic string, data interface{}) error {
	bytesData, err := json.Marshal(data)
	if err != nil {
		p.logger.Errorf("marshal data to bytes err: %v", err)
		return err
	}

	compressed, err := CompressWithSizePrepended(bytesData)
	if err != nil {
		p.logger.Errorf("compress data fail: %v", err)
		return err
	}

	t := p.client.Topic(topic)
	t.EnableMessageOrdering = true
	defer t.Stop()

	p.logger.Infof("publishing message to topic %s", topic)
	result := t.Publish(ctx, &pubsub.Message{
		OrderingKey: p.orderingKey,
		Data:        compressed,
	})

	id, err := result.Get(ctx)
	if err != nil {
		p.logger.Errorf("error publishing message id %s to topic %s: %v", id, topic, err)
	}

	return err
}
