#!/usr/bin/env bash
set -euo pipefail

SSH_USER="${SSH_USER:-root}"
SSH_HOST="${SSH_HOST:-maicare.online}"
REMOTE_DIR="${REMOTE_DIR:-hrapp/hrbackend}"
DEPLOY_BRANCH="${DEPLOY_BRANCH:-master}"
SERVICE_NAME="${SERVICE_NAME:-app}"

ssh "${SSH_USER}@${SSH_HOST}" /bin/bash <<EOF
set -euo pipefail

cd "${REMOTE_DIR}"
git pull origin "${DEPLOY_BRANCH}"
docker compose stop "${SERVICE_NAME}"
docker compose rm -f "${SERVICE_NAME}"
docker compose up -d --build "${SERVICE_NAME}"
EOF
