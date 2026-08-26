package mon

type (
	DatabaseOption func(opts *DatabaseOptions)

	DatabaseOptions struct {
		Uri      string
		Database string
		Options  []Option
		Indexes  []*Index
	}
)

func NewDatabaseOptions(options ...DatabaseOption) *DatabaseOptions {
	opts := &DatabaseOptions{}
	for _, opt := range options {
		opt(opts)
	}

	return opts
}

func WithIndex(idx ...*Index) DatabaseOption {
	return func(opts *DatabaseOptions) {
		opts.Indexes = idx
	}
}

func WithClientOptions(options ...Option) DatabaseOption {
	return func(opts *DatabaseOptions) {
		opts.Options = options
	}
}

func WithUri(uri string) DatabaseOption {
	return func(opts *DatabaseOptions) {
		opts.Uri = uri
	}
}

func WithDatabase(database string) DatabaseOption {
	return func(opts *DatabaseOptions) {
		opts.Database = database
	}
}
