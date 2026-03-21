# S-Bahn München Telemetry Application: Agent Blueprint

## 1. Project Overview
The **S-Bahn München Telemetry Application** (`sbahnmuc-ag`) is a highly secure, privacy-compliant, cross-platform telemetry system designed to track and analyze real-time delays of the S-Bahn München transit network. The system leverages official GTFS-Realtime (protobuf) feeds to provide up-to-date transit data across cloud, local, and mobile environments.

### Core Architecture
- **Hexagonal (Ports and Adapters):** The core business logic is decoupled from external dependencies via interfaces:
    - `TransitProvider`: For fetching and parsing GTFS-RT feeds.
    - `StorageProvider`: For persisting time-series and state data.
- **Domain Focus:** Continuous polling (every 5 minutes), efficient protobuf parsing, and secure local storage.

### Technology Stack
- **Backend:** Golang (Go) for the core polling engine and business logic.
- **Mobile:** Android (Kotlin) for the UI, utilizing `gomobile bind` to integrate the Go engine.
- **Cloud (GCP):** Cloud Run Jobs (serverless compute), Cloud Scheduler (cron triggering), and BigQuery (analytical time-series database).
- **Local/Edge:** Linux systemd daemons with SQLite utilizing Write-Ahead Logging (WAL).
- **Security:** SQLCipher (AES-256) for local database encryption on Android.

---

## 2. Building and Running
### Backend (Go)
```bash
# Initialize/Update dependencies
go mod tidy

# Run tests
go test ./...

# Build the polling engine
go build -o sbahn-telemetry ./cmd/engine
```

### Mobile (Android/Go)
```bash
# Bind Go code to Android Archive (.aar)
gomobile bind -target=android -o ./android/app/libs/telemetry.aar ./internal/engine

# Build Android App (via Gradle)
./gradlew assembleDebug
```

### Cloud (GCP)
```bash
# Deploy the Go engine as a Cloud Run Job
gcloud run jobs deploy sbahn-polling-job --source . --region <REGION>

# TODO: Configure Cloud Scheduler to trigger every 5 minutes (cron: */5 * * * *)
```

### Local Daemon
- **TODO:** Install the generated `sbahn-telemetry.service` systemd unit to `/etc/systemd/system/`.

---

## 3. Development Conventions
### Coding Standards
- **Interface-First:** Always define `TransitProvider` and `StorageProvider` implementations to maintain architectural boundaries.
- **Concurrency:** Utilize Go `goroutines` for non-blocking network requests and background polling.
- **Memory Optimization:** Use `io.Reader` streams for protobuf parsing to minimize RAM usage, especially on mobile devices.
- **JNI Best Practices:** Keep the boundary between Go and Kotlin "coarse-grained" to minimize overhead.

### Testing (TDD)
- **Mandatory TDD:** Write unit tests for all new parsing logic and database interactions before implementation.
- **Mocking:** Use mock implementations of `TransitProvider` and `StorageProvider` for isolated engine testing.
- **Reliability:** Ensure tests cover network timeouts, malformed protobuf data, and race conditions.

### Security & Privacy
- **Privacy by Design:** Adhere to GDPR; process user location data strictly locally on-device.
- **Data Minimization:** Default to strict data minimization; do not transmit sensitive user data off-device.
- **Encryption:** All local SQLite databases on Android MUST be encrypted via SQLCipher (AES-256).

### Safety Constraints
- **Explicit Confirmation:** NEVER execute terminal commands without explicit, in-line confirmation from the user.
- **Restricted Access:** Read/write operations are limited to the immediate project context.
- **Destructive Actions:** Preface potentially destructive commands (e.g., `rm`, `mv`) with a clear warning.

Consider for more infos ALWAYS the informatio in the file S-Bahn-München-Telemetry-Application.pdf

Always check README.md for more and current information.