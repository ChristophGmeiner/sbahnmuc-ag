package engine

import (
	"context"
	"log"
	"time"

	"github.com/christophgmeiner/sbahnmuc-ag/internal/domain"
)

// Engine is the core background poller handling the transit telemetry ingestion.
type Engine struct {
	transit domain.TransitProvider
	storage domain.StorageProvider
	ticker  *time.Ticker
	done    chan bool
}

// NewEngine initializes the core polling engine.
func NewEngine(transit domain.TransitProvider, storage domain.StorageProvider) *Engine {
	return &Engine{
		transit: transit,
		storage: storage,
		done:    make(chan bool),
	}
}

// Start begins the 5-minute continuous polling loop in the background.
func (e *Engine) Start(ctx context.Context) {
	// First poll immediately
	e.poll(ctx)

	// Setup 1-minute ticker
	e.ticker = time.NewTicker(1 * time.Minute)

	go func() {
		for {
			select {
			case <-e.ticker.C:
				e.poll(ctx)
			case <-e.done:
				log.Println("Stopping telemetry engine ticker.")
				e.ticker.Stop()
				return
			case <-ctx.Done():
				log.Println("Context cancelled, stopping telemetry engine ticker.")
				e.ticker.Stop()
				return
			}
		}
	}()
}

// Stop gracefully stops the background polling.
func (e *Engine) Stop() {
	e.done <- true
}

// poll executes a single cycle of fetching and saving data.
func (e *Engine) poll(ctx context.Context) {
	log.Println("Starting telemetry poll cycle...")

	records, logsPayload, err := e.transit.FetchDelays(ctx)
	if err != nil {
		log.Printf("ERROR fetching transit data: %v\n", err)
		return
	}

	if len(records) == 0 && len(logsPayload) == 0 {
		log.Println("No delay records or API logs retrieved.")
		return
	}

	log.Printf("Fetched %d delay records and %d API logs. Saving to storage...\n", len(records), len(logsPayload))

	if err := e.storage.SaveRecords(ctx, records, logsPayload); err != nil {
		log.Printf("ERROR saving transit data: %v\n", err)
		return
	}

	log.Println("Successfully saved telemetry records.")
}
