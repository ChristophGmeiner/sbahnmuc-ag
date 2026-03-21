---
name: test-driven-development
description: Ensures the continuous GTFS-RT polling engine and protobuf parser are thoroughly tested.
---
# Test-Driven Development (TDD)

Detailed instructions for maintaining high test coverage and reliability for the application.

## When to use this skill
- Use this when adding new parsing logic, building statistical aggregations, or modifying the backend database implementations.
- This is helpful for ensuring high reliability in the core continuous polling engine.

## How to use it
- **Test First:** Always write unit tests using Go's standard `testing` package before implementing the actual function logic.
- **Mocking:** Create mock implementations of the `TransitProvider` and `StorageProvider` interfaces to isolate and test the core engine's behavior without making live network calls.
- **Edge Cases:** Ensure tests cover network timeouts, malformed protobuf data, and concurrent map write scenarios.