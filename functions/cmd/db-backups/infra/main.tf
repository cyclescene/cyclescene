terraform {
  required_version = ">= 1.6"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "gcs" {}
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Service Account for DB backups with storage permissions
module "backup_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "cyclescene-backup-uploader"
  display_name = "DB Backup Service Account"
  description  = "Service account for database backup job with GCS access"
  project_id   = var.project_id

  roles = [
    "roles/storage.objectViewer",
    "roles/storage.objectCreator"
  ]
}

# Storage bucket for database backups
module "backup_storage_bucket" {
  source = "../../../../infrastructure/modules/storage-bucket"

  bucket_name                 = "${var.project_id}-db-backups"
  location                    = var.region
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
  force_destroy               = false
  versioning_enabled          = true

  # Lifecycle rule to delete old backups after 30 days
  lifecycle_rules = [
    {
      action = {
        type = "Delete"
      }
      condition = {
        age = 30
      }
    }
  ]

  labels = {
    environment = var.environment
    purpose     = "database-backups"
    managed_by  = "opentofu"
  }
}

# Cloud Run Job for database backups
module "db_backup_job" {
  source = "../../../../infrastructure/modules/cloud-run-job"

  job_name              = "turso-backup-job"
  image                 = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/turso-backup-job/turso-backup-image:latest"
  service_account_email = module.backup_service_account.email

  env_vars = merge(
    var.env_vars,
    {
      BACKUP_BUCKET = module.backup_storage_bucket.bucket_name
    }
  )

  cpu_limit    = "1"
  memory_limit = "512Mi"
  timeout      = "1800s" # 30 minutes for backup
  max_retries  = 3

  labels = {
    environment = var.environment
    service     = "db-backup"
    managed_by  = "opentofu"
  }
}

# Grant the backup service account permission to invoke the job
resource "google_cloud_run_v2_job_iam_member" "scheduler_invoker" {
  name     = module.db_backup_job.job_name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${module.backup_service_account.email}"
}

# Cloud Scheduler to trigger backups daily
module "backup_schedule" {
  source = "../../../../infrastructure/modules/cloud-scheduler"

  job_name    = "db-backup-daily"
  description = "Trigger database backup job daily"
  schedule    = var.backup_schedule
  time_zone   = var.backup_timezone

  http_target = {
    uri         = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${module.db_backup_job.job_name}:run"
    http_method = "POST"
    oidc_token = {
      service_account_email = module.backup_service_account.email
    }
  }

  retry_count          = 2
  max_retry_duration   = "0s"
  min_backoff_duration = "5s"
  max_backoff_duration = "3600s"
}
