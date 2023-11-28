package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type Client struct {
	client *pubsub.Client
	topic  *pubsub.Topic
	logger *zap.SugaredLogger
}

func InitPubsub(ctx context.Context, projectID, topicID string, opts ...option.ClientOption) (*Client, error) {
	l := zap.S()
	l.With("project id", projectID)

	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		l.Errorw("error create new publisher", "error", err)

		return nil, err
	}

	t := client.Topic(topicID)
	t.EnableMessageOrdering = true

	return &Client{
		client: client,
		topic:  t,
		logger: l,
	}, nil
}

func (p *Client) Publish(ctx context.Context, orderingKey string,
	data []byte, extra map[string]string,
) (string, error) {
	p.logger.Infof("publishing message to topic %s", p.topic.ID())
	result := p.topic.Publish(ctx, &pubsub.Message{
		Data:        data,
		Attributes:  extra,
		OrderingKey: orderingKey,
	})

	id, err := result.Get(ctx)
	if err != nil {
		p.logger.Errorf("error publishing message id %s to topic %s: %v", id, p.topic.ID(), err)
	}

	return id, err
}

func (p *Client) initTopic(topicID string) *pubsub.Topic {
	t := p.client.Topic(topicID)
	t.EnableMessageOrdering = true

	return t
}
