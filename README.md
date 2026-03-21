# S-Bahn München Telemetry System

A highly secure, privacy-compliant, cross-platform telemetry system designed to track and analyze real-time delays of the S-Bahn München transit network using the official Deutsche Bahn Timetables API.

## Project Structure

```text
sbahnmuc-ag/
├── android/                 # Android project (Kotlin/Gradle)
├── api/                     # Protobuf/GTFS-RT definitions
├── assets/                  # Static data (e.g., station identifiers)
├── build/                   # Compiled binaries and AAR files (git-ignored)
├── cmd/                     # Application entry points
│   ├── daemon/              # Main polling daemon for local/server use
│   ├── cloud-job/           # GCP Cloud Run Job entry point
│   └── mobile-bind/         # Go-to-Android binding entry point
├── deploy/                  # Deployment configurations
│   ├── local/               # systemd unit files
│   └── cloud/               # Terraform/GCP configs
├── internal/                # Private application logic
│   ├── adapters/            # External interface implementations (Storage, Transit)
│   ├── domain/              # Core Models & Ports (Hexagonal Architecture)
│   └── engine/              # Polling and orchestration logic
├── scripts/                 # Utility scripts (Python, Bash)
├── test/                    # Integration tests and test utilities
├── GEMINI.md                # Agent-specific documentation and mandates
├── go.mod                   # Go module definition
└── README.md                # This file
```

## Architecture

This project strictly follows the **Hexagonal (Ports and Adapters)** architecture pattern, separating core business logic from external dependencies:
- **Core Engine (Golang)**: A background poller executing every 1 minute.
- **Transit Provider (DB API)**: Integrates with the official DB Timetables API via client credentials to fetch realtime delay metrics.
- **Storage Provider (SQLite)**: Stores time-series data locally using SQLite with Write-Ahead Logging (WAL) for safety and concurrent read/writes.

## Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [Deutsche Bahn Open Data Client Credentials](https://developers.deutschebahn.com/db-api-marketplace/apis/)

## Configuration

The application requires a `.env` file in the root directory containing your DB API credentials:

```env
db_client_id=your_client_id_here
db_client_secret_api_key=your_api_key_here
DB_PATH=telemetry.db
```

## Running Locally

To build and run the daemon locally:

```bash
# 1. Fetch dependencies
go mod tidy

# 2. Run unit tests
go test ./... -v

# 3. Build the core daemon
go build -o build/sbahn-telemetry ./cmd/daemon

# 4. Run the daemon
./build/sbahn-telemetry
```

## Mobile & Cloud (GCP)

### Android Binding
To generate the Android Archive (`.aar`) for integration into the Kotlin app:
```bash
gomobile bind -target=android -o ./android/app/libs/telemetry.aar ./internal/engine
```

### Cloud Run Job
To deploy as a GCP Cloud Run Job:
```bash
gcloud run jobs deploy sbahn-polling-job --source . --region <REGION>
```

## Engine Features & Polling Logic

- **Real-Time Delays Polling**: The engine polls 28 verified S-Bahn München stations every 1 minute.
- **Strict Data Filtering**: Filters exclusively for lines starting with "S" (S-Bahn).
- **API Metadata Analytics**: Tracks request latency and status codes for diagnostics.
- **Station Master Data**: Periodically resolves EVA IDs to human-readable names and DS100 codes.

## Viewing Results Locally

Query the `telemetry.db` file using `sqlite3`:

```bash
# See strictly delayed trains
sqlite3 telemetry.db "SELECT route_id, delay_seconds, station_id, expected_arrival FROM delay_records WHERE delay_seconds > 0 ORDER BY delay_seconds DESC;"

# See API diagnostics
sqlite3 telemetry.db "SELECT * FROM station_request_logs WHERE status_code != 200;"
```
