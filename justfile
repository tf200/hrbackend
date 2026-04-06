set dotenv-load := true

up:
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE:-$DB_SOURCE}" up

down:
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE:-$DB_SOURCE}" down 1

force version:
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE:-$DB_SOURCE}" force {{version}}

up-remote:
    @if [ -z "${MIGRATION_DB_SOURCE_REMOTE}" ]; then echo "MIGRATION_DB_SOURCE_REMOTE is required"; exit 1; fi
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE_REMOTE}" up

down-remote:
    @if [ -z "${MIGRATION_DB_SOURCE_REMOTE}" ]; then echo "MIGRATION_DB_SOURCE_REMOTE is required"; exit 1; fi
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE_REMOTE}" down 1

force-remote version:
    @if [ -z "${MIGRATION_DB_SOURCE_REMOTE}" ]; then echo "MIGRATION_DB_SOURCE_REMOTE is required"; exit 1; fi
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE_REMOTE}" force {{version}}

permissions-check:
    go run ./scripts/permissions_catalog check

permissions-sync:
    go run ./scripts/permissions_catalog sync

seed-admin:
    go run ./scripts/seed_admin

seed-mock:
    go run ./scripts/seed_mock

seed-admin-remote:
    @if [ -z "${MIGRATION_DB_SOURCE_REMOTE}" ]; then echo "MIGRATION_DB_SOURCE_REMOTE is required"; exit 1; fi
    MIGRATION_DB_SOURCE="${MIGRATION_DB_SOURCE_REMOTE}" go run ./scripts/seed_admin

seed-mock-remote:
    @if [ -z "${MIGRATION_DB_SOURCE_REMOTE}" ]; then echo "MIGRATION_DB_SOURCE_REMOTE is required"; exit 1; fi
    MIGRATION_DB_SOURCE="${MIGRATION_DB_SOURCE_REMOTE}" go run ./scripts/seed_mock

lines:
    golines -w .

deploy:
    @if [ -z "${SSH_USER}" ]; then echo "SSH_USER is required (.env)"; exit 1; fi
    @if [ -z "${SSH_HOST}" ]; then echo "SSH_HOST is required (.env)"; exit 1; fi
    @if [ -z "${REMOTE_DIR}" ]; then echo "REMOTE_DIR is required (.env)"; exit 1; fi
    @if [ -z "${DEPLOY_BRANCH}" ]; then echo "DEPLOY_BRANCH is required (.env)"; exit 1; fi
    @if [ -z "${SERVICE_NAME}" ]; then echo "SERVICE_NAME is required (.env)"; exit 1; fi
    ./scripts/deploy/deploy.sh
