#!/bin/bash
# Test script for strava-sync service
# Usage: ./test_sync.sh [--use-real-key] [--force]
#
# Options:
#   --use-real-key   Use the encryption key from .env (for real testing)
#   --force          Force sync all connections (ignore 3-day interval)
#   (default)        Use a TEST encryption key (will fail to decrypt real data)

set -e

cd "$(dirname "$0")/../.."

# Load .env file if it exists
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

# Check for required env vars
if [ -z "$TURSO_DB_URL" ]; then
    echo "ERROR: TURSO_DB_URL not set. Please configure your .env file."
    exit 1
fi

if [ -z "$STRAVA_CLIENT_ID" ] || [ -z "$STRAVA_CLIENT_SECRET" ]; then
    echo "ERROR: STRAVA_CLIENT_ID and STRAVA_CLIENT_SECRET must be set."
    exit 1
fi

# Enable debug logging
export STRAVA_DEBUG=true
export APP_ENV=dev

# Lower limits for testing
export SYNC_MAX_CONNECTIONS=5
export SYNC_MAX_REQUESTS_15MIN=20

echo "========================================"
echo "  Strava Sync Test"
echo "========================================"
echo ""

USE_REAL_KEY=false
FORCE_SYNC=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        --use-real-key)
            USE_REAL_KEY=true
            ;;
        --force)
            FORCE_SYNC=true
            ;;
    esac
done

if [ "$USE_REAL_KEY" == "true" ]; then
    if [ -z "$STRAVA_TOKEN_ENCRYPTION_KEY" ]; then
        echo "ERROR: STRAVA_TOKEN_ENCRYPTION_KEY not set in .env"
        exit 1
    fi
    echo "Using REAL encryption key from .env"
    echo "This will sync actual Strava connections."
else
    # Use TEST encryption key (32 bytes of zeros, base64 encoded)
    export STRAVA_TOKEN_ENCRYPTION_KEY="AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
    echo "Using TEST encryption key (will fail to decrypt real data)"
    echo "Use --use-real-key to test with actual data."
fi

if [ "$FORCE_SYNC" == "true" ]; then
    export SYNC_FORCE=true
    echo "FORCE MODE: Will sync all connections (ignoring 3-day interval)"
fi

echo ""
echo "Configuration:"
echo "  Max connections: $SYNC_MAX_CONNECTIONS"
echo "  Max requests/15min: $SYNC_MAX_REQUESTS_15MIN"
echo "  Debug: $STRAVA_DEBUG"
echo "  Force sync: ${SYNC_FORCE:-false}"
echo ""
echo "Starting sync..."
echo ""
echo "----------------------------------------"

go run ./cmd/strava-sync

echo "----------------------------------------"
echo ""
echo "Sync complete!"
echo ""
echo "========================================"
