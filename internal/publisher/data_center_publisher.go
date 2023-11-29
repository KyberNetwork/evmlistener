package publisher

import (
	"context"
	"strconv"

	"github.com/KyberNetwork/evmlistener/pkg/common"
	"github.com/KyberNetwork/evmlistener/pkg/pubsub"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type DataCenterPublisher struct {
	client *pubsub.Client
	config Config
	logger *zap.SugaredLogger
}

func NewDataCenterPublisher(client *pubsub.Client, cfg Config) *DataCenterPublisher {
	l := zap.S()
	l.With("orderingKey", cfg.OrderingKey)

	return &DataCenterPublisher{
		client: client,
		config: cfg,
		logger: l,
	}
}

func (p *DataCenterPublisher) Publish(ctx context.Context, msg types.Message) error {
	p.logger.Infow("Publish message to queue",
		"topic", p.config.Topic,
		"numRevertedBlocks", len(msg.RevertedBlocks),
		"numNewBlocks", len(msg.NewBlocks))

	if len(msg.RevertedBlocks) > 0 {
		p.logger.Warnf("%d of blocks is re-orged", len(msg.RevertedBlocks))
	}

	for _, b := range msg.NewBlocks {
		data, err := p.packMsgData(b)
		if err != nil {
			p.logger.Errorf("error on packing message to publish: %v", err)

			return err
		}
		extra := p.extractExtraData(b)

		msgID, err := p.client.Publish(ctx, p.config.OrderingKey, data, extra)
		if err != nil {
			p.logger.Errorf("error publish block %d to pubsub: %v", b.Header.Number.Uint64(), err)

			return err
		}

		p.logger.Debugf("Done publish block %d with message id %s", b.Header.Number.Uint64(), msgID)
	}

	return nil
}

func (p *DataCenterPublisher) extractExtraData(block types.Block) map[string]string {
	return map[string]string{
		"block_number":    block.Header.Number.String(),
		"block_hash":      block.Hash,
		"parent_hash":     block.Header.ParentHash,
		"block_timestamp": strconv.Itoa(int(block.Header.Timestamp)),
	}
}

func (p *DataCenterPublisher) packMsgData(b types.Block) ([]byte, error) {
	block := b.ToProtobuf()
	bytesData, err := proto.Marshal(block)
	if err != nil {
		p.logger.Errorf("marshal data to bytes err: %v", err)

		return nil, err
	}

	compressed, err := common.CompressWithSizePrepended(bytesData)
	if err != nil {
		p.logger.Errorf("compress data fail: %v", err)

		return nil, err
	}

	return compressed, nil
}
