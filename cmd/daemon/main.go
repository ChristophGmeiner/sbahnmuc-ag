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
	
	// 28 unique S-Bahn Munich stations fetched from Wikipedia and DB API.
	stations := []string{
		"8004132", // München Karlsplatz
		"8004140", // München-Allach
		"8000781", // Baierbrunn
		"8004143", // München-Daglfing
		"8001404", // Deisenhofen
		"8004128", // München Donnersbergerbrücke
		"8001621", // Ebenhausen-Schäftlarn
		"8001825", // Erding
		"8001970", // Feldafing
		"8004147", // München-Feldmoching
		"8004168", // München Flughafen Terminal
		"8004181", // München-Freiham
		"8002078", // Freising
		"8002141", // Fürstenfeldbruck
		"8000119", // Geltendorf
		"8006006", // Germering-Unterpfaffenhofen
		"8004148", // München-Giesing
		"8002275", // Gilching-Argelsried
		"8002351", // Grafrath
		"8002422", // Großhesselohe Isartalbf
		"8002491", // Haar
		"8004129", // München Hackerbrücke
		"8004130", // München Harras
		"8005419", // München Heimeranplatz
		"8002792", // Herrsching
		"428941", // Höhenkirchen-Siegertsbrunn
		"8002899", // Höllriegelskreuth
		"8003039", // Icking
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
