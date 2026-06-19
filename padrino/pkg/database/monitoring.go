// ---
// title: Database Monitoring
// description: Provides health checks, connectivity pings, and statistics for the database client.
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
