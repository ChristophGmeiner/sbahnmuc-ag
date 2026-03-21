# S-Bahn München Telemetry Application: Agent Blueprint

## 1. Project Overview
The **S-Bahn München Telemetry Application** (`sbahnmuc-ag`) is a highly secure, privacy-compliant, cross-platform telemetry system designed to track and analyze real-time delays of the S-Bahn München transit network. The system leverages official DB Timetables (XML) feeds to provide up-to-date transit data across cloud, local, and mobile environments.

### Core Architecture
- **Hexagonal (Ports and Adapters):** The core business logic is decoupled from external dependencies via interfaces:
    - `TransitProvider`: For fetching and parsing transit feeds.
    - `StorageProvider`: For persisting time-series and state data.
- **Domain Focus:** Continuous polling (every 1 minute), efficient XML parsing, and secure local storage.

### Technology Stack
- **Backend:** Golang (Go) for the core polling engine and business logic.
- **Mobile:** Android (Kotlin) for the UI, utilizing `gomobile bind` to integrate the Go engine.
- **Cloud (GCP):** Cloud Run Jobs (serverless compute), Cloud Scheduler (cron triggering), and BigQuery (analytical time-series database).
- **Local/Edge:** Linux systemd daemons with SQLite utilizing Write-Ahead Logging (WAL).
- **Security:** SQLCipher (AES-256) for local database encryption on Android.

---

## 2. Building and Running
### Backend (Go Daemon)
```bash
# Initialize/Update dependencies
go mod tidy

# Run tests
go test ./...

# Build the polling engine
go build -o build/sbahn-telemetry ./cmd/daemon
```

### Mobile (Android/Go)
```bash
# Bind Go code to Android Archive (.aar)
gomobile bind -target=android -o ./android/app/libs/telemetry.aar ./internal/engine

# Build Android App (via Gradle)
cd android && ./gradlew assembleDebug
```

### Cloud (GCP)
```bash
# Deploy the Go engine as a Cloud Run Job
gcloud run jobs deploy sbahn-polling-job --source . --region <REGION>
```

### Local Daemon
- Install the generated `sbahn-telemetry.service` systemd unit to `/etc/systemd/system/`.

---

## 3. Development Conventions
### Project Structure
- `cmd/`: Entry points for different platforms (daemon, cloud, mobile).
- `internal/`: Core business logic and adapters.
- `deploy/`: Infrastructure as Code and service configurations.
- `assets/`: Static configuration data (stations, etc).
- `scripts/`: Maintenance and utility scripts.

### Coding Standards
- **Interface-First:** Always define `TransitProvider` and `StorageProvider` implementations to maintain architectural boundaries.
- **Concurrency:** Utilize Go `goroutines` for non-blocking network requests and background polling.
- **Memory Optimization:** Use efficient parsing to minimize RAM usage, especially on mobile devices.

### Testing (TDD)
- **Mandatory TDD:** Write unit tests for all new parsing logic and database interactions before implementation.
- **Mocking:** Use mock implementations of `TransitProvider` and `StorageProvider` for isolated engine testing.

### Security & Privacy
- **Privacy by Design:** Adhere to GDPR; process user location data strictly locally on-device.
- **Encryption:** All local SQLite databases on Android MUST be encrypted via SQLCipher (AES-256).
