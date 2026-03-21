package domain

import "time"

// DelayRecord represents a single telemetry ping for a transit vehicle.
type DelayRecord struct {
	ID                 string    // Unique identifier for the record (e.g., UUID or DB generated)
	RouteID            string    // The transit line, e.g., "S1", "S2"
	TripID             string    // Unique ID for the specific journey
	StationID          string    // Station identifier where the delay is measured
	ExpectedArrival    time.Time // When the train was scheduled to arrive
	ActualArrival      time.Time // When the train actually arrived (or is predicted to arrive)
	DelaySeconds       int       // Derived difference between expected and actual
	IngestionTimestamp time.Time // When this record was captured by the system
}

// Metrics represents pre-calculated analytical data for a specific route.
type Metrics struct {
	RouteID            string
	PunctualityRatio   float64 // Percentage of trains arriving < 6 mins
	MedianDelaySeconds float64 // 50th percentile
	P90DelaySeconds    float64 // 90th percentile
	P95DelaySeconds    float64 // 95th percentile
	TotalTrips         int     // Total number of trips analyzed
	GeneratedAt        time.Time
}

// StationRequestLog tracks exactly what the DB API returned per station ping.
type StationRequestLog struct {
	ID                 string    // Unique log UUID
	StationID          string    // The EVA number pinged
	StatusCode         int       // HTTP Status Code (200, 400, etc)
	Timestamp          time.Time // Exact time of the request
}
