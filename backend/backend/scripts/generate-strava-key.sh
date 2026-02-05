#!/bin/bash
# Generate a secure 32-byte encryption key for Strava token storage
# Usage: ./scripts/generate-strava-key.sh

set -e

echo "Generating STRAVA_TOKEN_ENCRYPTION_KEY..."
echo ""

# Generate 32 random bytes and encode as base64
KEY=$(openssl rand -base64 32)

echo "✓ Generated encryption key"
echo ""
echo "Add this to your environment (GitHub Secrets, .env, etc.):"
echo ""
echo "STRAVA_TOKEN_ENCRYPTION_KEY=\"$KEY\""
echo ""
echo "For GitHub Actions, add as a secret:"
echo "  Name: TF_VAR_strava_token_encryption_key"
echo "  Value: $KEY"
echo ""
echo "For local development, add to backend/.env:"
echo "  STRAVA_TOKEN_ENCRYPTION_KEY=\"$KEY\""
echo ""
echo "⚠️  IMPORTANT: Store this securely and never commit to git!"
echo "⚠️  If you lose this key, all stored refresh tokens become unusable."
