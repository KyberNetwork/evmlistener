package publisher

import (
	"context"
	"strconv"

	"github.com/KyberNetwork/evmlistener/pkg/common"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type DataCentralPublisher struct {
	client Client
	config Config
	logger *zap.SugaredLogger
}

func NewDataCentralPublisher(client Client, cfg Config) *DataCentralPublisher {
	l := zap.S()
	l.With("orderingKey", cfg.OrderingKey)

	return &DataCentralPublisher{
		client: client,
		config: cfg,
		logger: l,
	}
}

func (p *DataCentralPublisher) Publish(ctx context.Context, msg types.Message) error {
	p.logger.Infow("Publish message to queue",
		"topic", p.config.Topic,
		"numRevertedBlocks", len(msg.RevertedBlocks),
		"numNewBlocks", len(msg.NewBlocks))

	for _, b := range msg.NewBlocks {
		data, err := p.packMsgData(b)
		if err != nil {
			p.logger.Errorf("error on packing message to publish: %v", err)

			return err
		}
		extra := p.extractExtraData(b)

		err = p.client.Publish(ctx, p.config, data, extra)
		if err != nil {
			p.logger.Errorf("error publish block %d to pubsub: %v", b.Number.Uint64(), err)

			return err
		}
	}

	return nil
}

func (p *DataCentralPublisher) extractExtraData(block types.Block) map[string]string {
	return map[string]string{
		"block_number":    block.Number.String(),
		"block_hash":      block.Hash,
		"parent_hash":     block.ParentHash,
		"block_timestamp": strconv.Itoa(int(block.Timestamp)),
	}
}

func (p *DataCentralPublisher) packMsgData(b types.Block) ([]byte, error) {
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
