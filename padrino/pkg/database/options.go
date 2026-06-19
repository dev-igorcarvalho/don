// ---
// title: Database Client Options
// description: Defines functional options for configuring the database client.
// last_updated: 2026-05-05
// type: Implementation
// ---

package database

// options holds the configuration for the database client.
type options struct {
	writer *Config
	reader *Config
}

// Option defines a functional configuration for the database client.
type Option func(*options)

// WithWriter configures the Client with a writer connection.
func WithWriter(cfg Config) Option {
	return func(o *options) {
		o.writer = &cfg
	}
}

// WithReader configures the Client with a reader connection (optional).
func WithReader(cfg Config) Option {
	return func(o *options) {
		o.reader = &cfg
	}
}
