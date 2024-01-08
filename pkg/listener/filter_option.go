package listener

type Option func(opt *FilterOption)

type FilterOption struct {
	filterContracts []string
	filterTopics    [][]string
	withLogs        bool
}

func WithEventLogs(contracts []string, topics [][]string) Option {
	return func(opt *FilterOption) {
		opt.withLogs = true
		opt.filterContracts = contracts
		opt.filterTopics = topics
	}
}
