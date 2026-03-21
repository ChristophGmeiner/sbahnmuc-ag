---
name: golang-pro
description: Enforces advanced Go patterns, optimal concurrency using goroutines, and memory management necessary for the polling engine.
---
# Golang Professional Development

Detailed instructions for the agent to write highly optimized, concurrent, and idiomatic Go code.

## When to use this skill
- Use this when developing or modifying the core Go telemetry polling engine.
- This is helpful for writing robust concurrent network requests and GTFS-RT protobuf parsing.

## How to use it
- **Concurrency:** Use `goroutines` and channels to handle the 5-minute polling interval without blocking the main execution thread.
- **Memory Management:** Implement `io.Reader` streams to process large protobuf payloads incrementally, minimizing the RAM footprint for mobile environments.
- **Architecture:** Strictly adhere to interface-driven design (Ports and Adapters). All database and network interactions must be abstracted behind interfaces like `TransitProvider` and `StorageProvider`.