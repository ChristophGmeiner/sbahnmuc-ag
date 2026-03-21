package domain

import "context"

// TransitProvider defines the interface for fetching real-time transit data.
// This decouples the engine from specific external APIs (e.g., DB API, GTFS-RT).
type TransitProvider interface {
	// FetchDelays retrieves the current delay records for the target network.
	// It is expected to handle its own network requests, parsing, and filtering.
	FetchDelays(ctx context.Context) ([]DelayRecord, []StationRequestLog, error)

	// FetchStations retrieves station master data for the configured stations.
	FetchStations(ctx context.Context) ([]Station, error)
}

// StorageProvider defines the interface for persisting data and retrieving metrics.
// This decouples the engine from specific databases (e.g., SQLite, BigQuery).
type StorageProvider interface {
	// SaveRecords persists a batch of delay records and their associated API request logs.
	SaveRecords(ctx context.Context, records []DelayRecord, logs []StationRequestLog) error

	// SaveStations persists station master data to the local store.
	SaveStations(ctx context.Context, stations []Station) error

	// GetAggregatedMetrics computes and retrieves statistical metrics for a route.
	// For local deployments, this may compute on the fly. For cloud, it may query pre-aggregated views.
	GetAggregatedMetrics(ctx context.Context, routeID string) (Metrics, error)
	
	// Close safely closes the underlying database connection.
	Close() error
}
