# Design Index (Phase 2)

Phase 2 (Technical Design) started 2026-09-11 for product `my-app` (ORBIT).
Phase 1 approved by Charli Stiow on 2026-09-11.

All Phase 2 validation checks passed (including tech stack risk/mitigation
records added 2026-09-11) — design is ready for final approval.

## Solutions
- [solution-001](solution-001-flexible-baseline-consumption-bank.md) — Flexible Baseline + Consumption Bank
- [solution-002](solution-002-server-computed-aggregation.md) — Server-computed Aggregation for Views & Insight

## Architecture
- [architecture-001](architecture-001-modular-monolith.md) — Modular monolith, 7 components, ADR-001

## Tech Stack
- [techstack-001](techstack-001-mobile-backend-database.md) — React Native/Expo, Go, PostgreSQL, plus failure-mode/mitigation records per category
- [techstack-002](techstack-002-nextjs-web-pivot.md) — Next.js + Tailwind + HeroUI (`frontend_web`), 2026-09-17 pivot replacing React Native/Expo as the client, approved by Charli Stiow

## Database
- [database-001](database-001-schema.md) — 9 entities

## API
- [api-001](api-001-endpoints.md) — 19 REST endpoints

## Security & Integrations
- [security-001](security-001-requirements.md) — 4 security requirements, Google Sign-In integration, 2 NFRs

## Development Plan
- [devplan-001](devplan-001-tasks.md) — 14 dependency-ordered tasks (TASK-001 to TASK-014)
