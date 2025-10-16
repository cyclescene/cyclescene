# Example usage of all modules
# This file shows how to use each module - you can copy these examples into your main.tf

# ===================================
# Example: Service Account
# ===================================
module "cloud_run_service_account" {
  source = "./modules/service-account"

  account_id   = "my-cloud-run-sa"
  display_name = "Cloud Run Service Account"
  description  = "Service account for Cloud Run services and jobs"
  project_id   = var.project_id

  roles = [
    "roles/cloudsql.client",
    "roles/secretmanager.secretAccessor",
    "roles/storage.objectViewer"
  ]
}

# ===================================
# Example: Storage Bucket
# ===================================
module "app_storage_bucket" {
  source = "./modules/storage-bucket"

  bucket_name                 = "${var.project_id}-app-storage"
  location                    = var.region
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
  force_destroy               = false
  versioning_enabled          = true

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
      member = "serviceAccount:${module.cloud_run_service_account.email}"
    }
  }

  labels = {
    environment = var.environment
    managed_by  = "opentofu"
  }
}

# ===================================
# Example: Cloud Run Service
# ===================================
module "api_service" {
  source = "./modules/cloud-run-service"

  service_name          = "my-api-service"
  region                = var.region
  image                 = "gcr.io/${var.project_id}/my-api:latest"
  service_account_email = module.cloud_run_service_account.email

  # Environment variables - fill these in with your actual values
  env_vars = {
    NODE_ENV            = "production"
    DATABASE_URL        = "your-database-connection-string"
    API_KEY             = "your-api-key"
    STORAGE_BUCKET      = module.app_storage_bucket.bucket_name
    LOG_LEVEL           = "info"
  }

  # Resource limits
  cpu_limit              = "2"
  memory_limit           = "1Gi"
  cpu_always_allocated   = false

  # Scaling
  min_instances = 0
  max_instances = 10

  # Network
  container_port        = 8080
  allow_unauthenticated = true
  timeout               = "300s"

  labels = {
    environment = var.environment
    service     = "api"
  }
}

# ===================================
# Example: Cloud Run Job
# ===================================
module "backup_job" {
  source = "./modules/cloud-run-job"

  job_name              = "database-backup-job"
  region                = var.region
  image                 = "gcr.io/${var.project_id}/backup-job:latest"
  service_account_email = module.cloud_run_service_account.email

  # Environment variables - fill these in with your actual values
  env_vars = {
    DATABASE_URL   = "your-database-connection-string"
    BACKUP_BUCKET  = module.app_storage_bucket.bucket_name
    BACKUP_PREFIX  = "backups/"
    RETENTION_DAYS = "30"
  }

  # Resource limits
  cpu_limit    = "2"
  memory_limit = "2Gi"

  # Job configuration
  timeout     = "3600s"  # 1 hour
  max_retries = 3
  task_count  = 1
  parallelism = 1

  labels = {
    environment = var.environment
    job_type    = "backup"
  }
}

# ===================================
# Example: Another Cloud Run Service with different configuration
# ===================================
module "admin_dashboard" {
  source = "./modules/cloud-run-service"

  service_name          = "admin-dashboard"
  region                = var.region
  image                 = "gcr.io/${var.project_id}/admin-dashboard:latest"
  service_account_email = module.cloud_run_service_account.email

  # Environment variables
  env_vars = {
    NEXT_PUBLIC_API_URL = module.api_service.service_url
    SESSION_SECRET      = "your-session-secret"
    ADMIN_EMAIL         = "admin@example.com"
  }

  # Higher resources for admin dashboard
  cpu_limit    = "4"
  memory_limit = "2Gi"

  # Keep at least 1 instance warm
  min_instances = 1
  max_instances = 5

  # Requires authentication
  allow_unauthenticated = false

  labels = {
    environment = var.environment
    service     = "admin"
  }
}

# ===================================
# Example: Pub/Sub Topic and Subscription
# ===================================
module "event_topic" {
  source = "./modules/pubsub-trigger"

  topic_name          = "app-events"
  create_subscription = true
  subscription_name   = "app-events-sub"

  # Retry configuration
  retry_policy = {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  # Grant publisher access to service accounts
  publisher_service_accounts = [
    module.cloud_run_service_account.email
  ]

  labels = {
    environment = var.environment
    managed_by  = "opentofu"
  }
}

# ===================================
# Example: Pub/Sub Push Subscription to Cloud Run
# ===================================
module "webhook_topic" {
  source = "./modules/pubsub-trigger"

  topic_name = "webhook-events"

  # Push subscription configuration for Cloud Run
  create_subscription         = true
  subscription_name           = "webhook-to-cloud-run"
  push_endpoint               = module.api_service.service_url
  oidc_service_account_email  = module.cloud_run_service_account.email

  # Retry and dead letter configuration
  retry_policy = {
    minimum_backoff = "10s"
    maximum_backoff = "300s"
  }

  ack_deadline_seconds = 30

  labels = {
    environment = var.environment
    type        = "webhook"
  }
}

# ===================================
# Example: Cloud Scheduler - Trigger Cloud Run Job via HTTP
# ===================================
module "daily_backup_schedule" {
  source = "./modules/cloud-scheduler"

  job_name    = "daily-backup-trigger"
  description = "Trigger daily database backup job"
  schedule    = "0 2 * * *"  # Every day at 2 AM UTC
  time_zone   = "America/New_York"
  region      = var.region

  # HTTP target for Cloud Run Job
  http_target = {
    uri         = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${module.backup_job.job_name}:run"
    http_method = "POST"
    oidc_token = {
      service_account_email = module.cloud_run_service_account.email
    }
  }

  # Retry configuration
  retry_count          = 3
  max_retry_duration   = "0s"
  min_backoff_duration = "5s"
  max_backoff_duration = "3600s"
}

# ===================================
# Example: Cloud Scheduler - Trigger Cloud Run Service via HTTP
# ===================================
module "hourly_cleanup_schedule" {
  source = "./modules/cloud-scheduler"

  job_name    = "hourly-cleanup"
  description = "Run cleanup task every hour"
  schedule    = "0 * * * *"  # Every hour
  time_zone   = "UTC"
  region      = var.region

  # Auto-create service account and grant invoker permissions
  create_service_account = true
  cloud_run_service_name = module.api_service.service_name

  # HTTP target for Cloud Run Service
  http_target = {
    uri         = "${module.api_service.service_url}/api/cleanup"
    http_method = "POST"
    headers = {
      "Content-Type" = "application/json"
    }
    body = base64encode(jsonencode({
      task = "cleanup"
      mode = "automatic"
    }))
    oidc_token = {
      service_account_email = module.cloud_run_service_account.email
      audience              = module.api_service.service_url
    }
  }

  retry_count = 2
}

# ===================================
# Example: Cloud Scheduler - Publish to Pub/Sub
# ===================================
module "weekly_report_schedule" {
  source = "./modules/cloud-scheduler"

  job_name    = "weekly-report"
  description = "Trigger weekly report generation"
  schedule    = "0 9 * * MON"  # Every Monday at 9 AM
  time_zone   = "America/New_York"
  region      = var.region

  # Pub/Sub target
  pubsub_target = {
    topic_name = module.event_topic.topic_id
    data       = base64encode(jsonencode({
      report_type = "weekly"
      recipients  = ["team@example.com"]
    }))
    attributes = {
      type = "report"
      frequency = "weekly"
    }
  }

  retry_count = 1
}
