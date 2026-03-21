package engine

import (
	"context"
	"testing"
	"time"

	"github.com/christophgmeiner/sbahnmuc-ag/internal/domain"
)

type mockTransit struct {
	records []domain.DelayRecord
	err     error
	calls   int
}

func (m *mockTransit) FetchDelays(ctx context.Context) ([]domain.DelayRecord, error) {
	m.calls++
	return m.records, m.err
}

type mockStorage struct {
	saved []domain.DelayRecord
	err   error
	calls int
}

func (m *mockStorage) SaveRecords(ctx context.Context, records []domain.DelayRecord) error {
	m.calls++
	m.saved = append(m.saved, records...)
	return m.err
}

func (m *mockStorage) GetAggregatedMetrics(ctx context.Context, routeID string) (domain.Metrics, error) {
	return domain.Metrics{}, nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestEnginePolling(t *testing.T) {
	transit := &mockTransit{
		records: []domain.DelayRecord{
			{RouteID: "S1", DelaySeconds: 60},
		},
	}
	storage := &mockStorage{}

	eng := NewEngine(transit, storage)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eng.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	eng.Stop()

	if transit.calls != 1 {
		t.Errorf("expected 1 fetch call, got %d", transit.calls)
	}

	if storage.calls != 1 {
		t.Errorf("expected 1 save call, got %d", storage.calls)
	}

	if len(storage.saved) != 1 {
		t.Errorf("expected 1 saved record, got %d", len(storage.saved))
	}
}
