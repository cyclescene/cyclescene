terraform {
  required_version = ">= 1.6"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "gcs" {
    bucket = "cyclescene-479119-terraform-state"
    prefix = "terraform/state/scraper"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Service account for Cloud Scheduler to trigger the job
module "scheduler_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "scraper-scheduler"
  display_name = "Scraper Scheduler SA"
  description  = "Service account for Cloud Scheduler to trigger scraper job"
  project_id   = var.project_id

  roles = [
    "roles/run.invoker"  # Permission to invoke Cloud Run jobs
  ]
}

# Service account for the scraper job itself
module "scraper_job_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "scraper-job"
  display_name = "Scraper Job SA"
  description  = "Service account for scraper job to access Google APIs and databases"
  project_id   = var.project_id

  roles = [
    "roles/serviceusage.serviceUsageConsumer"  # Required to call Google APIs (geocoding)
  ]
}

# Cloud Run Job for PDX scraper
module "scraper_job" {
  source = "../../../../infrastructure/modules/cloud-run-job"

  job_name               = "pdx-scraper"
  image                  = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/pdx-scraper/scraper-image:latest"
  service_account_email  = module.scraper_job_service_account.email

  env_vars = var.env_vars

  cpu_limit    = "1"
  memory_limit = "512Mi"
  timeout      = "600s"
  max_retries  = 3

  labels = {
    environment = var.environment
    service     = "scraper"
    managed_by  = "opentofu"
  }
}

# Grant the scheduler service account permission to invoke the job
resource "google_cloud_run_v2_job_iam_member" "scheduler_invoker" {
  name     = module.scraper_job.job_name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${module.scheduler_service_account.email}"
}

# Cloud Scheduler to trigger scraper every 6 hours
module "scraper_schedule" {
  source = "../../../../infrastructure/modules/cloud-scheduler"

  job_name    = "scraper-every-6h"
  description = "Trigger scraper job every 6 hours"
  schedule    = "0 */6 * * *" # Every 6 hours (00:00, 06:00, 12:00, 18:00)
  time_zone   = var.scraper_timezone

  http_target = {
    uri         = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${module.scraper_job.job_name}:run"
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
