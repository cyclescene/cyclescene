#!/bin/bash

# Migration script for monitoring database
# This script runs geni migrations for the monitoring database
# It reads credentials from backend/.env

set -e

# Load monitoring database credentials
if [ ! -f "backend/.env" ]; then
    echo "Error: backend/.env not found"
    exit 1
fi

# Extract monitoring database credentials from backend/.env
export DATABASE_URL=$(grep "^TURSO_MONITORING_DB_URL=" backend/.env | cut -d'=' -f2)
export DATABASE_TOKEN=$(grep "^TURSO_MONITORING_DB_RW_TOKEN=" backend/.env | cut -d'=' -f2)

if [ -z "$DATABASE_URL" ] || [ -z "$DATABASE_TOKEN" ]; then
    echo "Error: TURSO_MONITORING_DB_URL or TURSO_MONITORING_DB_RW_TOKEN not found in backend/.env"
    echo ""
    echo "Please set up the monitoring database first:"
    echo "1. Create database: turso db create cycle-scene-monitoring"
    echo "2. Get credentials: turso db show cycle-scene-monitoring"
    echo "3. Add to backend/.env:"
    echo "   TURSO_MONITORING_DB_URL=libsql://..."
    echo "   TURSO_MONITORING_DB_RW_TOKEN=eyJ..."
    exit 1
fi

echo "Running monitoring database migrations..."
echo "Database: $DATABASE_URL"

cd db/monitoring
DATABASE_URL="$DATABASE_URL" DATABASE_TOKEN="$DATABASE_TOKEN" geni up

echo "Monitoring database migrations completed successfully!"
