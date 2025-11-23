# CI/CD Setup Guide - GitHub Actions + GCP WIF

This guide walks you through setting up the complete CI/CD pipeline for CycleScene using GitHub Actions and Google Cloud Platform with Workload Identity Federation (WIF).

## Overview

The CI/CD pipeline will:

1. **Detect changed services** - Only build/push/deploy services that have changes
2. **Test and build** - Run Go tests, linting, and build binaries
3. **Push to Artifact Registry** - Push Docker images with proper tagging
4. **Deploy to Cloud Run** - Use OpenTofu to deploy infrastructure on main branch

### Services Included

- api
- image-optimizer
- scraperv2
- token-cleaner
- db-backups

## Prerequisites

- Active GCP account with a new project (`cyclescene-479119`)
- `gcloud` CLI installed and authenticated
- Access to GitHub repo: `cyclescene/cyclescene`
- Permissions to create:
  - Service Accounts
  - Workload Identity Pools
  - IAM role bindings
  - Artifact Registry repositories

## Setup Instructions

### Step 1: Run the Setup Script

The setup script automates all GCP infrastructure configuration:

```bash
cd .github
chmod +x setup-gcp-wif.sh
./setup-gcp-wif.sh
```

This script will:
- Enable required Google APIs
- Create Artifact Registry repository
- Create Workload Identity Pool and OIDC Provider
- Create service account with proper IAM roles
- Configure WIF bindings for GitHub

### Step 2: Add GitHub Secrets

After running the setup script, it will output two values:

1. `WIF_PROVIDER` - The Workload Identity Provider resource name
2. `WIF_SERVICE_ACCOUNT` - The service account email

Add these as repository secrets:

1. Go to: https://github.com/cyclescene/cyclescene/settings/secrets/actions

#### Required Authentication Secrets
2. Click "New repository secret"
3. Add:
   - **Name:** `WIF_PROVIDER`
   - **Value:** `projects/XXXXX/locations/us-west1/workloadIdentityPools/github-actions-pool/providers/github-provider`

4. Click "New repository secret"
5. Add:
   - **Name:** `WIF_SERVICE_ACCOUNT`
   - **Value:** `github-actions@cyclescene-479119.iam.gserviceaccount.com`

#### Service Secrets

**Database (Turso)**
6. Click "New repository secret"
7. Add:
   - **Name:** `TURSO_DB_URL`
   - **Value:** `libsql://cyclescene-spacesedan.aws-us-west-2.turso.io`

8. Click "New repository secret"
9. Add:
   - **Name:** `TURSO_DB_RW_TOKEN`
   - **Value:** Your Turso read-write token

10. Click "New repository secret"
11. Add:
   - **Name:** `TURSO_DB_RO_TOKEN`
   - **Value:** Your Turso read-only token

**Cloud Storage & Email**
12. Click "New repository secret"
13. Add:
   - **Name:** `STAGING_BUCKET_NAME`
   - **Value:** `cyclescene-479119-user-media-staging`

14. Click "New repository secret"
15. Add:
   - **Name:** `RESEND_API_KEY`
   - **Value:** Your Resend API key (for magic link emails)

### Local Development Setup

For local development, update the secrets in your terraform.tfvars files:

**API Service** (`functions/cmd/api/infra/terraform.tfvars`):
```hcl
turso_db_url       = "libsql://cyclescene-spacesedan.aws-us-west-2.turso.io"
turso_db_rw_token  = "your_actual_turso_rw_token"
staging_bucket_name = "cyclescene-user-media-staging"
resend_api_key     = "your_actual_resend_api_key"
```

**Scraper, Token Cleaner, DB Backups** (respective `terraform.tfvars`):
```hcl
env_vars = {
  TURSO_DB_URL       = "libsql://cyclescene-spacesedan.aws-us-west-2.turso.io"
  TURSO_DB_RW_TOKEN  = "your_actual_turso_rw_token"
  TURSO_DB_RO_TOKEN  = "your_actual_turso_ro_token"
  # ... other vars
}
```

When deploying locally:
```bash
cd functions/cmd/{service}/infra
tofu init
tofu apply
```

When deploying via CI/CD, all secrets will be passed via environment variables and command-line arguments from GitHub secrets.

### Step 3: Verify Workflow

The workflow is configured in `.github/workflows/go-services-ci.yml`

**Key behaviors:**

- **Pull Requests to main**: Run tests and linting, but don't push/deploy
- **Push to dev/* branches**: Build, test, push to registry (no deploy)
- **Push to main**: Build, test, push to registry, AND deploy to Cloud Run

### Step 4: Test the Pipeline

1. **Create a test branch:**
   ```bash
   git checkout -b dev/test-ci
   ```

2. **Make a small change to a service** (e.g., update a comment in `functions/cmd/api/handler.go`)

3. **Push and monitor:**
   ```bash
   git add .
   git commit -m "test: ci/cd pipeline"
   git push -u origin dev/test-ci
   ```

4. **Check GitHub Actions:**
   - Go to: https://github.com/cyclescene/cyclescene/actions
   - You should see the workflow running
   - It will build, test, lint, and push images (no deploy on dev/* branch)

5. **Verify images in Artifact Registry:**
   ```bash
   gcloud artifacts docker images list us-west1-docker.pkg.dev/cyclescene-479119/cyclescene \
     --project=cyclescene-479119
   ```

## Workflow Behavior

### Change Detection Logic

The workflow automatically detects which services changed by checking if:
- Files in `functions/cmd/{service}/` were modified
- Shared code in `functions/internal/` was modified
- Go module files changed
- The workflow file itself changed

Only those services will be built and pushed.

### Image Tagging Strategy

- **main branch**: `latest`, `short-commit-sha` (e.g., `abc1234`)
- **dev/* branches**: `branch-name`, `short-commit-sha` (e.g., `dev-test-ci-abc1234`)

### Deployment

Deployment only happens when:
1. Changes are pushed to `main` branch
2. The `push-to-registry` job succeeds
3. Images are deployed using OpenTofu in each service's `infra/` directory

## Manual Deployment

If you need to deploy manually:

```bash
cd functions/cmd/{service}/infra
tofu init
tofu apply
```

## Troubleshooting

### Authentication Failures

If you see "Authenticate to Google Cloud" failures:

1. Verify secrets are set correctly:
   ```bash
   gh secret list --repo cyclescene/cyclescene
   ```

2. Check WIF setup:
   ```bash
   gcloud iam workload-identity-pools describe github-actions-pool \
     --location=us-west1 \
     --project=cyclescene-479118
   ```

3. Verify service account has correct roles:
   ```bash
   gcloud projects get-iam-policy cyclescene-479119 \
     --flatten="bindings[].members" \
     --filter="bindings.members:github-actions@*"
   ```

### Deployment Failures

Check Cloud Run logs:
```bash
gcloud run services describe {service-name} \
  --region us-west1 \
  --project cyclescene-479118
```

Check Terraform state:
```bash
cd functions/cmd/{service}/infra
tofu state list
tofu show
```

### Image Not Found in Registry

Verify the image was pushed:
```bash
gcloud artifacts docker images list us-west1-docker.pkg.dev/cyclescene-479119/cyclescene \
  --project=cyclescene-479119
```

Or for a specific service:
```bash
gcloud artifacts docker images describe \
  us-west1-docker.pkg.dev/cyclescene-479119/cyclescene/api/api-image:latest \
  --project=cyclescene-479119
```

## Environment Variables

All environment variables are defined in the workflow:

```yaml
env:
  GCP_PROJECT_ID: cyclescene-479119
  ARTIFACT_REGISTRY_REGION: us-west1
  ARTIFACT_REGISTRY_REPO: cyclescene
```

These map to image paths like:
```
us-west1-docker.pkg.dev/cyclescene-479119/cyclescene/{service}/{service}-image:{tag}
```

## Cost Optimization Tips

1. **Parallel builds**: The workflow builds multiple services in parallel (up to 3)
2. **Sequential deployments**: Services deploy one at a time to avoid quota issues
3. **Change detection**: Only changed services are built/deployed (saves cost)
4. **Docker caching**: Each step uses layer caching to speed up builds

## Advanced Configuration

### Limiting Parallel Deployments

In the `deploy-services` job, adjust:
```yaml
max-parallel: 1  # Change to 2 or 3 if needed
```

### Adding Environment-Specific Deployments

To support dev/staging/production, create separate jobs:

```yaml
deploy-staging:
  if: startsWith(github.ref, 'refs/heads/dev/')
  # Deploy to staging environment

deploy-production:
  if: github.ref == 'refs/heads/main'
  # Deploy to production
```

### Adding Frontend Builds

Create `.github/workflows/frontend-ci.yml` for Svelte apps:

```yaml
name: Frontend CI/CD

on:
  push:
    paths:
      - 'frontends/**'
```

## Support

For issues or questions, check:
- GitHub Actions logs: https://github.com/cyclescene/cyclescene/actions
- GCP Cloud Run logs: https://console.cloud.google.com/run
- Terraform state: `functions/cmd/{service}/infra/terraform.tfstate`
