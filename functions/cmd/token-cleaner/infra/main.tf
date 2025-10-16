terraform {
  required_version = ">= 1.6"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "gcs" {
    bucket = "cyclescene-terraform-state"
    prefix = "terraform/state/token-cleaner"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Service account for Cloud Scheduler to trigger the job
module "scheduler_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "token-cleaner-scheduler"
  display_name = "Token Cleaner Scheduler SA"
  description  = "Service account for Cloud Scheduler to trigger token cleaner job"
  project_id   = var.project_id

  roles = [
    "roles/run.invoker"  # Permission to invoke Cloud Run jobs
  ]
}

# Token Cleaner - Cloud Run Job that runs daily at midnight
module "token_cleaner_job" {
  source = "../../../../infrastructure/modules/cloud-run-job"

  job_name = "submission-token-cleaner"
  image    = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/submission-token-cleaner/token-cleaner-image:latest"

  env_vars = var.env_vars

  cpu_limit    = "1"
  memory_limit = "512Mi"
  timeout      = "600s"
  max_retries  = 3

  labels = {
    environment = var.environment
    service     = "token-cleaner"
    managed_by  = "opentofu"
  }
}

# Grant the scheduler service account permission to invoke the job
resource "google_cloud_run_v2_job_iam_member" "scheduler_invoker" {
  name     = module.token_cleaner_job.job_name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${module.scheduler_service_account.email}"
}

# Cloud Scheduler to trigger the job daily at midnight
module "token_cleaner_schedule" {
  source = "../../../../infrastructure/modules/cloud-scheduler"

  job_name    = "token-cleaner-daily"
  description = "Trigger token cleaner job daily at midnight"
  schedule    = "0 0 * * *" # Every day at midnight UTC
  time_zone   = "UTC"

  http_target = {
    uri         = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${module.token_cleaner_job.job_name}:run"
    http_method = "POST"
    oidc_token = {
      service_account_email = module.scheduler_service_account.email
    }
  }

  retry_count          = 2
  max_retry_duration   = "0s"
  min_backoff_duration = "5s"
  max_backoff_duration = "3600s"
}
