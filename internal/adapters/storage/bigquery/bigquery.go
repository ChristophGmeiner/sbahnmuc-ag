package bigquery

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/christophgmeiner/sbahnmuc-ag/internal/domain"
	"google.golang.org/api/iterator"
)

type BigQueryStorage struct {
	client    *bigquery.Client
	datasetID string
}

// Ensure interface is implemented
var _ domain.StorageProvider = (*BigQueryStorage)(nil)

func NewBigQueryStorage(ctx context.Context, projectID, datasetID string) (*BigQueryStorage, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}

	storage := &BigQueryStorage{
		client:    client,
		datasetID: datasetID,
	}

	if err := storage.initSchema(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize bigquery schema: %w", err)
	}

	return storage, nil
}

func (s *BigQueryStorage) initSchema(ctx context.Context) error {
	dataset := s.client.Dataset(s.datasetID)
	_, err := dataset.Metadata(ctx)
	if err != nil {
		// Attempt to create dataset if it doesn't exist
		log.Printf("[BigQuery] Auto-creating missing dataset: %s", s.datasetID)
		if err := dataset.Create(ctx, &bigquery.DatasetMetadata{Location: "EU"}); err != nil {
			return fmt.Errorf("could not create dataset %s: %w", s.datasetID, err)
		}
	}

	// 1. delay_records table (partitioned by ingestion_timestamp, clustered by route_id, station_id)
	delayRecordsTable := dataset.Table("delay_records")
	_, err = delayRecordsTable.Metadata(ctx)
	if err != nil {
		meta := &bigquery.TableMetadata{
			Schema: bigquery.Schema{
				&bigquery.FieldSchema{Name: "id", Type: bigquery.StringFieldType, Required: true},
				&bigquery.FieldSchema{Name: "route_id", Type: bigquery.StringFieldType, Required: true},
				&bigquery.FieldSchema{Name: "trip_id", Type: bigquery.StringFieldType, Required: true},
				&bigquery.FieldSchema{Name: "station_id", Type: bigquery.StringFieldType, Required: true},
				&bigquery.FieldSchema{Name: "expected_arrival", Type: bigquery.TimestampFieldType},
				&bigquery.FieldSchema{Name: "actual_arrival", Type: bigquery.TimestampFieldType},
				&bigquery.FieldSchema{Name: "delay_seconds", Type: bigquery.IntegerFieldType},
				&bigquery.FieldSchema{Name: "ingestion_timestamp", Type: bigquery.TimestampFieldType, Required: true},
			},
			TimePartitioning: &bigquery.TimePartitioning{
				Type:  bigquery.DayPartitioningType,
				Field: "expected_arrival",
			},
			Clustering: &bigquery.Clustering{
				Fields: []string{"route_id", "station_id"},
			},
		}
		if err := delayRecordsTable.Create(ctx, meta); err != nil {
			return fmt.Errorf("could not create delay_records table: %w", err)
		}
		log.Printf("[BigQuery] Auto-created missing table: delay_records")
	}

	// 2. station_request_logs table (partitioned by timestamp, clustered by station_id)
	stationLogsTable := dataset.Table("station_request_logs")
	_, err = stationLogsTable.Metadata(ctx)
	if err != nil {
		meta := &bigquery.TableMetadata{
			Schema: bigquery.Schema{
				&bigquery.FieldSchema{Name: "id", Type: bigquery.StringFieldType, Required: true},
				&bigquery.FieldSchema{Name: "station_id", Type: bigquery.StringFieldType, Required: true},
				&bigquery.FieldSchema{Name: "status_code", Type: bigquery.IntegerFieldType},
				&bigquery.FieldSchema{Name: "timestamp", Type: bigquery.TimestampFieldType, Required: true},
			},
			TimePartitioning: &bigquery.TimePartitioning{
				Type:  bigquery.DayPartitioningType,
				Field: "timestamp",
			},
			Clustering: &bigquery.Clustering{
				Fields: []string{"station_id"},
			},
		}
		if err := stationLogsTable.Create(ctx, meta); err != nil {
			return fmt.Errorf("could not create station_request_logs table: %w", err)
		}
		log.Printf("[BigQuery] Auto-created missing table: station_request_logs")
	}

	// 3. stations table (dimension table)
	stationsTable := dataset.Table("stations")
	_, err = stationsTable.Metadata(ctx)
	if err != nil {
		meta := &bigquery.TableMetadata{
			Schema: bigquery.Schema{
				&bigquery.FieldSchema{Name: "eva", Type: bigquery.StringFieldType, Required: true},
				&bigquery.FieldSchema{Name: "name", Type: bigquery.StringFieldType, Required: true},
				&bigquery.FieldSchema{Name: "ds100", Type: bigquery.StringFieldType},
			},
		}
		if err := stationsTable.Create(ctx, meta); err != nil {
			return fmt.Errorf("could not create stations table: %w", err)
		}
		log.Printf("[BigQuery] Auto-created missing table: stations")
	}

	return nil
}

type delayRecordRow struct {
	ID                 string                 `bigquery:"id"`
	RouteID            string                 `bigquery:"route_id"`
	TripID             string                 `bigquery:"trip_id"`
	StationID          string                 `bigquery:"station_id"`
	ExpectedArrival    bigquery.NullTimestamp `bigquery:"expected_arrival"`
	ActualArrival      bigquery.NullTimestamp `bigquery:"actual_arrival"`
	DelaySeconds       bigquery.NullInt64     `bigquery:"delay_seconds"`
	IngestionTimestamp bigquery.NullTimestamp `bigquery:"ingestion_timestamp"`
}

type stationLogRow struct {
	ID         string                 `bigquery:"id"`
	StationID  string                 `bigquery:"station_id"`
	StatusCode bigquery.NullInt64     `bigquery:"status_code"`
	Timestamp  bigquery.NullTimestamp `bigquery:"timestamp"`
}

type stationRow struct {
	EVA   string `bigquery:"eva"`
	Name  string `bigquery:"name"`
	DS100 string `bigquery:"ds100"`
}

func (s *BigQueryStorage) SaveRecords(ctx context.Context, records []domain.DelayRecord, logs []domain.StationRequestLog) error {
	dataset := s.client.Dataset(s.datasetID)

	// Insert Delay Records
	if len(records) > 0 {
		var delayRows []delayRecordRow
		for _, r := range records {
			delayRows = append(delayRows, delayRecordRow{
				ID:                 r.ID,
				RouteID:            r.RouteID,
				TripID:             r.TripID,
				StationID:          r.StationID,
				ExpectedArrival:    bigquery.NullTimestamp{Timestamp: r.ExpectedArrival, Valid: !r.ExpectedArrival.IsZero()},
				ActualArrival:      bigquery.NullTimestamp{Timestamp: r.ActualArrival, Valid: !r.ActualArrival.IsZero()},
				DelaySeconds:       bigquery.NullInt64{Int64: int64(r.DelaySeconds), Valid: true},
				IngestionTimestamp: bigquery.NullTimestamp{Timestamp: r.IngestionTimestamp, Valid: !r.IngestionTimestamp.IsZero()},
			})
		}
		inserter := dataset.Table("delay_records").Inserter()
		if err := inserter.Put(ctx, delayRows); err != nil {
			if strings.Contains(err.Error(), "notFound") {
				return fmt.Errorf("failed to insert delay records to BQ: %w (Note: BigQuery Streaming API takes a few minutes to recognize newly created tables. The next poll should succeed.)", err)
			}
			return fmt.Errorf("failed to insert delay records to BQ: %w", err)
		}
	}

	// Insert Request Logs
	if len(logs) > 0 {
		var logRows []stationLogRow
		for _, l := range logs {
			logRows = append(logRows, stationLogRow{
				ID:         l.ID,
				StationID:  l.StationID,
				StatusCode: bigquery.NullInt64{Int64: int64(l.StatusCode), Valid: true},
				Timestamp:  bigquery.NullTimestamp{Timestamp: l.Timestamp, Valid: !l.Timestamp.IsZero()},
			})
		}
		inserter := dataset.Table("station_request_logs").Inserter()
		if err := inserter.Put(ctx, logRows); err != nil {
			if strings.Contains(err.Error(), "notFound") {
				return fmt.Errorf("failed to insert station logs to BQ: %w (Note: BigQuery Streaming API takes a few minutes to recognize newly created tables. The next poll should succeed.)", err)
			}
			return fmt.Errorf("failed to insert station logs to BQ: %w", err)
		}
	}

	return nil
}

func (s *BigQueryStorage) SaveStations(ctx context.Context, stations []domain.Station) error {
	if len(stations) == 0 {
		return nil
	}

	// To handle upserts cleanly in a dimension table vs append-only streaming,
	// we use a standard simple MERGE technique via DML query.
	// We'll prepare the data as parameters or simply batch insert to a temp and merge.
	// For simplicity and since station counts are low, we can formulate an INSERT statement or just stream it.
	// Wait, DBAPI fetch happens frequently. Stations are mostly static.
	// The SQLite code does INSERT OR REPLACE. We can do a basic MERGE.

	// Create a query building strings or use a temp table approach.
	// Given 28 stations, a parameterized MERGE query is very simple.
	queryStr := "MERGE `" + s.client.Project() + "." + s.datasetID + ".stations` T USING ( "
	// build source values
	for i := range stations {
		if i > 0 {
			queryStr += " UNION ALL "
		}
		queryStr += fmt.Sprintf("SELECT @eva%d AS eva, @name%d AS name, @ds100%d AS ds100", i, i, i)
	}
	queryStr += ` ) S ON T.eva = S.eva 
		WHEN MATCHED THEN UPDATE SET name = S.name, ds100 = S.ds100
		WHEN NOT MATCHED THEN INSERT (eva, name, ds100) VALUES(S.eva, S.name, S.ds100)`

	q := s.client.Query(queryStr)
	for i, st := range stations {
		q.Parameters = append(q.Parameters,
			bigquery.QueryParameter{Name: fmt.Sprintf("eva%d", i), Value: st.EVA},
			bigquery.QueryParameter{Name: fmt.Sprintf("name%d", i), Value: st.Name},
			bigquery.QueryParameter{Name: fmt.Sprintf("ds100%d", i), Value: st.DS100},
		)
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to start stations merge job: %w", err)
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for stations merge job: %w", err)
	}
	if status.Err() != nil {
		return fmt.Errorf("stations merge job failed: %w", status.Err())
	}

	return nil
}

func (s *BigQueryStorage) GetAggregatedMetrics(ctx context.Context, routeID string) (domain.Metrics, error) {
	queryStr := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_trips,
			IFNULL(SUM(CASE WHEN delay_seconds < 360 THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as punctuality_ratio,
			IFNULL(APPROX_QUANTILES(delay_seconds, 100)[OFFSET(50)], 0) as median_delay,
			IFNULL(APPROX_QUANTILES(delay_seconds, 100)[OFFSET(90)], 0) as p90_delay,
			IFNULL(APPROX_QUANTILES(delay_seconds, 100)[OFFSET(95)], 0) as p95_delay
		FROM `+"`%s.%s.delay_records`"+` 
		WHERE route_id = @route_id
	`, s.client.Project(), s.datasetID)

	q := s.client.Query(queryStr)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "route_id", Value: routeID},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return domain.Metrics{}, fmt.Errorf("failed to run metrics query: %w", err)
	}

	type metricRow struct {
		TotalTrips       int64   `bigquery:"total_trips"`
		PunctualityRatio float64 `bigquery:"punctuality_ratio"`
		MedianDelay      int64   `bigquery:"median_delay"`
		P90Delay         int64   `bigquery:"p90_delay"`
		P95Delay         int64   `bigquery:"p95_delay"`
	}

	var row metricRow
	err = it.Next(&row)
	if err == iterator.Done {
		return domain.Metrics{RouteID: routeID}, nil
	}
	if err != nil {
		return domain.Metrics{}, fmt.Errorf("failed to iterate metrics result: %w", err)
	}

	return domain.Metrics{
		RouteID:            routeID,
		TotalTrips:         int(row.TotalTrips),
		PunctualityRatio:   row.PunctualityRatio,
		MedianDelaySeconds: float64(row.MedianDelay),
		P90DelaySeconds:    float64(row.P90Delay),
		P95DelaySeconds:    float64(row.P95Delay),
	}, nil
}

func (s *BigQueryStorage) Close() error {
	return s.client.Close()
}
