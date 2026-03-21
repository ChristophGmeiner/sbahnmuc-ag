---
name: frontend-mobile-development-component-scaffold
description: Helps scaffold the Android mobile UI layers that interact with the Go backend.
---
# Android Frontend Development Scaffold

Detailed instructions for building the Android native UI and handling OS lifecycle constraints.

## When to use this skill
- Use this when creating the Kotlin UI, setting up background tasks, or integrating the `gomobile` `.aar` library.
- This is helpful for ensuring the mobile app complies with strict Android battery and background execution limits.

## How to use it
- **Background Execution:** Scaffold a Bound Foreground Service that displays a persistent notification to guarantee the 5-minute polling interval is not terminated by the Android OS.[2]
- **Thread Management:** Utilize Kotlin coroutines to invoke the Go telemetry engine.[3] Ensure the Go process runs on a background thread so the UI thread does not freeze.[3]
- **UI Updates:** Only update the Android View components on the main UI thread after the background Go processing has returned its metrics.[3]