---
name: api-design-principles
description: Assists with structuring the data ingestion and the local storage retrieval logic.
---
# API Design Principles

Detailed instructions for establishing clear, consistent, and decoupled interfaces across the application.

## When to use this skill
- Use this when designing the structural boundaries between the Android UI and the Go backend, or when structuring cloud ingestion APIs.
- This is helpful for maintaining a clean architecture and efficient data transfer.

## How to use it
- **Cross-Platform Boundaries:** When designing the Java Native Interface (JNI) via `gomobile bind`, ensure calls are coarse-grained. Never pass massive data objects across the boundary.
- **Data Encapsulation:** The Go engine must encapsulate all GTFS-RT network requests and parsing complexity, returning only primitive types or simple, flat structs to the Kotlin UI layer.
- **Cloud Design:** Design BigQuery streaming structures to be append-only, ensuring tables are correctly partitioned by timestamp and clustered by route ID.