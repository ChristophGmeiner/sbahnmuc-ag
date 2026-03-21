# Security Guardian Skill
**Description**: Automatically audits code for plain-text secrets, hardcoded credentials, and dangerous patterns (SQLi, XSS, eval).
**Triggers**: "check security", "scan for secrets", "audit this plan", "is this safe?"

## Instructions
1. **Pre-Implementation**: Before proceeding with any `implementation_plan.md`, you MUST run the `audit.py` script on the proposed diffs.
2. **Secret Detection**: If any high-entropy strings or known patterns (API keys, tokens) are found, halt execution and issue a **CRITICAL WARNING**.
3. **Behavioral Check**: Look for "obvious insecure behavior" such as:
   - Using `shell=True` in subprocesses.
   - Using `eval()` or `exec()`.
   - Hardcoded IP addresses or `0.0.0.0` bindings.
4. **Mitigation**: For every finding, provide a "Suggested Mitigation" block (e.g., using `python-dotenv` or Secret Manager).

## Tools
- `scan_workspace`: Runs `python3 scripts/audit.py` 