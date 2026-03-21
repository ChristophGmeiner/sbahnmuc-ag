package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/christophgmeiner/sbahnmuc-ag/internal/domain"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage initializes a new SQLite database connection and sets up WAL mode.
func NewSQLiteStorage(dsn string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable Write-Ahead Logging for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Create tables if they don't exist
	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &SQLiteStorage{db: db}, nil
}

func createSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS delay_records (
		id TEXT PRIMARY KEY,
		route_id TEXT NOT NULL,
		trip_id TEXT NOT NULL,
		station_id TEXT NOT NULL,
		expected_arrival DATETIME,
		actual_arrival DATETIME,
		delay_seconds INTEGER,
		ingestion_timestamp DATETIME
	);
	
	CREATE INDEX IF NOT EXISTS idx_route_timestamp ON delay_records(route_id, ingestion_timestamp);

	CREATE TABLE IF NOT EXISTS station_request_logs (
		id TEXT PRIMARY KEY,
		station_id TEXT NOT NULL,
		status_code INTEGER,
		timestamp DATETIME
	);

	CREATE TABLE IF NOT EXISTS stations (
		eva TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		ds100 TEXT
	);
	`
	_, err := db.Exec(query)
	return err
}

func (s *SQLiteStorage) SaveRecords(ctx context.Context, records []domain.DelayRecord, logs []domain.StationRequestLog) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Prepare Delay Records
	stmtRecords, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO delay_records 
		(id, route_id, trip_id, station_id, expected_arrival, actual_arrival, delay_seconds, ingestion_timestamp) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmtRecords.Close()

	// 2. Prepare Request Logs
	stmtLogs, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO station_request_logs 
		(id, station_id, status_code, timestamp) 
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmtLogs.Close()

	// Batch insert Delay Records
	for _, r := range records {
		_, err = stmtRecords.ExecContext(ctx,
			r.ID, r.RouteID, r.TripID, r.StationID,
			r.ExpectedArrival, r.ActualArrival, r.DelaySeconds, r.IngestionTimestamp,
		)
		if err != nil {
			return err
		}
	}

	// Batch insert Request Logs
	for _, l := range logs {
		_, err = stmtLogs.ExecContext(ctx, l.ID, l.StationID, l.StatusCode, l.Timestamp)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLiteStorage) SaveStations(ctx context.Context, stations []domain.Station) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO stations (eva, name, ds100) 
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, st := range stations {
		if _, err := stmt.ExecContext(ctx, st.EVA, st.Name, st.DS100); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLiteStorage) GetAggregatedMetrics(ctx context.Context, routeID string) (domain.Metrics, error) {
	// Simple aggregated metrics for local mode
	query := `
	SELECT 
		COUNT(*) as total_trips,
		IFNULL(SUM(CASE WHEN delay_seconds < 360 THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as punctuality_ratio,
		IFNULL(AVG(delay_seconds), 0) as avg_delay 
	FROM delay_records 
	WHERE route_id = ?
	`
	
	var totalTrips int
	var punctualityRatio, avgDelay float64

	err := s.db.QueryRowContext(ctx, query, routeID).Scan(&totalTrips, &punctualityRatio, &avgDelay)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Metrics{RouteID: routeID}, nil
		}
		return domain.Metrics{}, err
	}

	return domain.Metrics{
		RouteID:            routeID,
		TotalTrips:         totalTrips,
		PunctualityRatio:   punctualityRatio,
		MedianDelaySeconds: avgDelay, 
	}, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
