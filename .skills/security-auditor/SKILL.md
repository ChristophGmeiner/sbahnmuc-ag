---
name: security-auditor
description: Reviews local SQLite SQLCipher encryption implementations and ensures API data handling adheres to privacy best practices.
---
# Security Auditor

Detailed instructions for reviewing code to ensure the highest security and GDPR compliance standards.

## When to use this skill
- Use this when reviewing database storage logic, API data handling, or evaluating privacy configurations.
- This is helpful for enforcing Privacy by Design principles and robust local encryption.

## How to use it
- **Encryption Standards:** Verify that local storage implementations correctly utilize SQLCipher for AES-256 encryption at rest.[1]
- **Data Minimization:** Ensure default privacy settings are applied (e.g., opt-in consent models) and that personal user data (like GPS location) is processed strictly locally and never transmitted off-device.[1]
- **Code Scanning:** Check for high-risk command patterns or hardcoded credentials within the codebase and ensure strict access controls are maintained.