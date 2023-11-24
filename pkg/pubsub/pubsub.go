package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"go.uber.org/zap"
)

type Client struct {
	client *pubsub.Client
	logger *zap.SugaredLogger
}

func NewPubsub(ctx context.Context, projectID string) (*Client, error) {
	l := zap.S()
	l.With("project id", projectID)

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		l.Errorw("error create new publisher", "error", err)

		return nil, err
	}

	return &Client{
		client: client,
		logger: l,
	}, nil
}

func (p *Client) Publish(ctx context.Context, topic, orderingKey string, data []byte, extra map[string]string) error {
	t := p.client.Topic(topic)
	t.EnableMessageOrdering = true
	defer t.Stop()

	p.logger.Infof("publishing message to topic %s", topic)
	result := t.Publish(ctx, &pubsub.Message{
		Data:        data,
		Attributes:  extra,
		OrderingKey: orderingKey,
	})

	id, err := result.Get(ctx)
	if err != nil {
		p.logger.Errorf("error publishing message id %s to topic %s: %v", id, topic, err)
	}

	return err
}
