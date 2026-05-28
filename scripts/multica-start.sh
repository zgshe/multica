#!/bin/bash
# Multica Services Start Script
# Starts: database (docker), backend, frontend

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== Starting Multica Services ==="

# 1. Start database (docker container)
echo "[1/3] Starting database..."
if docker ps --format '{{.Names}}' | grep -q "^multica-postgres-1$"; then
    echo "  Database already running"
else
    docker start multica-postgres-1 2>/dev/null || docker start kanban-db 2>/dev/null || echo "  Warning: database container not found"
fi

# 2. Start backend
echo "[2/3] Starting backend..."
if pgrep -f "bin/server" > /dev/null 2>&1; then
    echo "  Backend already running"
else
    cd "$PROJECT_DIR/server"
    export DATABASE_URL="postgres://multica:multica@127.0.0.1:5433/multica?sslmode=disable"
    export JWT_SECRET="multica-dev-secret-change-in-production"
    export PORT=8080
    export FRONTEND_ORIGIN="http://localhost:3005"
    export APP_ENV=""
    nohup ./bin/server > /tmp/multica-server.log 2>&1 &
    echo "  Backend started (PID: $!)"
fi

# 3. Start frontend
echo "[3/3] Starting frontend..."
if pgrep -f "next-server" > /dev/null 2>&1; then
    echo "  Frontend already running"
else
    cd "$PROJECT_DIR"
    FRONTEND_PORT=3005 nohup pnpm dev:web > /tmp/multica-frontend.log 2>&1 &
    echo "  Frontend started (PID: $!)"
fi

echo ""
echo "=== Multica Services Started ==="
echo "  Frontend: http://localhost:3005"
echo "  Backend:  http://localhost:8080"
echo ""
echo "Logs:"
echo "  Backend:  tail -f /tmp/multica-server.log"
echo "  Frontend: tail -f /tmp/multica-frontend.log"