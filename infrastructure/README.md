# Infrastructure as Code with OpenTofu

This directory contains the OpenTofu/Terraform configuration for deploying Cloud Run services and jobs to GCP.

## Directory Structure

```
infrastructure/
├── main.tf                 # Main provider configuration
├── variables.tf            # Global variables
├── outputs.tf              # Global outputs
├── examples.tf             # Example usage of modules
├── terraform.tfvars.example # Example variables file
├── modules/
│   ├── cloud-run-service/  # Cloud Run service module
│   ├── cloud-run-job/      # Cloud Run job module
│   ├── service-account/    # Service account module
│   ├── storage-bucket/     # Storage bucket module
│   ├── cloud-scheduler/    # Cloud Scheduler (cron) module
│   └── pubsub-trigger/     # Pub/Sub topic and subscription module
└── environments/
    ├── dev/
    ├── staging/
    └── prod/
```

## Modules

### 1. Cloud Run Service (`modules/cloud-run-service`)

Deploys a Cloud Run service with configurable resources, scaling, and environment variables.

**Example usage:**
```hcl
module "my_service" {
  source = "./modules/cloud-run-service"

  service_name          = "my-api"
  region                = "us-central1"
  image                 = "gcr.io/my-project/my-api:latest"
  service_account_email = module.service_account.email

  env_vars = {
    NODE_ENV     = "production"
    DATABASE_URL = "postgresql://..."
    API_KEY      = "your-key"
  }

  cpu_limit    = "2"
  memory_limit = "1Gi"
  min_instances = 0
  max_instances = 10

  allow_unauthenticated = true
}
```

### 2. Cloud Run Job (`modules/cloud-run-job`)

Deploys a Cloud Run job for scheduled or on-demand batch processing.

**Example usage:**
```hcl
module "backup_job" {
  source = "./modules/cloud-run-job"

  job_name              = "daily-backup"
  region                = "us-central1"
  image                 = "gcr.io/my-project/backup:latest"
  service_account_email = module.service_account.email

  env_vars = {
    BACKUP_BUCKET = "my-backups"
    DATABASE_URL  = "postgresql://..."
  }

  cpu_limit    = "2"
  memory_limit = "2Gi"
  timeout      = "3600s"
  max_retries  = 3
}
```

### 3. Service Account (`modules/service-account`)

Creates a service account with configurable IAM roles.

**Example usage:**
```hcl
module "service_account" {
  source = "./modules/service-account"

  account_id   = "my-cloud-run-sa"
  display_name = "Cloud Run Service Account"
  project_id   = var.project_id

  roles = [
    "roles/cloudsql.client",
    "roles/secretmanager.secretAccessor",
    "roles/storage.objectViewer"
  ]
}
```

### 4. Storage Bucket (`modules/storage-bucket`)

Creates a GCS bucket with lifecycle rules, CORS, and IAM configuration.

**Example usage:**
```hcl
module "storage_bucket" {
  source = "./modules/storage-bucket"

  bucket_name    = "my-app-storage"
  location       = "US"
  storage_class  = "STANDARD"
  versioning_enabled = true

  lifecycle_rules = [
    {
      action = {
        type = "Delete"
      }
      condition = {
        age = 90
      }
    }
  ]

  iam_members = {
    "viewer" = {
      role   = "roles/storage.objectViewer"
      member = "serviceAccount:my-sa@project.iam.gserviceaccount.com"
    }
  }
}
```

### 5. Cloud Scheduler (`modules/cloud-scheduler`)

Creates Cloud Scheduler jobs to trigger Cloud Run services/jobs or Pub/Sub topics on a cron schedule.

**Example usage - Trigger Cloud Run Job:**
```hcl
module "daily_backup" {
  source = "./modules/cloud-scheduler"

  job_name    = "daily-backup"
  description = "Trigger backup job daily"
  schedule    = "0 2 * * *"  # 2 AM daily
  time_zone   = "America/New_York"
  region      = "us-central1"

  http_target = {
    uri         = "https://us-central1-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/my-project/jobs/backup-job:run"
    http_method = "POST"
    oidc_token = {
      service_account_email = module.service_account.email
    }
  }

  retry_count = 3
}
```

**Example usage - Trigger Cloud Run Service:**
```hcl
module "hourly_cleanup" {
  source = "./modules/cloud-scheduler"

  job_name    = "hourly-cleanup"
  schedule    = "0 * * * *"  # Every hour
  region      = "us-central1"

  create_service_account = true
  cloud_run_service_name = "my-service"

  http_target = {
    uri         = "https://my-service-xyz.run.app/cleanup"
    http_method = "POST"
    oidc_token = {
      service_account_email = module.service_account.email
    }
  }
}
```

**Example usage - Publish to Pub/Sub:**
```hcl
module "weekly_report" {
  source = "./modules/cloud-scheduler"

  job_name  = "weekly-report"
  schedule  = "0 9 * * MON"  # Monday 9 AM
  region    = "us-central1"

  pubsub_target = {
    topic_name = "projects/my-project/topics/reports"
    data       = base64encode(jsonencode({ type = "weekly" }))
  }
}
```

**Common Cron Schedules:**
- `0 * * * *` - Every hour
- `0 */6 * * *` - Every 6 hours
- `0 2 * * *` - Daily at 2 AM
- `0 9 * * MON` - Every Monday at 9 AM
- `0 0 1 * *` - First day of every month

### 6. Pub/Sub Trigger (`modules/pubsub-trigger`)

Creates Pub/Sub topics and subscriptions for event-driven architectures and push notifications to Cloud Run.

**Example usage - Basic Topic and Subscription:**
```hcl
module "events_topic" {
  source = "./modules/pubsub-trigger"

  topic_name          = "app-events"
  create_subscription = true
  subscription_name   = "app-events-sub"

  retry_policy = {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  publisher_service_accounts = [
    module.service_account.email
  ]
}
```

**Example usage - Push to Cloud Run:**
```hcl
module "webhook_events" {
  source = "./modules/pubsub-trigger"

  topic_name = "webhooks"

  create_subscription        = true
  push_endpoint              = "https://my-service-xyz.run.app/webhook"
  oidc_service_account_email = module.service_account.email

  ack_deadline_seconds = 30

  retry_policy = {
    minimum_backoff = "10s"
    maximum_backoff = "300s"
  }
}
```

**Example usage - With Dead Letter Queue:**
```hcl
module "critical_events" {
  source = "./modules/pubsub-trigger"

  topic_name = "critical-events"

  create_subscription   = true
  dead_letter_topic     = "projects/my-project/topics/dlq"
  max_delivery_attempts = 5

  retry_policy = {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }
}
```

## Getting Started

### Prerequisites

1. Install OpenTofu or Terraform
2. Authenticate with GCP: `gcloud auth application-default login`
3. Set your GCP project: `gcloud config set project YOUR_PROJECT_ID`

### Initial Setup

1. Create a GCS bucket for Terraform state:
   ```bash
   gsutil mb gs://your-terraform-state-bucket
   gsutil versioning set on gs://your-terraform-state-bucket
   ```

2. Update `main.tf` backend configuration:
   ```hcl
   backend "gcs" {
     bucket = "your-terraform-state-bucket"
     prefix = "terraform/state"
   }
   ```

3. Copy and configure variables:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your values
   ```

### Deployment

```bash
# Initialize OpenTofu
tofu init

# Review the plan
tofu plan

# Apply changes
tofu apply

# Destroy resources (when needed)
tofu destroy
```

## CI/CD with Cloud Build

The `cloudbuild.yaml` file in the project root automates the deployment process:

1. Builds and pushes container images
2. Initializes OpenTofu
3. Plans infrastructure changes
4. Applies changes

### Trigger Cloud Build

```bash
gcloud builds submit --config=cloudbuild.yaml \
  --substitutions=_ENVIRONMENT=dev,_REGION=us-central1,_STATE_BUCKET=your-terraform-state-bucket
```

### Environment-Specific Deployments

You can create environment-specific configurations in the `environments/` directory:

```bash
infrastructure/environments/
├── dev/
│   └── terraform.tfvars
├── staging/
│   └── terraform.tfvars
└── prod/
    └── terraform.tfvars
```

## Common IAM Roles

### Cloud Run
- `roles/run.invoker` - Invoke Cloud Run services
- `roles/run.developer` - Deploy and manage Cloud Run services
- `roles/run.admin` - Full access to Cloud Run

### Cloud Scheduler
- `roles/cloudscheduler.admin` - Manage Cloud Scheduler jobs
- `roles/cloudscheduler.jobRunner` - Run Cloud Scheduler jobs

### Pub/Sub
- `roles/pubsub.publisher` - Publish messages to topics
- `roles/pubsub.subscriber` - Subscribe to topics
- `roles/pubsub.editor` - Manage Pub/Sub resources

### Other Common Roles
- `roles/cloudsql.client` - Connect to Cloud SQL
- `roles/secretmanager.secretAccessor` - Access secrets
- `roles/storage.objectViewer` - Read from GCS buckets
- `roles/storage.objectCreator` - Write to GCS buckets
- `roles/datastore.user` - Access Firestore

## Tips

1. **Environment Variables**: Always use environment variables for sensitive data or environment-specific configuration
2. **Service Accounts**: Create separate service accounts for different services with least-privilege access
3. **State Management**: Always use remote state (GCS) for team collaboration
4. **Secrets**: Use Secret Manager for sensitive values instead of plain environment variables
5. **Resource Limits**: Set appropriate CPU and memory limits to control costs
6. **Scaling**: Configure min/max instances based on your traffic patterns

## Troubleshooting

- **Permission denied errors**: Ensure your service account has the necessary IAM roles
- **Module not found**: Run `tofu init` to download module dependencies
- **State lock errors**: Check if another process is running or manually unlock with `tofu force-unlock`
- **Image pull errors**: Verify your container image exists and the service account has access

## Additional Resources

- [OpenTofu Documentation](https://opentofu.org/docs/)
- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [GCP Service Accounts](https://cloud.google.com/iam/docs/service-accounts)
