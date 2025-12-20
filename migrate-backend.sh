#!/bin/bash

# Migration script for backend database
# This script runs geni migrations for the backend database
# It reads credentials from backend/.env

set -e

# Load backend database credentials
if [ ! -f "backend/.env" ]; then
    echo "Error: backend/.env not found"
    exit 1
fi

# Extract DATABASE_URL and DATABASE_TOKEN from backend/.env
export DATABASE_URL=$(grep "^TURSO_DB_URL=" backend/.env | cut -d'=' -f2)
export DATABASE_TOKEN=$(grep "^TURSO_DB_RW_TOKEN=" backend/.env | cut -d'=' -f2)

if [ -z "$DATABASE_URL" ] || [ -z "$DATABASE_TOKEN" ]; then
    echo "Error: TURSO_DB_URL or TURSO_DB_RW_TOKEN not found in backend/.env"
    exit 1
fi

echo "Running backend database migrations..."
echo "Database: $DATABASE_URL"

cd db/main
DATABASE_URL="$DATABASE_URL" DATABASE_TOKEN="$DATABASE_TOKEN" geni up

echo "Backend migrations completed successfully!"
