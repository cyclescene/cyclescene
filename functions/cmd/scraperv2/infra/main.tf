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
    prefix = "services/scraperv2"
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

# Allow GitHub Actions WIF service account to act as the scheduler service account
resource "google_service_account_iam_member" "wif_can_act_as_scheduler" {
  service_account_id = module.scheduler_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:github-actions@${var.project_id}.iam.gserviceaccount.com"
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

# Allow GitHub Actions WIF service account to act as the scraper job service account
resource "google_service_account_iam_member" "wif_can_act_as_scraper_job" {
  service_account_id = module.scraper_job_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:github-actions@${var.project_id}.iam.gserviceaccount.com"
}

# Cloud Run Job for PDX scraper
module "scraper_job" {
  source = "../../../../infrastructure/modules/cloud-run-job"

  job_name               = "pdx-scraper"
  image                  = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/scraperv2:${var.image_tag}"
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

  job_name    = "scraper-every-3h"
  description = "Trigger scraper job every 3 hours"
  schedule    = "0 */3 * * *" # Every 3 hours (00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00)
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
