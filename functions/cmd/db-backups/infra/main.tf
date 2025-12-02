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

# Get the project number for API URLs
data "google_client_config" "current" {}

data "google_project" "project" {
  project_id = var.project_id
}

# Service Account for Cloud Scheduler to trigger the backup job
module "scheduler_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "backup-scheduler"
  display_name = "Backup Scheduler SA"
  description  = "Service account for Cloud Scheduler to trigger backup job"
  project_id   = var.project_id

  roles = [
    "roles/run.invoker",                           # Permission to invoke Cloud Run jobs
    "roles/serviceusage.serviceUsageConsumer"      # Permission to call Google APIs (IAM, etc)
  ]
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

# Allow GitHub Actions WIF service account to act as the scheduler service account
resource "google_service_account_iam_member" "wif_can_act_as_scheduler" {
  service_account_id = module.scheduler_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:github-actions@${var.project_id}.iam.gserviceaccount.com"
}

# Allow the scheduler service account to act as itself (needed for Terraform)
resource "google_service_account_iam_member" "scheduler_can_act_as_itself" {
  service_account_id = module.scheduler_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${module.scheduler_service_account.email}"
}

# Allow GitHub Actions WIF service account to act as the backup service account
resource "google_service_account_iam_member" "wif_can_act_as_backup" {
  service_account_id = module.backup_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:github-actions@${var.project_id}.iam.gserviceaccount.com"
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
  image                 = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/db-backups:${var.image_tag}"
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

# Grant the scheduler service account permission to invoke the job
resource "google_cloud_run_v2_job_iam_member" "scheduler_invoker" {
  name     = module.db_backup_job.job_name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${module.scheduler_service_account.email}"
}

# Grant the scheduler service account permission to create its own OIDC tokens
resource "google_project_iam_member" "scheduler_token_creator" {
  project = var.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${module.scheduler_service_account.email}"
}

# Allow the scheduler service account to be used by Cloud Scheduler
resource "google_service_account_iam_member" "scheduler_user" {
  service_account_id = module.scheduler_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${module.scheduler_service_account.email}"
}

# Cloud Scheduler to trigger backups daily
module "backup_schedule" {
  source = "../../../../infrastructure/modules/cloud-scheduler"

  job_name    = "db-backup-daily"
  description = "Trigger database backup job daily"
  schedule    = var.backup_schedule
  time_zone   = var.backup_timezone

  http_target = {
    uri         = "https://run.googleapis.com/v2/projects/${var.project_id}/locations/${var.region}/jobs/${module.db_backup_job.job_name}:run"
    http_method = "POST"
    headers = {
      "Content-Type" = "application/json"
    }
    oauth_token = {
      service_account_email = module.scheduler_service_account.email
      scope                 = "https://www.googleapis.com/auth/cloud-platform"
    }
  }

  retry_count          = 2
  max_retry_duration   = "0s"
  min_backoff_duration = "5s"
  max_backoff_duration = "3600s"
}
