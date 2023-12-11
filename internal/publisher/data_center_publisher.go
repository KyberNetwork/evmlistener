package publisher

import (
	"context"

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

	data, err := p.packMsgData(msg)
	if err != nil {
		p.logger.Errorf("error on packing message to publish: %v", err)

		return err
	}

	// extra := p.extractExtraData(b)

	msgID, err := p.client.Publish(ctx, p.config.OrderingKey, data, nil)
	if err != nil {
		p.logger.Errorf("error publish to pubsub: %v", err)

		return err
	}

	p.logger.Debugf("Done publish %d new blocks, %d reverted blocks with message id %s",
		len(msg.NewBlocks), len(msg.RevertedBlocks), msgID)

	return nil
}

// func (p *DataCenterPublisher) extractExtraData(block types.Block) map[string]string {
//	return map[string]string{
//		"block_number":    block.Header.Number.String(),
//		"block_hash":      block.Hash,
//		"parent_hash":     block.Header.ParentHash,
//		"block_timestamp": strconv.Itoa(int(block.Header.Timestamp)),
//	}
//}

func (p *DataCenterPublisher) packMsgData(msg types.Message) ([]byte, error) {
	msgData := msg.ToProtobuf()
	bytesData, err := proto.Marshal(msgData)
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
