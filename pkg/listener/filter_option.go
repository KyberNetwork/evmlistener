package listener

type Option func(opt *FilterOption)

type FilterOption struct {
	filterContracts  []string
	filterTopics     [][]string
	withLogs         bool
	spamLogThreshold int
}

func WithEventLogs(contracts []string, topics [][]string) Option {
	return func(opt *FilterOption) {
		opt.withLogs = true
		opt.filterContracts = contracts
		opt.filterTopics = topics
	}
}

// WithSpamLogThreshold enables dropping all but the last log per (address, topic0) pair when a
// block contains more than threshold occurrences. 0 disables the filter.
func WithSpamLogThreshold(threshold int) Option {
	return func(opt *FilterOption) {
		opt.spamLogThreshold = threshold
	}
}
