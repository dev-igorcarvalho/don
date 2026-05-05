package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Run("WithWriter", func(t *testing.T) {
		cfg := Config{Driver: "postgres", DSN: "writer-dsn"}
		opts := &options{}

		optFunc := WithWriter(cfg)
		optFunc(opts)

		assert.NotNil(t, opts.writer)
		assert.Equal(t, cfg.Driver, opts.writer.Driver)
		assert.Equal(t, cfg.DSN, opts.writer.DSN)
	})

	t.Run("WithReader", func(t *testing.T) {
		cfg := Config{Driver: "postgres", DSN: "reader-dsn"}
		opts := &options{}

		optFunc := WithReader(cfg)
		optFunc(opts)

		assert.NotNil(t, opts.reader)
		assert.Equal(t, cfg.Driver, opts.reader.Driver)
		assert.Equal(t, cfg.DSN, opts.reader.DSN)
	})

	t.Run("Multiple options", func(t *testing.T) {
		writerCfg := Config{Driver: "postgres", DSN: "writer"}
		readerCfg := Config{Driver: "postgres", DSN: "reader"}
		opts := &options{}

		WithWriter(writerCfg)(opts)
		WithReader(readerCfg)(opts)

		assert.NotNil(t, opts.writer)
		assert.NotNil(t, opts.reader)
		assert.Equal(t, "writer", opts.writer.DSN)
		assert.Equal(t, "reader", opts.reader.DSN)
	})

	t.Run("Overwrite options", func(t *testing.T) {
		cfg1 := Config{DSN: "dsn1"}
		cfg2 := Config{DSN: "dsn2"}
		opts := &options{}

		WithWriter(cfg1)(opts)
		assert.Equal(t, "dsn1", opts.writer.DSN)

		WithWriter(cfg2)(opts)
		assert.Equal(t, "dsn2", opts.writer.DSN)
	})
}
