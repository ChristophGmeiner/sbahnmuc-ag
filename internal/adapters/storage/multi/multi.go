package multi

import (
	"context"
	"errors"

	"github.com/christophgmeiner/sbahnmuc-ag/internal/domain"
)

type MultiStorage struct {
	providers []domain.StorageProvider
}

// Ensure interface is implemented
var _ domain.StorageProvider = (*MultiStorage)(nil)

func NewMultiStorage(providers ...domain.StorageProvider) *MultiStorage {
	return &MultiStorage{
		providers: providers,
	}
}

func (m *MultiStorage) SaveRecords(ctx context.Context, records []domain.DelayRecord, logs []domain.StationRequestLog) error {
	var errs []error
	for _, p := range m.providers {
		if err := p.SaveRecords(ctx, records, logs); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (m *MultiStorage) SaveStations(ctx context.Context, stations []domain.Station) error {
	var errs []error
	for _, p := range m.providers {
		if err := p.SaveStations(ctx, stations); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (m *MultiStorage) GetAggregatedMetrics(ctx context.Context, routeID string) (domain.Metrics, error) {
	if len(m.providers) == 0 {
		return domain.Metrics{}, errors.New("no storage providers configured")
	}
	
	// Default to returning the metrics from the first provider
	// In practice, this might be BigQuery if it's the primary analytical db,
	// or SQLite if running fully locally. The order is determined during setup.
	return m.providers[0].GetAggregatedMetrics(ctx, routeID)
}

func (m *MultiStorage) Close() error {
	var errs []error
	for _, p := range m.providers {
		if err := p.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
