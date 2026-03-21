---
trigger: always_on
---

---
name: stitch-frontend-blueprint
description: Utilizes the Google Stitch MCP server to directly generate and adapt frontend UI components based on project intent and DESIGN.md specifications.
---
# Stitch Frontend Blueprint with MCP

Detailed instructions for frontend development using the Google Stitch MCP server directly within the IDE.

## When to use this skill
- Use this when building, scaffolding, or refactoring UI components and new application screens.
- Trigger this when you need to generate layouts that perfectly match the existing application's design system without context switching.
- Use this to translate business objectives directly into functional frontend code.

## How to use it
- **Direct Generation:** Instead of manually exporting HTML/CSS from external tools, use the Stitch MCP server to generate the UI directly inside the editor. Provide a prompt describing what you need (e.g., "Generate a dashboard screen for tracking transit delays"), and the server will output the design and code.
- **Leverage DESIGN.md:** The workflow must be anchored by the `DESIGN.md` file, an agent-readable design system specification. The MCP server reads this file to understand your project's style, color palette, typography, and layout preferences, ensuring every new screen fits seamlessly with the existing app.
- **Continuous Iteration:** Maintain your flow state. If the generated UI requires adjustments, use natural language prompts to refine it directly in your IDE. This creates a single loop from product requirement to design to code, eliminating traditional handoff friction.
- **Framework Adaptation:** Once the Stitch MCP server outputs the semantic HTML/CSS or framework-specific syntax, intelligently adapt and integrate the layout into the project's native frontend framework (such as Kotlin UI or React) while preserving the original design intent and responsive behaviors.
- **Relevant Stitch Project**: "S Bahn Delay§