#!/bin/bash

# Migration script for directory analytics database
# This script runs geni migrations for the directory database
# It reads credentials from frontends/directory/.env

set -e

# Load directory analytics database credentials
if [ ! -f "frontends/directory/.env" ]; then
    echo "Error: frontends/directory/.env not found"
    exit 1
fi

# Extract TURSO_DATABASE_URL and TURSO_AUTH_TOKEN from frontends/directory/.env
export DATABASE_URL=$(grep "^TURSO_DATABASE_URL=" frontends/directory/.env | cut -d'=' -f2)
export DATABASE_TOKEN=$(grep "^TURSO_AUTH_TOKEN=" frontends/directory/.env | cut -d'=' -f2)

if [ -z "$DATABASE_URL" ] || [ -z "$DATABASE_TOKEN" ]; then
    echo "Error: TURSO_DATABASE_URL or TURSO_AUTH_TOKEN not found in frontends/directory/.env"
    exit 1
fi

echo "Running directory analytics database migrations..."
echo "Database: $DATABASE_URL"

cd frontends/directory
DATABASE_URL="$DATABASE_URL" DATABASE_TOKEN="$DATABASE_TOKEN" geni up

echo "Directory migrations completed successfully!"
