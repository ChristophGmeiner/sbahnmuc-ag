package dbapi

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/christophgmeiner/sbahnmuc-ag/internal/domain"
)

type Provider struct {
	client       *http.Client
	baseURL      string
	clientID     string
	clientSecret string
	stations     []string
}

func NewProvider(baseURL, clientID, clientSecret string, stations []string) *Provider {
	return &Provider{
		client:       &http.Client{Timeout: 10 * time.Second},
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		stations:     stations,
	}
}

type Timetable struct {
	XMLName xml.Name `xml:"timetable"`
	Station string   `xml:"station,attr"`
	Stops   []Stop   `xml:"s"`
}

type Stop struct {
	ID        string     `xml:"id,attr"`
	TrainLine *TrainLine `xml:"tl"`
	Arrival   *Event     `xml:"ar"`
	Departure *Event     `xml:"dp"`
}

type TrainLine struct {
	Category string `xml:"c,attr"`
	Number   string `xml:"n,attr"`
}

type Event struct {
	PlannedTime string `xml:"pt,attr"` // Format: YYMMDDHHmm
	ChangedTime string `xml:"ct,attr"`
	Line        string `xml:"l,attr"`
}

type StationsResult struct {
	XMLName xml.Name   `xml:"stations"`
	Station []xmlStation `xml:"station"`
}

type xmlStation struct {
	Name  string `xml:"name,attr"`
	EVA   string `xml:"eva,attr"`
	DS100 string `xml:"ds100,attr"`
}

func parseDBTime(t string) (time.Time, error) {
	if t == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	return time.Parse("0601021504", t)
}

func (p *Provider) FetchDelays(ctx context.Context) ([]domain.DelayRecord, []domain.StationRequestLog, error) {
	var allRecords []domain.DelayRecord
	var allLogs []domain.StationRequestLog
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, len(p.stations))

	// Limit concurrency to 5 parallel requests to avoid hitting rate limits
	sem := make(chan struct{}, 5)

	for _, eva := range p.stations {
		wg.Add(1)
		go func(evaNo string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			records, reqLog, err := p.fetchStation(ctx, evaNo)
			
			mu.Lock()
			if reqLog != nil {
				allLogs = append(allLogs, *reqLog)
			}
			if err != nil {
				errCh <- fmt.Errorf("station %s: %w", evaNo, err)
			} else {
				allRecords = append(allRecords, records...)
			}
			mu.Unlock()
		}(eva)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		log.Printf("Fetch warning: %v\n", err)
	}

	// We intentionally unconditionally return nil for the error.
	// If the API returns 400 for a broken station, we simply log the warning above. 
	// We MUST NOT fail the entire batch, even if all valid stations currently have 0 active delays!
	return allRecords, allLogs, nil
}

func (p *Provider) fetchStation(ctx context.Context, evaNo string) ([]domain.DelayRecord, *domain.StationRequestLog, error) {
	url := fmt.Sprintf("%s/v1/fchg/%s", p.baseURL, evaNo)
	log.Printf("[TransitProvider] Querying DB API for station %s...\n", evaNo)

	reqLog := &domain.StationRequestLog{
		ID:        fmt.Sprintf("log-%s-%d", evaNo, time.Now().UnixNano()),
		StationID: evaNo,
		Timestamp: time.Now(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		reqLog.StatusCode = 0
		return nil, reqLog, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("DB-Client-Id", p.clientID)
	req.Header.Set("DB-Api-Key", p.clientSecret)
	req.Header.Set("Accept", "application/xml")

	resp, err := p.client.Do(req)
	if err != nil {
		reqLog.StatusCode = 0
		return nil, reqLog, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	reqLog.StatusCode = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		return nil, reqLog, fmt.Errorf("API returned %d", resp.StatusCode)
	}

	var tt Timetable
	if err := xml.NewDecoder(resp.Body).Decode(&tt); err != nil {
		return nil, reqLog, fmt.Errorf("failed to decode XML: %w", err)
	}

	var records []domain.DelayRecord
	now := time.Now()

	for _, stop := range tt.Stops {
		var mainEvent *Event
		if stop.Arrival != nil && stop.Arrival.PlannedTime != "" {
			mainEvent = stop.Arrival
		} else if stop.Departure != nil && stop.Departure.PlannedTime != "" {
			mainEvent = stop.Departure
		} else {
			continue // No valid arrival or departure found
		}

		expected, err := parseDBTime(mainEvent.PlannedTime)
		if err != nil {
			log.Printf("[TransitProvider] Warning: invalid planned time '%s' for trip %s\n", mainEvent.PlannedTime, stop.ID)
			continue
		}

		actual := expected // fallback
		if mainEvent.ChangedTime != "" {
			if parsedActual, err := parseDBTime(mainEvent.ChangedTime); err == nil {
				actual = parsedActual
			}
		}

		delaySecs := int(actual.Sub(expected).Seconds())
		if delaySecs < 0 {
			delaySecs = 0
		}

		line := ""
		if stop.Arrival != nil && stop.Arrival.Line != "" {
			line = stop.Arrival.Line
		} else if stop.Departure != nil && stop.Departure.Line != "" {
			line = stop.Departure.Line
		} else if stop.TrainLine != nil {
			line = stop.TrainLine.Category + stop.TrainLine.Number
		}

		if line == "" {
			line = "Unknown"
		}

		// Strictly filter for S-Bahn trains. 
		// This permanently drops any "RE", "ICE", "RB", or "Unknown" records 
		// from being written into your SQLite local datastore.
		if !strings.HasPrefix(line, "S") {
			continue
		}

		records = append(records, domain.DelayRecord{
			ID:                 stop.ID,
			RouteID:            line,
			TripID:             fmt.Sprintf("trip-%s", stop.ID),
			StationID:          tt.Station,
			ExpectedArrival:    expected,
			ActualArrival:      actual,
			DelaySeconds:       delaySecs,
			IngestionTimestamp: now,
		})
	}

	return records, reqLog, nil
}

func (p *Provider) FetchStations(ctx context.Context) ([]domain.Station, error) {
	log.Printf("[TransitProvider] Querying DB API for station master data for %d configured stations...\n", len(p.stations))

	var stations []domain.Station
	for _, eva := range p.stations {
		url := fmt.Sprintf("%s/v1/station/%s", p.baseURL, eva)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			log.Printf("Warning: failed to create request for %s: %v\n", eva, err)
			continue
		}

		req.Header.Set("DB-Client-Id", p.clientID)
		req.Header.Set("DB-Api-Key", p.clientSecret)
		req.Header.Set("Accept", "application/xml")

		resp, err := p.client.Do(req)
		if err != nil {
			log.Printf("Warning: failed to execute request for %s: %v\n", eva, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Warning: API returned %d for %s\n", resp.StatusCode, eva)
			resp.Body.Close()
			continue
		}

		var res StationsResult
		if err := xml.NewDecoder(resp.Body).Decode(&res); err != nil {
			log.Printf("Warning: failed to decode XML for %s: %v\n", eva, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		for _, s := range res.Station {
			if s.EVA == eva {
				stations = append(stations, domain.Station{
					EVA:   s.EVA,
					Name:  s.Name,
					DS100: s.DS100,
				})
				break
			}
		}
	}

	return stations, nil
}
