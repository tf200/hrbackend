#!/usr/bin/env bash
set -euo pipefail

SSH_USER="${SSH_USER:-root}"
SSH_HOST="${SSH_HOST:-maicare.online}"
REMOTE_DIR="${REMOTE_DIR:-hrapp/hrbackend}"
SERVICE_NAME="${SERVICE_NAME:-app}"
IMAGE_NAME="hrbackend:latest"

echo "=== [1/5] Building Docker image locally ==="
docker build --build-arg ENV_FILE=.env.dev -t "${IMAGE_NAME}" .

echo "=== [2/5] Creating remote directory on VPS ==="
ssh "${SSH_USER}@${SSH_HOST}" "mkdir -p ${REMOTE_DIR}"

echo "=== [3/5] Copying docker-compose.dev.yml to VPS ==="
scp docker-compose.dev.yml "${SSH_USER}@${SSH_HOST}:${REMOTE_DIR}/docker-compose.yml"

echo "=== [4/5] Uploading local Docker image to VPS ==="
docker save "${IMAGE_NAME}" | gzip | ssh "${SSH_USER}@${SSH_HOST}" "gunzip | docker load"

echo "=== [5/5] Recreating and starting services on VPS ==="
ssh "${SSH_USER}@${SSH_HOST}" /bin/bash <<EOF
set -euo pipefail
cd "${REMOTE_DIR}"
# Recreate only containers that have configuration or image changes (like app)
docker compose up -d
# Clean up any dangling/old images to save disk space on VPS
docker image prune -f
EOF

echo "=== Deployment Completed Successfully! ==="

