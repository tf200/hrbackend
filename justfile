set dotenv-load := true

up:
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE:-$DB_SOURCE}" up

down:
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE:-$DB_SOURCE}" down 1

force version:
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE:-$DB_SOURCE}" force {{version}}

permissions-check:
    go run ./scripts/permissions_catalog check

permissions-sync:
    go run ./scripts/permissions_catalog sync

seed-admin:
    go run ./scripts/seed_admin

lines:
    golines -w .
