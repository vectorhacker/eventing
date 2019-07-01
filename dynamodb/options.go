package dynamodb

// StoreOption applies an optional configuration to the Store
type StoreOption func(s *store)

// WithRangeKey sets the range key
func WithRangeKey(rangeKey string) StoreOption {
	return func(s *store) {
		s.rangeKey = rangeKey
	}
}

// WithHashKey sets the hash key
func WithHashKey(hashKey string) StoreOption {
	return func(s *store) {
		s.hashKey = hashKey
	}
}

// WithRecordsPerPartition sets the number of records in a partition
func WithRecordsPerPartition(records int) StoreOption {
	return func(s *store) {
		s.itemsPerPartition = records
	}
}
