package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/christophgmeiner/sbahnmuc-ag/internal/domain"
)

func TestSQLiteStorage(t *testing.T) {
	// Use memory database for quick tests or a temp file to test WAL
	testDB := "./test_storage.db"
	defer os.Remove(testDB)

	storage, err := NewSQLiteStorage(testDB)
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	now := time.Now().Truncate(time.Second)

	records := []domain.DelayRecord{
		{
			ID:                 "rec-1",
			RouteID:            "S1",
			TripID:             "trip-1",
			StationID:          "stat-1",
			ExpectedArrival:    now.Add(-10 * time.Minute),
			ActualArrival:      now.Add(-8 * time.Minute),
			DelaySeconds:       120, // < 6 mins (on-time by DB standards)
			IngestionTimestamp: now,
		},
		{
			ID:                 "rec-2",
			RouteID:            "S1",
			TripID:             "trip-2",
			StationID:          "stat-2",
			ExpectedArrival:    now.Add(-20 * time.Minute),
			ActualArrival:      now.Add(-10 * time.Minute),
			DelaySeconds:       600, // > 6 mins (delayed)
			IngestionTimestamp: now,
		},
	}

	// Test SaveRecords
	err = storage.SaveRecords(ctx, records)
	if err != nil {
		t.Fatalf("SaveRecords failed: %v", err)
	}

	// Test GetAggregatedMetrics
	metrics, err := storage.GetAggregatedMetrics(ctx, "S1")
	if err != nil {
		t.Fatalf("GetAggregatedMetrics failed: %v", err)
	}

	if metrics.TotalTrips != 2 {
		t.Errorf("expected 2 trips, got %d", metrics.TotalTrips)
	}

	// 1 out of 2 is on time (< 360 seconds) -> 50%
	if metrics.PunctualityRatio != 50.0 {
		t.Errorf("expected 50.0 punctuality, got %f", metrics.PunctualityRatio)
	}

	// Avg delay = (120 + 600) / 2 = 360
	if metrics.MedianDelaySeconds != 360 {
		t.Errorf("expected 360 avg delay, got %f", metrics.MedianDelaySeconds)
	}
}
