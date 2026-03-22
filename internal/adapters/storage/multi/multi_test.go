package multi_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/christophgmeiner/sbahnmuc-ag/internal/adapters/storage/multi"
	"github.com/christophgmeiner/sbahnmuc-ag/internal/domain"
)

type mockProvider struct {
	SaveRecordsErr    error
	SaveStationsErr   error
	Metrics           domain.Metrics
	MetricsErr        error
	CloseErr          error
	SaveRecordsCalled int
}

func (m *mockProvider) SaveRecords(ctx context.Context, records []domain.DelayRecord, logs []domain.StationRequestLog) error {
	m.SaveRecordsCalled++
	return m.SaveRecordsErr
}

func (m *mockProvider) SaveStations(ctx context.Context, stations []domain.Station) error {
	return m.SaveStationsErr
}

func (m *mockProvider) GetAggregatedMetrics(ctx context.Context, routeID string) (domain.Metrics, error) {
	return m.Metrics, m.MetricsErr
}

func (m *mockProvider) Close() error {
	return m.CloseErr
}

func TestMultiStorage_SaveRecords(t *testing.T) {
	ctx := context.Background()

	p1 := &mockProvider{}
	p2 := &mockProvider{SaveRecordsErr: errors.New("provider 2 error")}
	p3 := &mockProvider{}

	ms := multi.NewMultiStorage(p1, p2, p3)

	err := ms.SaveRecords(ctx, nil, nil)

	// Verify all providers were called despite the error.
	if p1.SaveRecordsCalled != 1 || p2.SaveRecordsCalled != 1 || p3.SaveRecordsCalled != 1 {
		t.Errorf("Expected SaveRecords to be called 1 time on each provider")
	}

	// Verify the error is wrapped
	if err == nil || !strings.Contains(err.Error(), "provider 2 error") {
		t.Errorf("Expected combined error containing 'provider 2 error', got %v", err)
	}
}

func TestMultiStorage_GetAggregatedMetrics(t *testing.T) {
	ctx := context.Background()

	p1 := &mockProvider{
		Metrics: domain.Metrics{RouteID: "S1", TotalTrips: 10},
	}
	p2 := &mockProvider{
		Metrics: domain.Metrics{RouteID: "S2", TotalTrips: 5},
	}

	ms := multi.NewMultiStorage(p1, p2)

	metrics, err := ms.GetAggregatedMetrics(ctx, "S1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should return metrics from the very first provider
	if metrics.TotalTrips != 10 {
		t.Errorf("Expected metrics from first provider (10 trips), got %d", metrics.TotalTrips)
	}
}

func TestMultiStorage_Empty(t *testing.T) {
	ms := multi.NewMultiStorage()
	_, err := ms.GetAggregatedMetrics(context.Background(), "S1")
	if err == nil {
		t.Errorf("Expected error from GetAggregatedMetrics when empty")
	}
}
