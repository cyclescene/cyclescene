#!/bin/bash

# Migration script for marketing analytics database
# This script runs geni migrations for the directory analytics database
# It reads credentials from frontends/directory/.env

set -e

# Load marketing analytics database credentials
if [ ! -f "frontends/directory/.env" ]; then
    echo "Error: frontends/directory/.env not found"
    exit 1
fi

# Extract ANALYTICS_DB_URL and ANALYTICS_DB_TOKEN from frontends/directory/.env
export DATABASE_URL=$(grep "^ANALYTICS_DB_URL=" frontends/directory/.env | cut -d'=' -f2)
export DATABASE_TOKEN=$(grep "^ANALYTICS_DB_TOKEN=" frontends/directory/.env | cut -d'=' -f2)

if [ -z "$DATABASE_URL" ] || [ -z "$DATABASE_TOKEN" ]; then
    echo "Error: ANALYTICS_DB_URL or ANALYTICS_DB_TOKEN not found in frontends/directory/.env"
    exit 1
fi

echo "Running marketing analytics database migrations..."
echo "Database: $DATABASE_URL"

cd frontends/directory
DATABASE_URL="$DATABASE_URL" DATABASE_TOKEN="$DATABASE_TOKEN" geni up

echo "Marketing analytics migrations completed successfully!"
