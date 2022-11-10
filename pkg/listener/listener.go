package listener

import (
	"sync"

	"github.com/KyberNetwork/bclistener/pkg/block"
	"github.com/KyberNetwork/bclistener/pkg/listener/job"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// Config contains configuration options for listener service.
type Config struct {
	MaxNumBlocks    int
	FilterAddresses map[common.Address]bool // nil for skip filtering.
}

// Listener represents a listener service for on-chain events.
type Listener struct {
	mu              sync.Mutex              //nolint:unused
	maxNumBlocks    int                     //nolint:unused
	filterAddresses map[common.Address]bool //nolint:unused

	ethClient    *ethclient.Client  //nolint:unused
	jobSplitter  job.Splitter       //nolint:unused
	jobPublisher job.Publisher      //nolint:unused
	bs           block.Storage      //nolint:unused
	l            *zap.SugaredLogger //nolint:unused
}
