set dotenv-load := true

up:
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE:-$DB_SOURCE}" up

down:
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE:-$DB_SOURCE}" down 1

force version:
    migrate -path ./migrations -database "${MIGRATION_DB_SOURCE:-$DB_SOURCE}" force {{version}}
