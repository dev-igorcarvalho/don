// ---
// title: SQL Connection Pair
// description: Manages a pair of SQL database connections to support read/write splitting and health monitoring.
// last_updated: 2026-05-05
// type: Implementation
// ---

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Client manages a pair of database connections for read/write splitting.
type Client struct {
	writer *sql.DB
	reader *sql.DB
}

// Writer returns the writer database connection.
func (p *Client) Writer() *sql.DB {
	return p.writer
}

// Reader returns the reader database connection.
func (p *Client) Reader() *sql.DB {
	return p.reader
}

// Stats holds aggregated statistics for the database connections.
type Stats struct {
	Writer sql.DBStats
	Reader sql.DBStats
}

// HealthStatus represents the aggregated health of the Client.
type HealthStatus struct {
	WriterAlive bool
	ReaderAlive bool
	OpenConns   int
	IdleConns   int
	Message     string
}

// NewClient creates a new Client using the provided writer and reader configurations.
func NewClient(ctx context.Context, writerCfg, readerCfg Config) (*Client, error) {
	writer, err := newSQL(ctx, writerCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize writer: %w", err)
	}

	reader, err := newSQL(ctx, readerCfg)
	if err != nil {
		_ = writer.Close()
		return nil, fmt.Errorf("failed to initialize reader: %w", err)
	}

	return &Client{
		writer: writer,
		reader: reader,
	}, nil
}

// Ping verifies the connectivity to both writer and reader databases.
func (p *Client) Ping(ctx context.Context) error {
	var errs []error

	if p.writer != nil {
		if err := p.writer.PingContext(ctx); err != nil {
			errs = append(errs, fmt.Errorf("writer ping failed: %w", err))
		}
	}

	if p.reader != nil {
		if err := p.reader.PingContext(ctx); err != nil {
			errs = append(errs, fmt.Errorf("reader ping failed: %w", err))
		}
	}

	return errors.Join(errs...)
}

// HealthCheck performs a ping on both databases and aggregates the results and stats.
func (p *Client) HealthCheck(ctx context.Context) HealthStatus {
	status := HealthStatus{}
	var errs []error

	if p.writer != nil {
		if err := p.writer.PingContext(ctx); err == nil {
			status.WriterAlive = true
		} else {
			errs = append(errs, fmt.Errorf("writer: %w", err))
		}
		stats := p.writer.Stats()
		status.OpenConns += stats.OpenConnections
		status.IdleConns += stats.Idle
	}

	if p.reader != nil {
		if err := p.reader.PingContext(ctx); err == nil {
			status.ReaderAlive = true
		} else {
			errs = append(errs, fmt.Errorf("reader: %w", err))
		}
		stats := p.reader.Stats()
		status.OpenConns += stats.OpenConnections
		status.IdleConns += stats.Idle
	}

	if len(errs) > 0 {
		status.Message = errors.Join(errs...).Error()
	} else {
		status.Message = "OK"
	}

	return status
}

// Stats returns the aggregated statistics for both writer and reader connections.
func (p *Client) Stats() Stats {
	var stats Stats
	if p.writer != nil {
		stats.Writer = p.writer.Stats()
	}
	if p.reader != nil {
		stats.Reader = p.reader.Stats()
	}
	return stats
}

// Close closes both writer and reader connections.
func (p *Client) Close() error {
	var errs []error
	if p.writer != nil {
		if err := p.writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close writer: %w", err))
		}
		p.writer = nil
	}
	if p.reader != nil {
		if err := p.reader.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close reader: %w", err))
		}
		p.reader = nil
	}

	return errors.Join(errs...)
}

// Shutdown gracefully shuts down the database connections.
func (p *Client) Shutdown(_ context.Context) error {
	return p.Close()
}
