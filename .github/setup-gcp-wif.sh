#!/bin/bash

################################################################################
# GCP Workload Identity Federation Setup Script
#
# This script automates the setup of Workload Identity Federation (WIF) for
# GitHub Actions to authenticate with GCP without storing long-lived secrets.
#
# Usage:
#   ./setup-gcp-wif.sh <project-id> <org-id> <github-repo> [region]
#
# Examples:
#   # Initial setup with org ID
#   ./setup-gcp-wif.sh cyclescene-479119 514443067230 cyclescene/cyclescene
#
#   # With custom region
#   ./setup-gcp-wif.sh cyclescene-479119 514443067230 cyclescene/cyclescene us-west1
#
# To find your organization ID:
#   gcloud organizations list
#
################################################################################

set -e

# Color output for better readability
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions for colored output
info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

header() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Load environment variables from .gcp-setup.env if it exists
ENV_FILE="$(dirname "$0")/.gcp-setup.env"
if [ -f "$ENV_FILE" ]; then
    info "Loading configuration from: $ENV_FILE"
    source "$ENV_FILE"
fi

# Parse arguments
if [ -z "$GCP_PROJECT_ID" ] && [ $# -ge 3 ]; then
    PROJECT_ID="$1"
    ORG_ID="$2"
    GITHUB_REPO="$3"
    REGION="${4:-us-west1}"
elif [ -z "$GCP_PROJECT_ID" ]; then
    error "Missing required configuration"
    echo ""
    echo "Either:"
    echo "  1. Create .gcp-setup.env in this directory with:"
    echo "     export GCP_PROJECT_ID=\"cyclescene-479119\""
    echo "     export GCP_ORG_ID=\"514443067230\""
    echo "     export GITHUB_REPO=\"cyclescene/cyclescene\""
    echo ""
    echo "  2. Or pass arguments:"
    echo "     ./setup-gcp-wif.sh <project-id> <org-id> <github-repo> [region]"
    echo ""
    echo "To find your organization ID:"
    echo "  gcloud organizations list"
    exit 1
else
    PROJECT_ID="$GCP_PROJECT_ID"
    ORG_ID="$GCP_ORG_ID"
    GITHUB_REPO="$GITHUB_REPO"
    REGION="${4:-us-west1}"
fi

# Configuration
POOL_NAME="github-pool"
PROVIDER_NAME="github-provider"
SA_NAME="github-actions"
ARTIFACT_REGISTRY_REPO="cyclescene"

SA_EMAIL="$SA_NAME@${PROJECT_ID}.iam.gserviceaccount.com"

# Display configuration
header "GCP Workload Identity Federation Setup"

echo "Configuration:"
echo "  Project ID:          $PROJECT_ID"
echo "  Organization ID:     $ORG_ID"
echo "  GitHub Repo:         $GITHUB_REPO"
echo "  Region:              $REGION"
echo "  Service Account:     $SA_EMAIL"
echo ""

# Confirm before proceeding
read -p "Continue with this configuration? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    error "Setup cancelled"
    exit 1
fi

# Set project as default
gcloud config set project "$PROJECT_ID" --quiet 2>/dev/null || true

# ============================================================================
# Step 1: Create GCP Project (if it doesn't exist)
# ============================================================================

header "Step 1: Verifying GCP Project"

info "Verifying project: $PROJECT_ID"

if gcloud projects describe "$PROJECT_ID" &>/dev/null; then
    success "Project already exists: $PROJECT_ID"
else
    warning "Project not found, attempting to create..."
    if gcloud projects create "$PROJECT_ID" \
        --organization="$ORG_ID" \
        --name="CycleScene" 2>&1; then
        success "Project created successfully"
        # Wait for project to be fully initialized
        sleep 5
    else
        error "Failed to create project in organization."
        echo ""
        echo "Please ensure:"
        echo "  1. Organization ID is correct: $ORG_ID"
        echo "  2. You have permission to create projects in this organization"
        echo "  3. You can create the project manually at: https://console.cloud.google.com/projectcreate"
        echo ""
        echo "After creating the project manually, run this script again."
        exit 1
    fi
fi

gcloud config set project "$PROJECT_ID" --quiet
success "Project set as default"

# ============================================================================
# Step 2: Enable Required APIs
# ============================================================================

header "Step 2: Enabling Required APIs"

APIS=(
    "iam.googleapis.com"
    "iamcredentials.googleapis.com"
    "cloudresourcemanager.googleapis.com"
    "sts.googleapis.com"
    "storage-component.googleapis.com"
    "run.googleapis.com"
    "artifactregistry.googleapis.com"
)

info "Enabling required APIs:"
for api in "${APIS[@]}"; do
    echo "  - $api"
done
echo ""

for api in "${APIS[@]}"; do
    info "Enabling $api..."

    if gcloud services enable "$api" \
        --project="$PROJECT_ID" \
        --quiet 2>/dev/null; then
        success "$api enabled"
    else
        warning "Failed to enable $api on first attempt, retrying..."
        sleep 2
        gcloud services enable "$api" --project="$PROJECT_ID" --quiet 2>/dev/null || true
    fi
done

success "API setup complete"

# ============================================================================
# Step 3: Create Artifact Registry
# ============================================================================

header "Step 3: Creating Artifact Registry"

info "Creating repository: $ARTIFACT_REGISTRY_REPO in $REGION"

if gcloud artifacts repositories describe $ARTIFACT_REGISTRY_REPO \
    --location=$REGION \
    --project=$PROJECT_ID >/dev/null 2>&1; then
    warning "Artifact Registry already exists"
else
    gcloud artifacts repositories create $ARTIFACT_REGISTRY_REPO \
        --repository-format=docker \
        --location=$REGION \
        --description="Docker images for CycleScene services" \
        --project=$PROJECT_ID
    success "Artifact Registry created"
fi

REGISTRY_URL="$REGION-docker.pkg.dev/$PROJECT_ID/$ARTIFACT_REGISTRY_REPO"
success "Registry URL: $REGISTRY_URL"

# ============================================================================
# Step 4: Create Workload Identity Pool
# ============================================================================

header "Step 4: Creating Workload Identity Pool"

info "Creating pool: $POOL_NAME"

if gcloud iam workload-identity-pools describe "$POOL_NAME" \
    --project="$PROJECT_ID" \
    --location=global >/dev/null 2>&1; then
    warning "Workload Identity Pool already exists"
else
    gcloud iam workload-identity-pools create "$POOL_NAME" \
        --project="$PROJECT_ID" \
        --location=global \
        --display-name="GitHub Actions" \
        --quiet
    success "Workload Identity Pool created"
fi

success "Pool verified: $POOL_NAME"

# ============================================================================
# Step 5: Create Workload Identity Provider
# ============================================================================

header "Step 5: Creating Workload Identity Provider"

info "Creating provider: $PROVIDER_NAME"
echo ""
echo "This provider configures GitHub's OIDC tokens as a trusted identity source."
echo ""

if gcloud iam workload-identity-pools providers describe-oidc "$PROVIDER_NAME" \
    --project="$PROJECT_ID" \
    --location=global \
    --workload-identity-pool="$POOL_NAME" >/dev/null 2>&1; then
    warning "Workload Identity Provider already exists"
else
    gcloud iam workload-identity-pools providers create-oidc "$PROVIDER_NAME" \
        --project="$PROJECT_ID" \
        --location=global \
        --workload-identity-pool="$POOL_NAME" \
        --display-name="GitHub" \
        --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository,attribute.aud=assertion.aud" \
        --issuer-uri="https://token.actions.githubusercontent.com" \
        --attribute-condition="assertion.aud == 'sts.googleapis.com'" \
        --quiet
    success "Workload Identity Provider created"
fi

success "Provider verified: $PROVIDER_NAME"

# ============================================================================
# Step 6: Retrieve WIF Provider Resource Name
# ============================================================================

header "Step 6: Retrieving WIF Provider Resource Name"

info "Fetching WIF provider resource name..."

WIF_PROVIDER=$(gcloud iam workload-identity-pools providers describe "$PROVIDER_NAME" \
    --project="$PROJECT_ID" \
    --location=global \
    --workload-identity-pool="$POOL_NAME" \
    --format='value(name)')

if [ -z "$WIF_PROVIDER" ]; then
    error "Could not retrieve WIF provider resource name"
    exit 1
fi

success "WIF Provider resource name retrieved"
echo ""
echo -e "  ${BLUE}$WIF_PROVIDER${NC}"
echo ""

# ============================================================================
# Step 7: Create Service Account
# ============================================================================

header "Step 7: Creating Service Account"

info "Service Account: $SA_EMAIL"
echo ""
echo "This service account will be impersonated by GitHub Actions workflows"
echo "to authenticate with GCP."
echo ""

if gcloud iam service-accounts describe "$SA_EMAIL" \
    --project="$PROJECT_ID" >/dev/null 2>&1; then
    warning "Service account already exists"
else
    gcloud iam service-accounts create "$SA_NAME" \
        --project="$PROJECT_ID" \
        --display-name="GitHub Actions Service Account" \
        --quiet
    success "Service account created"
fi

success "Service account verified: $SA_EMAIL"

# ============================================================================
# Step 8: Grant IAM Roles
# ============================================================================

header "Step 8: Granting IAM Roles"

echo "Granting roles to: $SA_EMAIL"
echo ""

ROLES=(
    "roles/run.admin"                      # Deploy to Cloud Run
    "roles/artifactregistry.writer"        # Push Docker images
    "roles/storage.admin"                  # Manage Terraform state and GCS
    "roles/iam.serviceAccountAdmin"        # Manage service accounts
    "roles/iam.securityAdmin"              # Manage IAM policies
    "roles/cloudscheduler.admin"           # Manage Cloud Scheduler
)

for role in "${ROLES[@]}"; do
    info "Granting $role..."
    gcloud projects add-iam-policy-binding "$PROJECT_ID" \
        --member="serviceAccount:$SA_EMAIL" \
        --role="$role" \
        --quiet >/dev/null 2>&1 || warning "Role may already be granted"
    success "  $role"
done

echo ""
success "All IAM roles granted"

# ============================================================================
# Step 9: Configure Workload Identity Binding
# ============================================================================

header "Step 9: Configuring Workload Identity Binding"

info "Binding WIF to GitHub repository: $GITHUB_REPO"
echo ""
echo "This allows workflows from the GitHub repository"
echo "to impersonate the service account."
echo ""

gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
    --project="$PROJECT_ID" \
    --role="roles/iam.workloadIdentityUser" \
    --member="principalSet://iam.googleapis.com/projects/$PROJECT_ID/locations/global/workloadIdentityPools/$POOL_NAME/attribute.repository/$GITHUB_REPO" \
    --quiet >/dev/null 2>&1 || warning "Binding may already be configured"

success "Workload Identity binding configured for: $GITHUB_REPO"

# ============================================================================
# Summary and Next Steps
# ============================================================================

header "✅ Setup Complete!"

echo "All GCP infrastructure for WIF is now configured."
echo ""
echo "Next Steps:"
echo ""
echo "1. Add GitHub Repository Secrets:"
echo ""
echo "   gh secret set WIF_PROVIDER --body \"$WIF_PROVIDER\""
echo "   gh secret set WIF_SERVICE_ACCOUNT --body \"$SA_EMAIL\""
echo ""
echo "   Or manually at: https://github.com/$GITHUB_REPO/settings/secrets/actions"
echo ""
echo "2. Verify the secrets are set:"
echo "   gh secret list"
echo ""
echo "Setup Summary:"
echo "  Project ID:          $PROJECT_ID"
echo "  Organization ID:     $ORG_ID"
echo "  Region:              $REGION"
echo "  Service Account:     $SA_EMAIL"
echo "  Artifact Registry:   $REGISTRY_URL"
echo "  WIF Pool:            $POOL_NAME"
echo "  WIF Provider:        $PROVIDER_NAME"
echo "  GitHub Repository:   $GITHUB_REPO"
echo ""
echo "Your GitHub Actions workflow can now authenticate to GCP"
echo "and push/deploy your services!"
echo ""
