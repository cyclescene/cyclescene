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

# Allow GitHub Actions WIF service account to act as the scheduler service account
resource "google_service_account_iam_member" "wif_can_act_as_scheduler" {
  service_account_id = module.scheduler_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:github-actions@${var.project_id}.iam.gserviceaccount.com"
}

# Service account for the token cleaner job itself
module "token_cleaner_sa" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "token-cleaner-job"
  display_name = "Token Cleaner Job SA"
  description  = "Service account for token cleaner job to access databases"
  project_id   = var.project_id

  roles = [
    "roles/serviceusage.serviceUsageConsumer"  # Required to call Google APIs
  ]
}

# Allow GitHub Actions WIF service account to act as the token cleaner job service account
resource "google_service_account_iam_member" "wif_can_act_as_token_cleaner_job" {
  service_account_id = module.token_cleaner_sa.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:github-actions@${var.project_id}.iam.gserviceaccount.com"
}

# Token Cleaner - Cloud Run Job that runs daily at midnight
module "token_cleaner_job" {
  source = "../../../../infrastructure/modules/cloud-run-job"

  job_name              = "submission-token-cleaner"
  image                 = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/token-cleaner:${var.image_tag}"
  service_account_email = module.token_cleaner_sa.email

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

# Grant the scheduler service account permission to create its own OIDC tokens
resource "google_project_iam_member" "scheduler_token_creator" {
  project = var.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${module.scheduler_service_account.email}"
}

# Allow the service account to be used by Cloud Scheduler
resource "google_service_account_iam_member" "scheduler_user" {
  service_account_id = module.scheduler_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${module.scheduler_service_account.email}"
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
    headers = {
      "Content-Type" = "application/json"
    }
    oidc_token = {
      service_account_email = module.scheduler_service_account.email
      audience              = "https://${var.region}-run.googleapis.com"
    }
  }

  retry_count          = 2
  max_retry_duration   = "0s"
  min_backoff_duration = "5s"
  max_backoff_duration = "3600s"
}
