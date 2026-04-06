# Mock Data Seeding Guidelines

## Purpose
Define a consistent, business-valid mock seeding strategy for this backend, with clear separation of concerns and predictable reruns.

This document is a guideline and implementation target. It does not execute seeding by itself.

## Core Decisions
- Use a single Go CLI entrypoint to orchestrate all seeders.
- Keep one seeder module per top-level table or aggregate.
- Use `gofakeit` only for value generation, wrapped behind domain factories.
- Prioritize business-coherent Demo/QA data.
- Use split rerun behavior:
- Reference/static data: idempotent ensure/upsert behavior.
- Scenario/mock data: append-only per run.
- Parent owns child when child lifecycle strongly depends on parent existence.

## Target Structure
- Entrypoint: `scripts/seed_mock/main.go`
- Seeder modules: `internal/seed/...`
- Shared runtime:
- `Seeder` interface: name, dependencies, profile support, seed function.
- `SeedEnv`: DB/store, faker, profile, run label, clock, logger.
- `SeedState`: alias-based registry for created/found IDs.

## Foreign Key and Ownership Strategy
- Do not rely on foreign keys alone to define seeding boundaries.
- Each top-level table/aggregate has its own seeder.
- Child tables with strong lifecycle coupling are seeded inside parent seeders.

Examples of parent-owned children:
- `custom_user` + `employee_profile` seeded together.
- `location_shift` owned by location seeder.
- `user_roles` owned by employee/admin seeder.
- `employee_education`, `certification`, `employee_experience`, `employee_contract_changes` owned by employee seeder.
- `handbook_steps` owned by handbook template seeder.
- `employee_handbooks`, assignment history, step progress owned by handbook assignment seeder.
- `pay_period_line_items` owned by pay period seeder.
- `calendar_event_attendees`, `calendar_event_reminders` owned by calendar event seeder.

## Data Realism Rules
- All seed values must conform to business logic and constraints.
- Use factories to enforce valid combinations, not random free-form generation.
- Keep stable aliases for FK linking (example: `employee.admin`, `location.main_1`).
- Avoid impossible states unless explicitly building an edge-case profile.

## Profiles
- `baseline`: minimal deterministic dataset for local development.
- `demo`: richer dataset for QA/demo scenarios, built on baseline foundations.

## Rerun Policy
- Reference/static entities are ensured/upserted on rerun.
- Scenario entities are appended with a `run_label` to avoid unique conflicts.
- Recommended unique keys to suffix with `run_label` where needed:
- Email fields
- Organization or template display names
- Employee/employment numbers

## Seeding Order (High Level)
1. Reference foundation: app org profile, optional admin ensure.
2. Organizations and locations.
3. Departments.
4. Users and employee profiles.
5. Employee-owned detail tables.
6. Handbook templates and assignments.
7. Schedules.
8. Time entries.
9. Leave requests and leave adjustments.
10. Payout requests and pay periods.
11. Calendar events.
12. Late arrivals and shift swaps.

## Existing DB Side Effects to Respect
- Inserting `location` triggers default `location_shift` creation.
- Inserting eligible `employee_profile` can initialize `leave_balances`.
- Employee creation flow can auto-assign active handbook templates when applicable.

Seeders must not fight these side effects; they should align with them.

## Testing Checklist
- Dependency graph order and cycle detection.
- Alias resolution and missing dependency errors.
- Factory output validity against constraints.
- Integration run for `baseline` on empty DB.
- Integration run for `demo` after baseline.
- Rerun behavior with different `run_label`.
- Subset seeding behavior with auto dependency handling.

## Non-Goals (For Now)
- No schema changes required for v1 seeding structure.
- No seed-run tracking table required initially.
- No implementation in this step; this document is the reference guideline.

