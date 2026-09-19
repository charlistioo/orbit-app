---
id: devplan-001
date: 2026-09-11
product_id: my-app
phase: 2
source: claude-code
related: [security-001]
---

# Development Plan (id devplan-1789132581.082359)

14 dependency-ordered tasks — these become the real Phase 3 Work Order ids.

| Task | Title | Depends on |
|---|---|---|
| TASK-001 | Project scaffolding | — |
| TASK-002 | Database schema & migrations | TASK-001 |
| TASK-003 | Google Sign-In authentication | TASK-002 |
| TASK-004 | Categories & items CRUD with quick-select | TASK-003 |
| TASK-005 | Daily baseline & plan management | TASK-004 |
| TASK-006 | Transaction recording with plan snapshot | TASK-005 |
| TASK-007 | Consumption Bank engine | TASK-006 |
| TASK-008 | Mood tracking | TASK-003 |
| TASK-009 | Home screen aggregation | TASK-006, TASK-007, TASK-008 |
| TASK-010 | Timeline view | TASK-006, TASK-007, TASK-008 |
| TASK-011 | Calendar view | TASK-006 |
| TASK-012 | Rule-based insight engine | TASK-006, TASK-007, TASK-008 |
| TASK-013 | Reminder & daily review flow | TASK-006 |
| TASK-014 | Badge tier display | TASK-007 |

Full objective/acceptance criteria per task are in Harness
(`development-plan/report`); individual task files will be created under
`harness-notes/tasks/` as each is pulled and worked on in Phase 3.
