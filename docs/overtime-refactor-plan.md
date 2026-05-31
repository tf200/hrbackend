# Overtime Refactor Plan

## Bug Context

The current `time_entries` feature was implemented as a general worked-hours model. That was a misunderstanding.

The intended product model is:

- Regular working hours are derived from `schedules`.
- Extra work is submitted as overtime.
- Overtime should have its own model, API, repository, service, queries, permissions, and payroll handling.
- The current `time_entries` model, service, handlers, SQL queries, and payroll usage should be removed over multiple controlled parts.

This is a large cross-cutting refactor. Do not attempt to replace everything in one change unless explicitly decided later.

## Current Incorrect Behavior

The app currently treats `time_entries` as payable work logs with these hour types:

- `normal`
- `overtime`
- `travel`
- `leave`
- `sick`
- `training`

This causes regular worked hours to be entered manually and counted from `time_entries`, while they should be derived from scheduled shifts.

Known impacted areas:

- `internal/domain/time_entry.go`
- `internal/service/time_entry_service.go`
- `internal/repository/time_entry_repo.go`
- `internal/repository/db/query/time_entry.sql`
- `internal/handler/time_entry_dto.go`
- `internal/handler/time_entry_handler.go`
- `internal/handler/time_entry_routes.go`
- `internal/seed/time_entries.go`
- `internal/app/app.go` time entry wiring
- `migrations/000001_init.up.sql` time entry schema and pay period references
- `migrations/000001_init.down.sql` time entry teardown
- `migrations/permissions_catalog.txt` time entry permissions
- payroll queries using `time_entries`
- employee/profile stats using `time_entries`
- generated SQLC files after query/schema regeneration

## Target Model

Replace `time_entries` with `overtime_entries`.

Overtime statuses are only:

- `submitted`
- `approved`
- `rejected`

There is no `draft` state.

Employee-created overtime should be submitted immediately unless a later product decision says otherwise.

## Overtime Entry Fields

The target overtime entry should contain:

- `id`
- `employee_id`
- `entry_date`
- `schedule_id`
- `minutes`
- `reason`
- `description`
- `status`
- `submitted_at`
- `approved_at`
- `approved_by_employee_id`
- `rejection_reason`
- `paid_period_id`
- `created_at`
- `updated_at`

`schedule_id` should remain nullable unless the product confirms that every overtime entry must be tied to a schedule.

## Overtime Reason Enum

Use stable API/database enum values instead of display labels:

- `client_crisis`
- `understaffing`
- `meeting_consultation`
- `training_education`
- `completing_administration`
- `handover`
- `emergency`
- `project_work`
- `event_activity`
- `other`

Suggested display labels for the frontend:

- Client crisis
- Understaffing
- Meeting / Consultation
- Training / Education
- Completing administration
- Handover
- Emergency
- Project work
- Event / Activity
- Other

## High-Level Refactor Parts

### Part 1: Add Overtime Feature Beside Time Entries

Goal: introduce the correct overtime domain and API without removing old time entry behavior yet.

This lowers risk and gives the frontend/backend a clean target API to integrate first.

Part 1 should not refactor payroll yet.

Part 1 should not remove `time_entries` yet.

Part 1 should not change schedule-derived payroll calculations yet.

### Part 2: Switch Payroll Inputs

Goal: make payroll derive regular hours from `schedules` and overtime hours from `overtime_entries`.

This will require replacing current payroll time-entry queries with schedule and overtime work item queries.

### Part 3: Update Stats, Summaries, Seeds, and Frontend Contract

Goal: remove remaining business reliance on `time_entries`.

This includes employee detail stats, payroll month summaries, seed data, and any API docs/client assumptions.

### Part 4: Remove Time Entries Completely

Goal: delete the old time entry code and schema references after the app no longer depends on it.

This includes removing old routes, permissions, domain types, repositories, services, queries, generated SQLC code, and migration schema.

## Part 1 Detailed Plan

Part 1 should add the new overtime feature while leaving existing `time_entries` untouched.

**Status: COMPLETED** (2026-05-27)

All files created and wired:

| Layer | Files |
|-------|-------|
| Schema | `migrations/000001_init.up.sql` (overtime types + table), `migrations/000001_init.down.sql` (teardown) |
| SQLC Queries | `internal/repository/db/query/overtime.sql` |
| SQLC Generated | `internal/repository/db/overtime.sql.go`, updated `models.go`, `querier.go` |
| Domain | `internal/domain/overtime.go` |
| Repository | `internal/repository/overtime_repo.go` |
| Service | `internal/service/overtime_service.go` |
| Handler | `internal/handler/overtime_dto.go`, `overtime_handler.go`, `overtime_routes.go` |
| Permissions | `internal/domain/permission/permission.go`, `migrations/permissions_catalog.txt` |
| Wiring | `internal/app/app.go` (repo, service, handler, routes) |

Verification:
- `go build ./...` passes
- `go vet ./internal/...` passes
- existing tests pass (`go test ./internal/service/... ./internal/repository/... ./internal/handler/...`)

### 1. Database Schema

Add new enum types:

- `overtime_status_enum`: `submitted`, `approved`, `rejected`
- `overtime_reason_enum`: values listed above

Add table `overtime_entries`.

Important constraints:

- `employee_id` is required.
- `entry_date` is required.
- `minutes` must be greater than 0.
- `reason` is required.
- `status` defaults to `submitted`.
- `submitted_at` defaults to current timestamp.
- `rejection_reason` is required when status is `rejected` if practical at DB level, otherwise enforce in service.

Indexes:

- `(employee_id, entry_date DESC)`
- `(status)`
- `(schedule_id)`
- `(paid_period_id)`

### 2. SQLC Queries

Add `internal/repository/db/query/overtime.sql` with queries for:

- create overtime entry
- get overtime entry by ID
- list all overtime entries paginated
- list my overtime entries paginated
- lock overtime entry by ID
- approve overtime entry
- reject overtime entry
- update overtime entry by admin
- update my overtime entry while submitted and unpaid
- current month overtime stats
- my current month overtime stats

Do not remove `time_entry.sql` in Part 1.

### 3. Domain Layer

Add `internal/domain/overtime.go`.

Define:

- overtime errors
- overtime status constants
- overtime reason constants
- `OvertimeEntry`
- `OvertimeEntryPage`
- `OvertimeStats`
- create/update/list/decision params
- repository interface
- transaction repository interface
- service interface

Do not reuse `TimeEntry` types.

### 4. Repository Layer

Add `internal/repository/overtime_repo.go`.

Responsibilities:

- implement domain repository interfaces
- validate optional `schedule_id` belongs to the same employee when provided
- map SQLC rows to domain structs
- keep SQLC types inside repository/db layer

### 5. Service Layer

Add `internal/service/overtime_service.go`.

Rules:

- employee-created entries use actor employee ID
- admin-created entries require explicit employee ID
- status is `submitted` on create
- reject requires a non-empty rejection reason
- approve/reject only allowed from `submitted`
- updates should be blocked once paid
- employee self-update should only be allowed for own unpaid `submitted` entries
- minutes must be greater than 0
- reason must be one of the allowed enum values
- description should be trimmed and nullable/empty-safe based on DTO decision

### 6. Handler Layer

Add:

- `internal/handler/overtime_dto.go`
- `internal/handler/overtime_handler.go`
- `internal/handler/overtime_routes.go`

Suggested routes:

- `POST /overtime-entries`
- `POST /overtime-entries/admin`
- `POST /overtime-entries/:id/decision`
- `PUT /overtime-entries/:id/admin`
- `PUT /overtime-entries/my/:id`
- `GET /overtime-entries`
- `GET /overtime-entries/stats`
- `GET /overtime-entries/my`
- `GET /overtime-entries/my/stats`
- `GET /overtime-entries/:id`
- `GET /overtime-entries/my/:id`

DTO fields for create:

- `employee_id` only on admin create
- `schedule_id`
- `entry_date`
- `minutes`
- `reason`
- `description`

DTO fields for response:

- `id`
- `employee_id`
- `employee_name`
- `schedule_id`
- `is_paid`
- `entry_date`
- `minutes`
- `reason`
- `description`
- `status`
- `submitted_at`
- `approved_at`
- `approved_by_employee_id`
- `approved_by_name`
- `rejection_reason`
- `created_at`
- `updated_at`

### 7. Permissions

Add permission constants:

- `OVERTIME.CREATE`
- `OVERTIME.CREATE_ALL`
- `OVERTIME.DECIDE`
- `OVERTIME.UPDATE`
- `OVERTIME.UPDATE_ALL`
- `OVERTIME.VIEW`
- `OVERTIME.VIEW_ALL`

Update:

- `internal/domain/permission/permission.go`
- `migrations/permissions_catalog.txt`

Run later after implementation:

- `just permissions-check`
- `just permissions-sync`

### 8. App Wiring

Wire overtime repository, service, handler, and routes in `internal/app/app.go`.

Do not remove time entry wiring in Part 1 unless explicitly approved.

### 9. Tests

Add targeted service tests if the current test setup supports it.

Minimum behavior to test:

- create rejects nil employee ID
- create rejects invalid reason
- create rejects non-positive minutes
- reject requires rejection reason
- approve/reject only works from `submitted`
- self-update rejects another employee's entry
- paid entries cannot be updated

### 10. Verification

Expected commands after Part 1 implementation:

- regenerate SQLC
- `go test ./internal/service/...`
- `go test ./internal/repository/...` if repository tests exist or are added
- `go test ./...`
- `just permissions-check` if permissions changed

## Explicitly Out Of Scope For Part 1

- Removing `time_entries`
- Removing time entry routes
- Refactoring payroll to schedules plus overtime
- Changing `pay_period_line_items`
- Updating employee profile worked-hour stats
- Updating payroll month summaries
- Updating seed data away from time entries
- Updating Swagger artifacts unless explicitly requested

## Open Decisions Before Part 2

The following must be decided before payroll is changed:

- Should `schedule_id` be required for overtime?
- Should schedules get `paid_period_id` to prevent double payment?
- Should overtime pay base rate plus premium, or premium only?
- Should completed schedules be payable automatically, or is another approval mechanism needed?
- Should schedule-derived hours include future shifts when previewing a current or future pay period?
