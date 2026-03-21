# S-Bahn München Telemetry System

A highly secure, privacy-compliant, cross-platform telemetry system designed to track and analyze real-time delays of the S-Bahn München transit network using the official Deutsche Bahn Timetables API.

## Architecture

This project strictly follows the Hexagonal (Ports and Adapters) architecture pattern, separating core business logic from external dependencies (HTTP transit feeds and local databases):
- **Core Engine (Golang)**: A background poller executing every 5 minutes.
- **Transit Provider (DB API)**: Integrates with the official DB Timetables API via client credentials to fetch realtime delay metrics.
- **Storage Provider (SQLite)**: Stores time-series data locally using SQLite with Write-Ahead Logging (WAL) for safety and concurrent read/writes.

## Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [Deutsche Bahn Open Data Client Credentials](https://developers.deutschebahn.com/db-api-marketplace/apis/)

## Configuration

The application requires a `.env` file in the root directory (or working directory of the process) containing your DB API credentials:

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

# 3. Build the core engine
go build -o sbahn-telemetry ./cmd/local

# 4. Run the executable synchronously
./sbahn-telemetry

# OR run it persistently in the background routing output to a log file
./sbahn-telemetry >> telemetry-daemon.log 2>&1 &

# View the live logs of the background daemon
tail -f telemetry-daemon.log

# Safely kill the background daemon
pkill -f sbahn-telemetry
```

## Engine Features & Polling Logic

- **Real-Time Delays Polling (`/v1/fchg/{evaNo}`)**: The backend retrieves realtime timetable updates using the DB Timetables API.
  - **Endpoint**: `https://apis.deutschebahn.com/db-api-marketplace/apis/timetables/v1/fchg/{evaNo}`
  - **Extracted Fields**:
    - `id`: Unique journey identifier (stop ID).
    - `route_id`: The categorized line name (e.g., "S1", "S2").
    - `expected_arrival`: Derived from `pt` (Planned Time) tag.
    - `actual_arrival`: Derived from `ct` (Changed Time) tag.
    - `delay_seconds`: Computed time difference (strictly >= 0).
  *(Documentation: [DB Timetables API](https://developers.deutschebahn.com/db-api-marketplace/apis/product/timetables))*
- **Rate Limit Optimization**: The Deutsche DB Timetables API strictly limits polling to 60 requests per minute. The polling engine is explicitly loaded with exactly 28 verified S-Bahn München EVA station identifiers (sourced from Wikipedia's master list) alongside a 1-minute ticker to constantly pull the highest-resolution telemetry safely without triggering 429 timeouts.
- **Strict Data Filtering**: To prevent noise, the parsing engine aggressively scans the DB `fchg` XML payloads and filters exclusively for line names starting with `S` (dropping all "Unknown", "ICE", "RE", and "RB" trains from local storage).
- **Zero-Delay Tracking**: The custom XML parser dynamically tracks trains even if they are lacking a `ct` (Changed Time) tag, identifying that they are operating 100% on schedule and successfully writing a `0` delay into the database.
- **API Request Metadata Analytics**: A secondary data pipeline seamlessly tracks every individual `http.Client.Do()` request, actively capturing the exact timestamp, target station, and returned HTTP Status Code into a standalone analytic table so 400 errors can be natively diagnosed.
- **Station Master Data Polling**: To provide full context for the raw EVA station numbers, the engine periodically queries the DB Timetables API `/v1/station/{eva}` endpoint directly for each of the 28 configured stations. This precisely resolves exact master data without wildcard query pollution, automatically extracting and storing crucial metadata into a separate `stations` table:
  - `eva`: The internal DB station identifier.
  - `name`: The human-readable station name (e.g., "München Marienplatz").
  - `ds100`: The official alphanumeric DB location code (e.g., "MMAR").
  *(Documentation: [DB Timetables API](https://developers.deutschebahn.com/db-api-marketplace/apis/product/timetables))*

## Linux Background Deployment (systemd)

> [!NOTE]
> These instructions and commands (`systemctl`, `journalctl`) are **exclusive to Linux**. macOS uses `launchd` and does not support `systemd`. To run this on a Mac, you can simply keep the `./sbahn-telemetry` process running in a normal terminal window or running persistently with the `&` background operator shown above.

A systemd unit file is provided to run this polling engine continuously in the background on Linux machines (e.g., a home server or Raspberry Pi):

```bash
# Assuming the binary and .env are in /opt/sbahn-telemetry
sudo mkdir -p /opt/sbahn-telemetry
sudo cp sbahn-telemetry /opt/sbahn-telemetry/
sudo cp .env /opt/sbahn-telemetry/
sudo cp deploy/local/sbahn-telemetry.service /etc/systemd/system/

sudo systemctl daemon-reload
sudo systemctl enable --now sbahn-telemetry.service

# View the logs
journalctl -fu sbahn-telemetry.service
```

## Viewing Results Locally

The daemon automatically creates a local database file named `telemetry.db` in your current folder and writes the delay data there. Because we enabled **WAL mode** (Write-Ahead Logging) in the SQLite setup, you can safely query the database while the Go daemon is writing to it simultaneously without locking it!

Open a new, separate terminal window and run these commands to inspect the datastore:

### Telemetry Queries (`delay_records`)

**See strictly delayed trains sorted by highest delay:**
```bash
sqlite3 telemetry.db "SELECT route_id, delay_seconds, station_id, expected_arrival FROM delay_records WHERE delay_seconds > 0 ORDER BY delay_seconds DESC;"
```

**See all perfectly on-time trains (0 seconds delay):**
```bash
sqlite3 telemetry.db "SELECT route_id, station_id, delay_seconds, actual_arrival FROM delay_records WHERE delay_seconds = 0;"
```

**See the total number of telemetry records collected:**
```bash
sqlite3 telemetry.db "SELECT COUNT(*) FROM delay_records;"
```

**Emergency: Purge and truncate all captured data to start fresh:**
```bash
sqlite3 telemetry.db "DELETE FROM delay_records;"
```

### Station Master Data (`stations`)

**View the tracked stations and their metadata:**
```bash
sqlite3 telemetry.db "SELECT * FROM stations LIMIT 15;"
```

### API Diagnostic Queries (`station_request_logs`)

**See the 15 most recent network pings across the 28 tracked stations:**
```bash
sqlite3 telemetry.db "SELECT * FROM station_request_logs LIMIT 15;"
```

**Diagnose failed DB endpoints (400 responses):**
```bash
sqlite3 telemetry.db "SELECT * FROM station_request_logs WHERE status_code != 200;"
```

### Database Schema

If you want to view the internal database schemas, you can print the exact table definitions by running:
```bash
sqlite3 telemetry.db ".schema"
```

**Interactive Querying:**
You can also open the interactive SQLite prompt to explore the data freely by running:
```bash
sqlite3 telemetry.db
```
*(Once inside the prompt, you can type `.tables`, `.schema`, or any standard SQL query ending with a semicolon. Type `.quit` to exit).*
