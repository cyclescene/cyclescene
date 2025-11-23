# Infrastructure Modules

This directory contains shared OpenTofu/Terraform modules for deploying Cloud Run services and jobs to GCP.

## Architecture

This project uses a **co-located infrastructure** approach where each service has its own `infra/` directory containing Terraform configuration that references these shared modules.

```
cycle-scene/
├── functions/
│   └── cmd/
│       ├── api/
│       │   ├── main.go
│       │   ├── Dockerfile
│       │   ├── Makefile
│       │   └── infra/              # API infrastructure
│       │       ├── main.tf
│       │       ├── variables.tf
│       │       └── terraform.tfvars
│       ├── token-cleaner/
│       │   └── infra/              # Token cleaner infrastructure
│       ├── db-backups/
│       │   └── infra/              # Backup infrastructure
│       └── scraperv2/
│           └── infra/              # Scraper infrastructure
└── infrastructure/
    └── modules/                    # Shared modules (this directory)
        ├── cloud-run-service/
        ├── cloud-run-job/
        ├── service-account/
        ├── storage-bucket/
        ├── cloud-scheduler/
        └── pubsub-trigger/
```

## Deployment Workflow

Each service deploys independently from its own directory:

```bash
# Deploy a specific service
cd functions/cmd/api
make deploy-all  # Builds Docker image, pushes to registry, and deploys infrastructure

# Or deploy just the infrastructure
cd functions/cmd/api/infra
tofu init
tofu apply
```

## Available Modules

### 1. Cloud Run Service (`modules/cloud-run-service`)

Deploys a Cloud Run service with configurable resources, scaling, and environment variables.

**Example usage:**
```hcl
module "api_service" {
  source = "../../../../infrastructure/modules/cloud-run-service"

  service_name          = "cyclescene-api-gateway"
  region                = var.region
  image                 = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/api/api-image:latest"
  service_account_email = module.service_account.email

  env_vars = {
    TURSO_DB_URL     = "libsql://..."
    TURSO_DB_RW_TOKEN = "..."
    PROJECT_ID       = var.project_id
  }

  cpu_limit    = "2"
  memory_limit = "1Gi"
  min_instances = 0
  max_instances = 10

  allow_unauthenticated = true
  container_port        = 8080
}
```

**Key features:**
- Auto-scaling configuration
- Environment variable management
- Resource limits (CPU/memory)
- Public or authenticated access
- Custom port configuration

### 2. Cloud Run Job (`modules/cloud-run-job`)

Deploys a Cloud Run job for scheduled or on-demand batch processing.

**Example usage:**
```hcl
module "backup_job" {
  source = "../../../../infrastructure/modules/cloud-run-job"

  job_name              = "turso-backup-job"
  region                = var.region
  image                 = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/backup/backup-image:latest"
  service_account_email = module.service_account.email

  env_vars = {
    TURSO_DB_URL     = "libsql://..."
    BACKUP_BUCKET    = module.storage_bucket.bucket_name
  }

  cpu_limit    = "2"
  memory_limit = "2Gi"
  timeout      = "1800s"  # 30 minutes
  max_retries  = 3
}
```

**Key features:**
- Configurable timeout and retries
- Resource limits
- Task concurrency control
- Environment variable support

### 3. Service Account (`modules/service-account`)

Creates a service account with configurable IAM roles.

**Example usage:**
```hcl
module "api_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "cyclescene-api"
  display_name = "API Service Account"
  project_id   = var.project_id

  roles = [
    "roles/storage.objectViewer",
    "roles/storage.objectCreator",
    "roles/iam.serviceAccountTokenCreator"  # For signing URLs
  ]
}

# Use in other resources
module "api_service" {
  source = "...modules/cloud-run-service"
  service_account_email = module.api_service_account.email
}
```

**Common role combinations:**
- **API with storage access**: `objectViewer`, `objectCreator`, `serviceAccountTokenCreator`
- **Backup jobs**: `objectViewer`, `objectCreator`
- **Scheduler invoker**: `run.invoker`

### 4. Storage Bucket (`modules/storage-bucket`)

Creates a GCS bucket with lifecycle rules, CORS, and IAM configuration.

**Example usage:**
```hcl
module "backup_bucket" {
  source = "../../../../infrastructure/modules/storage-bucket"

  bucket_name        = "${var.project_id}-db-backups"
  location           = "US"
  storage_class      = "STANDARD"
  versioning_enabled = true

  lifecycle_rules = [{
    action = { type = "Delete" }
    condition = { age = 30 }  # Delete after 30 days
  }]
}

# Use bucket in environment variables
module "backup_job" {
  source = "...modules/cloud-run-job"
  env_vars = {
    BACKUP_BUCKET = module.backup_bucket.bucket_name
  }
}
```

**Example with CORS for user uploads:**
```hcl
module "user_media_bucket" {
  source = "...modules/storage-bucket"

  bucket_name = "${var.project_id}-user-media"

  cors_rules = [{
    origin          = ["https://cyclescene.com"]
    method          = ["GET", "POST", "PUT", "HEAD"]
    response_header = ["Content-Type", "Content-Length"]
    max_age_seconds = 3600
  }]

  lifecycle_rules = [{
    action = {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
    condition = { age = 90 }  # Move to cheaper storage after 90 days
  }]
}
```

### 5. Cloud Scheduler (`modules/cloud-scheduler`)

Creates Cloud Scheduler jobs to trigger Cloud Run services/jobs on a cron schedule.

**Example usage - Trigger Cloud Run Job:**
```hcl
# Create service account for scheduler
module "scheduler_service_account" {
  source = "...modules/service-account"

  account_id = "backup-scheduler"
  project_id = var.project_id
  roles      = ["roles/run.invoker"]
}

# Grant permission to invoke the job
resource "google_cloud_run_v2_job_iam_member" "scheduler_invoker" {
  name     = module.backup_job.job_name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${module.scheduler_service_account.email}"
}

# Create the schedule
module "daily_backup" {
  source = "...modules/cloud-scheduler"

  job_name    = "daily-backup"
  description = "Trigger backup job daily at 2 AM"
  schedule    = "0 2 * * *"
  time_zone   = "UTC"
  region      = var.region

  http_target = {
    uri         = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${module.backup_job.job_name}:run"
    http_method = "POST"
    oidc_token = {
      service_account_email = module.scheduler_service_account.email
    }
  }

  retry_count = 3
}
```

**Common Cron Schedules:**
- `0 * * * *` - Every hour
- `0 */6 * * *` - Every 6 hours
- `0 2 * * *` - Daily at 2 AM
- `0 9 * * MON` - Every Monday at 9 AM
- `0 0 1 * *` - First day of every month

**Important**: Always create a dedicated service account for schedulers and grant explicit `run.invoker` permissions. The default App Engine service account may not have sufficient permissions.

### 6. Pub/Sub Trigger (`modules/pubsub-trigger`)

Creates Pub/Sub topics and subscriptions for event-driven architectures.

**Example usage - Push to Cloud Run:**
```hcl
module "webhook_events" {
  source = "../../../../infrastructure/modules/pubsub-trigger"

  topic_name = "webhook-events"

  create_subscription        = true
  subscription_name          = "webhook-processor"
  push_endpoint              = "https://my-service-xyz.run.app/webhook"
  oidc_service_account_email = module.service_account.email

  ack_deadline_seconds = 30

  retry_policy = {
    minimum_backoff = "10s"
    maximum_backoff = "300s"
  }
}
```

**Example with Dead Letter Queue:**
```hcl
module "critical_events" {
  source = "...modules/pubsub-trigger"

  topic_name = "critical-events"

  create_subscription   = true
  dead_letter_topic     = "projects/${var.project_id}/topics/dlq"
  max_delivery_attempts = 5

  retry_policy = {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }
}
```

## Service Examples

### API Service (Cloud Run Service)

```hcl
# functions/cmd/api/infra/main.tf
module "api_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id = "cyclescene-api"
  project_id = var.project_id

  roles = [
    "roles/storage.objectCreator",
    "roles/storage.objectViewer",
    "roles/iam.serviceAccountTokenCreator"  # For signed URLs
  ]
}

module "user_media_bucket" {
  source = "../../../../infrastructure/modules/storage-bucket"

  bucket_name = "${var.project_id}-user-media"

  cors_rules = [{
    origin = var.allowed_origins
    method = ["GET", "POST", "PUT", "HEAD"]
    response_header = ["Content-Type"]
    max_age_seconds = 3600
  }]
}

module "api_service" {
  source = "../../../../infrastructure/modules/cloud-run-service"

  service_name          = "cyclescene-api-gateway"
  image                 = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/api/api-image:latest"
  service_account_email = module.api_service_account.email

  env_vars = merge(var.env_vars, {
    MEDIA_BUCKET = module.user_media_bucket.bucket_name
    PROJECT_ID   = var.project_id
  })

  container_port        = 8080
  allow_unauthenticated = true
}
```

### Scheduled Job (Cloud Run Job + Cloud Scheduler)

```hcl
# functions/cmd/token-cleaner/infra/main.tf

# Service account for the job itself
module "job_service_account" {
  source = "...modules/service-account"

  account_id = "token-cleaner"
  project_id = var.project_id
  roles      = []  # Add roles as needed
}

# Service account for scheduler to invoke the job
module "scheduler_service_account" {
  source = "...modules/service-account"

  account_id = "token-cleaner-scheduler"
  project_id = var.project_id
  roles      = ["roles/run.invoker"]
}

module "token_cleaner_job" {
  source = "...modules/cloud-run-job"

  job_name              = "submission-token-cleaner"
  image                 = "..."
  service_account_email = module.job_service_account.email
  env_vars              = var.env_vars
}

# Grant scheduler permission to invoke the job
resource "google_cloud_run_v2_job_iam_member" "scheduler_invoker" {
  name     = module.token_cleaner_job.job_name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${module.scheduler_service_account.email}"
}

module "token_cleaner_schedule" {
  source = "...modules/cloud-scheduler"

  job_name  = "token-cleaner-daily"
  schedule  = "0 0 * * *"  # Daily at midnight
  time_zone = "UTC"

  http_target = {
    uri = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${module.token_cleaner_job.job_name}:run"
    http_method = "POST"
    oidc_token = {
      service_account_email = module.scheduler_service_account.email
    }
  }
}
```

### Backup Job with Storage

```hcl
# functions/cmd/db-backups/infra/main.tf

module "backup_service_account" {
  source = "...modules/service-account"

  account_id = "cyclescene-backup-uploader"
  project_id = var.project_id

  roles = [
    "roles/storage.objectViewer",
    "roles/storage.objectCreator"
  ]
}

module "backup_storage_bucket" {
  source = "...modules/storage-bucket"

  bucket_name        = "${var.project_id}-db-backups"
  versioning_enabled = true

  lifecycle_rules = [{
    action    = { type = "Delete" }
    condition = { age = 30 }
  }]
}

module "db_backup_job" {
  source = "...modules/cloud-run-job"

  job_name              = "turso-backup-job"
  service_account_email = module.backup_service_account.email

  env_vars = merge(var.env_vars, {
    BACKUP_BUCKET = module.backup_storage_bucket.bucket_name
  })

  timeout = "1800s"  # 30 minutes
}
```

## Getting Started

### Prerequisites

1. Install OpenTofu: `brew install opentofu` (or see [OpenTofu installation](https://opentofu.org/docs/intro/install/))
2. Authenticate with GCP: `gcloud auth application-default login`
3. Set your GCP project: `gcloud config set project YOUR_PROJECT_ID`

### Creating a New Service

1. Create the service directory structure:
   ```bash
   mkdir -p functions/cmd/my-service/infra
   ```

2. Create `infra/main.tf` referencing shared modules:
   ```hcl
   module "my_service" {
     source = "../../../../infrastructure/modules/cloud-run-service"
     # ... configuration
   }
   ```

3. Create `infra/variables.tf` and `infra/terraform.tfvars`

4. Deploy:
   ```bash
   cd functions/cmd/my-service/infra
   tofu init
   tofu apply
   ```

### State Management

Each service manages its own Terraform state. Consider using remote state in GCS for team collaboration:

```hcl
# infra/main.tf
terraform {
  backend "gcs" {
    bucket = "cyclescene-terraform-state"
    prefix = "services/my-service"
  }
}
```

## Common IAM Roles

### Cloud Run
- `roles/run.invoker` - Invoke Cloud Run services/jobs
- `roles/run.developer` - Deploy and manage Cloud Run services
- `roles/run.admin` - Full access to Cloud Run

### Storage
- `roles/storage.objectViewer` - Read from GCS buckets
- `roles/storage.objectCreator` - Write to GCS buckets
- `roles/iam.serviceAccountTokenCreator` - Sign URLs/tokens

### Cloud Scheduler
- `roles/cloudscheduler.admin` - Manage Cloud Scheduler jobs

### Pub/Sub
- `roles/pubsub.publisher` - Publish messages to topics
- `roles/pubsub.subscriber` - Subscribe to topics

## Tips

1. **Environment Variables**: Use `env_vars` map for configuration
2. **Service Accounts**: Create separate service accounts per service with least-privilege access
3. **Secrets**: Use Secret Manager for sensitive values instead of plain environment variables
4. **Resource Limits**: Set appropriate CPU and memory limits to control costs
5. **Scaling**: Configure min/max instances based on traffic patterns (API: 0-10, scheduled jobs: 0)
6. **Timeouts**: Set appropriate timeouts for jobs (default 600s, max 3600s)
7. **Scheduler Permissions**: Always create dedicated service accounts for schedulers with explicit IAM bindings

## Troubleshooting

### Permission Denied on Scheduled Jobs
If Cloud Scheduler cannot invoke your job:
1. Verify scheduler service account has `roles/run.invoker` role
2. Check IAM binding on the Cloud Run job/service
3. Ensure OIDC token is configured with correct service account email

### Docker Build Issues
If Docker can't find `go.mod`:
1. Verify Dockerfile build context is correct (should be `../..` from service directory)
2. Check Dockerfile COPY paths match your directory structure

### Terraform State Lock
If state is locked:
```bash
tofu force-unlock LOCK_ID
```

## Additional Resources

- [OpenTofu Documentation](https://opentofu.org/docs/)
- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Cloud Scheduler Documentation](https://cloud.google.com/scheduler/docs)
- [GCP Service Accounts](https://cloud.google.com/iam/docs/service-accounts)
