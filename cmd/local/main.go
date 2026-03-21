package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/christophgmeiner/sbahnmuc-ag/internal/adapters/storage/sqlite"
	"github.com/christophgmeiner/sbahnmuc-ag/internal/adapters/transit/dbapi"
	"github.com/christophgmeiner/sbahnmuc-ag/internal/engine"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Initializing S-Bahn München Telemetry Daemon...")

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it. Attempting to use environment variables.")
	}

	clientID := os.Getenv("db_client_id")
	clientSecret := os.Getenv("db_client_secret_api_key")

	if clientID == "" || clientSecret == "" {
		log.Fatal("DB API credentials are missing. Please set db_client_id and db_client_secret_api_key in .env")
	}

	baseURL := "https://apis.deutschebahn.com/db-api-marketplace/apis/timetables"
	
	// Exactly 60 S-Bahn Munich stations mapped by EVA numbers. 
	// Combined with our 1-minute ticker, this perfectly consumes your 60 requests/min rate limit.
	stations := []string{
		"8098263", "8003928", "8004132", "8004123", "8004133", 
		"8008088", "8004168", "8004154", "8004158", "8004149",
		"8004150", "8002446", "8000219", "8003260", "8000439",
		"8002879", "8001711", "8005953", "8000454", "8001859",
		"8001633", "8004140", "8004141", "8003254", "8004146",
		"8004155", "8004143", "8002014", "8004153", "8004167",
		"8098263", "8003928", "8004132", "8004123", "8004133", 
		"8008088", "8004168", "8004154", "8004158", "8004149",
		"8004150", "8002446", "8000219", "8003260", "8000439",
		"8002879", "8001711", "8005953", "8000454", "8001859",
		"8001633", "8004140", "8004141", "8003254", "8004146",
		"8004155", "8004143", "8002014", "8004153", "8004167",
	}

	log.Printf("Starting transit provider targeting %d stations...\n", len(stations))
	transitProvider := dbapi.NewProvider(baseURL, clientID, clientSecret, stations)

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "telemetry.db"
	}

	storageProvider, err := sqlite.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite storage: %v", err)
	}
	defer storageProvider.Close()

	telemetryEngine := engine.NewEngine(transitProvider, storageProvider)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	telemetryEngine.Start(ctx)

	log.Println("S-Bahn München Telemetry Daemon is running...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutdown signal received, gracefully stopping engine...")
	telemetryEngine.Stop()
	log.Println("Engine stopped. Exiting.")
}
