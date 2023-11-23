package listener

import (
	"context"
	"math/big"
	"sync"
	"syscall"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	pkgmetric "github.com/KyberNetwork/kyber-trace-go/pkg/metric"
	"github.com/ethereum/go-ethereum"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	bufLen = 100

	maxQueueLen = 256

	metricNameLastReceivedBlockNumber = "evmlistener_last_received_block_number"
	metricNameLastCheckedBlockNumber  = "evmlistener_last_checked_block_number"
	metricNameLastHandledBlockNumber  = "evmlistener_last_handled_block_number"
)

// Listener represents a listener service for on-chain events.
type Listener struct {
	l *zap.SugaredLogger

	wsEVMClient   evmclient.IClient
	httpEVMClient evmclient.IClient
	handler       *Handler

	sanityEVMClient     evmclient.IClient
	sanityCheckInterval time.Duration

	mu                     sync.Mutex
	lastReceivedBlock      *types.Block
	lastHandledBlockNumber *big.Int
	lastCheckedBlockNumber *big.Int
	resuming               bool

	queue       *Queue
	maxQueueLen int
}

// New ...
func New(
	l *zap.SugaredLogger, wsEVMClient evmclient.IClient,
	httpEVMClient evmclient.IClient, handler *Handler,
	sanityEVMClient evmclient.IClient, sanityCheckInterval time.Duration,
) *Listener {
	return &Listener{
		l: l,

		wsEVMClient:   wsEVMClient,
		httpEVMClient: httpEVMClient,
		handler:       handler,

		sanityEVMClient:     sanityEVMClient,
		sanityCheckInterval: sanityCheckInterval,

		queue:       NewQueue(maxQueueLen),
		maxQueueLen: maxQueueLen,
	}
}

func (l *Listener) publishBlock(ch chan<- types.Block, seq uint64, block *types.Block) {
	if l.queue == nil {
		ch <- *block

		return
	}

	expectedSeq := l.queue.SequenceNumber()
	if seq < expectedSeq {
		return
	}

	if int(seq-expectedSeq) >= l.maxQueueLen {
		for i := 0; i <= int(seq-expectedSeq)-l.maxQueueLen; i++ {
			b, _ := l.queue.Dequeue()
			if b != nil {
				ch <- *b
			}
		}
	}

	l.queue.Insert(seq, block)
	for !l.queue.Empty() {
		b, _ := l.queue.Peek()
		if b == nil {
			return
		}

		ch <- *b
		l.queue.Dequeue()
	}
}

func (l *Listener) handleNewHeader(ctx context.Context, header *types.Header) (types.Block, error) {
	var err error
	var logs []types.Log

	l.l.Debugw("Handle for new head", "hash", header.Hash)

	logs, err = getLogsByBlockHash(ctx, l.httpEVMClient, header.Hash)
	if err != nil {
		l.l.Errorw("Fail to get logs by block hash", "hash", header.Hash, "error", err)

		return types.Block{}, err
	}

	block, err := getFullBlockByHash(ctx, l.httpEVMClient, header.Hash)
	if err != nil {
		l.l.Errorw("Fail to get block by block hash", "hash", header.Hash, "error", err)

		return types.Block{}, err
	}

	l.l.Debugw("Handle new head success", "hash", header.Hash)

	return toBlock(logs, block), nil
}

func (l *Listener) getBlocks(ctx context.Context, fromBlock, toBlock uint64) ([]types.Block, error) {
	g, ctx := errgroup.WithContext(ctx)

	blocks := make([]types.Block, toBlock-fromBlock+1)
	for i := range blocks {
		i := i
		blkNum := uint64(i) + fromBlock
		g.Go(func() error {
			block, err := getBlockByNumber(ctx, l.httpEVMClient, new(big.Int).SetUint64(blkNum))
			if err != nil {
				l.l.Errorw("Fail to get block by number", "number", blkNum, "error", err)

				return err
			}

			blocks[i] = block

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (l *Listener) handleOldHeaders(ctx context.Context, blockCh chan<- types.Block) error {
	blockNumber, err := l.httpEVMClient.BlockNumber(ctx)
	if err != nil {
		l.l.Errorw("Fail to get latest block number", "error", err)

		return err
	}

	savedBlock, err := l.handler.blockKeeper.Head()
	if err != nil {
		l.l.Errorw("Fail to get last saved block", "error", err)

		return err
	}

	fromBlock := savedBlock.Number.Uint64()
	if blockNumber <= fromBlock+1 {
		l.setResuming(false)

		return nil
	}

	const batchSize = 32

	l.l.Infow("Synchronize for new headers", "fromBlock", fromBlock, "toBlock", blockNumber)
	for i := fromBlock + 1; i < blockNumber; i += batchSize {
		toBlock := i + batchSize - 1
		if toBlock >= blockNumber {
			toBlock = blockNumber - 1
		}

		l.l.Infow("Get blocks from node", "from", i, "to", toBlock)
		blocks, err := l.getBlocks(ctx, i, toBlock)
		if err != nil {
			return err
		}

		if len(blocks) > 0 {
			l.mu.Lock()
			l.lastReceivedBlock = &blocks[len(blocks)-1]
			l.mu.Unlock()
		}

		for i := range blocks {
			blockCh <- blocks[i]
		}
	}

	l.l.Infow("Finish synchronize blocks", "fromBlock", fromBlock, "toBlock", blockNumber)

	return nil
}

func (l *Listener) subscribeNewBlockHead(ctx context.Context, blockCh chan<- types.Block) error {
	l.l.Info("Start subscribing for new head of the chain")
	headerCh := make(chan *types.Header, 1)
	sub, err := l.wsEVMClient.SubscribeNewHead(ctx, headerCh)
	if err != nil {
		l.l.Errorw("Fail to subscribe new head", "error", err)

		return err
	}

	defer sub.Unsubscribe()

	err = l.handleOldHeaders(ctx, blockCh)
	if err != nil {
		l.l.Errorw("Fail to handle old headers", "error", err)

		return err
	}

	seq := uint64(1)
	for {
		select {
		case <-ctx.Done():
			l.l.Infow("Stop subscribing for new head")

			return nil
		case err = <-sub.Err():
			l.l.Errorw("Error while subscribing new head", "error", err)

			return err
		case header := <-headerCh:
			l.l.Debugw("Receive new head of the chain", "header", header)

			l.mu.Lock()
			if l.lastReceivedBlock == nil || l.lastReceivedBlock.Timestamp < header.Time {
				l.lastReceivedBlock = &types.Block{
					Number:     header.Number,
					Hash:       header.Hash,
					Timestamp:  header.Time,
					ParentHash: header.ParentHash,
				}
			}
			l.mu.Unlock()

			go func(seq uint64, head *types.Header) {
				b, err := l.handleNewHeader(ctx, head)
				if err != nil {
					l.l.Fatalw("Fail to handle new head", "header", header, "error", err)
				}

				l.publishBlock(blockCh, seq, &b)
			}(seq, header)

			seq++
		}
	}
}

func (l *Listener) syncBlocks(ctx context.Context, blockCh chan types.Block) error {
	for {
		err := l.subscribeNewBlockHead(ctx, blockCh)
		if err == nil {
			return nil
		}

		l.l.Errorw("Error occur while sync blocks", "error", err)
		if !websocket.IsCloseError(err, websocket.CloseAbnormalClosure,
			websocket.CloseNormalClosure, websocket.CloseServiceRestart) &&
			!errors.Is(err, syscall.ECONNRESET) &&
			!errors.Is(err, ethereum.NotFound) &&
			err.Error() != errStringUnknownBlock {
			return err
		}

		l.l.Infow("Re-subscribe for new block head from node")
	}
}

// Run listens for new block head and handle it.
func (l *Listener) Run(ctx context.Context) error {
	l.l.Info("Start listener service")
	defer l.l.Info("Stop listener service")

	l.l.Info("Init handler")
	err := l.handler.Init(ctx)
	if err != nil {
		l.l.Errorw("Fail to init handler", "error", err)

		return err
	}

	l.setResuming(true)

	// Start go routine for sanity checking.
	go func() {
		err := l.runSanityCheck(ctx)
		if err != nil {
			l.l.Fatalw("Sanity check failed", "error", err)
		}
	}()

	// Synchronize blocks from node.
	blockCh := make(chan types.Block, bufLen)
	go func() {
		err := l.syncBlocks(ctx, blockCh)
		if err != nil {
			l.l.Fatalw("Fail to sync blocks", "error", err)
		}

		close(blockCh)
	}()

	// Start metrics collector.
	if err := l.startMetricsCollector(ctx); err != nil {
		l.l.Errorw("Fail to start metrics collector", "error", err)

		return err
	}
	defer func() { //nolint:contextcheck
		_ = pkgmetric.Flush(context.Background())
	}()

	l.l.Info("Start handling for new blocks")
	for b := range blockCh {
		l.l.Debugw("Receive new block",
			"hash", b.Hash, "parent", b.ParentHash, "numLogs", len(b.Logs))
		err = l.handler.Handle(ctx, b)
		if err != nil {
			l.l.Errorw("Fail to handle new block", "hash", b.Hash, "error", err)

			return err
		}

		l.mu.Lock()
		l.lastHandledBlockNumber = b.Number
		l.mu.Unlock()
	}

	return nil
}

func (l *Listener) startMetricsCollector(_ context.Context) error {
	// Register callback for collecting last received block number.
	_, err := pkgmetric.Meter().Int64ObservableGauge(
		metricNameLastReceivedBlockNumber,
		metric.WithInt64Callback(func(_ context.Context, obsrv metric.Int64Observer) error {
			l.mu.Lock()
			lastReceivedBlock := l.lastReceivedBlock
			l.mu.Unlock()

			if lastReceivedBlock != nil && lastReceivedBlock.Number != nil {
				obsrv.Observe(lastReceivedBlock.Number.Int64())
			}

			return nil
		}),
	)
	if err != nil {
		l.l.Errorw("Fail to register metrics collector for last received block number", "error", err)

		return err
	}

	// Register callback for collecting last handled block number.
	_, err = pkgmetric.Meter().Int64ObservableGauge(
		metricNameLastHandledBlockNumber,
		metric.WithInt64Callback(func(_ context.Context, obsrv metric.Int64Observer) error {
			l.mu.Lock()
			lastHandledBlockNumber := l.lastHandledBlockNumber
			l.mu.Unlock()

			if lastHandledBlockNumber != nil {
				obsrv.Observe(lastHandledBlockNumber.Int64())
			}

			return nil
		}),
	)
	if err != nil {
		l.l.Errorw("Fail to register metrics collector for last handled block number", "error", err)

		return err
	}

	// Register callback for collecting last checked block number.
	_, err = pkgmetric.Meter().Int64ObservableGauge(
		metricNameLastCheckedBlockNumber,
		metric.WithInt64Callback(func(_ context.Context, obsrv metric.Int64Observer) error {
			l.mu.Lock()
			lastCheckedBlockNumber := l.lastCheckedBlockNumber
			l.mu.Unlock()

			if lastCheckedBlockNumber != nil {
				obsrv.Observe(lastCheckedBlockNumber.Int64())
			}

			return nil
		}),
	)
	if err != nil {
		l.l.Errorw("Fail to register metrics collector for last checked block number", "error", err)

		return err
	}

	return nil
}

func (l *Listener) sanityCheck(ctx context.Context, validSecond uint64) error {
	header, err := getHeaderByNumber(ctx, l.sanityEVMClient, nil)
	if err != nil {
		return err
	}

	l.mu.Lock()
	l.lastCheckedBlockNumber = header.Number
	lastBlock := l.lastReceivedBlock
	l.mu.Unlock()
	if lastBlock == nil {
		return nil
	}

	if l.isResuming() {
		// Catchup to the lastest block.
		if lastBlock.Timestamp >= header.Time-validSecond {
			l.setResuming(false)
		}

		return nil
	}

	if lastBlock.Timestamp < header.Time-validSecond {
		return errors.New("sanity check failed")
	}

	return nil
}

func (l *Listener) runSanityCheck(ctx context.Context) error {
	if l.sanityEVMClient == nil {
		return nil
	}

	intervalSecond := uint64(l.sanityCheckInterval / time.Second)
	if intervalSecond == 0 {
		intervalSecond = 1
	}

	ticker := time.NewTicker(l.sanityCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := l.sanityCheck(ctx, intervalSecond)
			if err != nil {
				return err
			}
		}
	}
}

func (l *Listener) setResuming(v bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.resuming = v
}

func (l *Listener) isResuming() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.resuming
}
