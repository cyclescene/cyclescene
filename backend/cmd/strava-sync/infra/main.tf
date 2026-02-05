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

# Get project data
data "google_project" "project" {
  project_id = var.project_id
}

# Service Account for Cloud Scheduler (triggers job)
module "scheduler_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "strava-sync-scheduler"
  display_name = "Strava Sync Scheduler SA"
  description  = "Service account for Cloud Scheduler to trigger strava-sync job"
  project_id   = var.project_id

  roles = [
    "roles/run.invoker",                           # Permission to invoke Cloud Run jobs
    "roles/serviceusage.serviceUsageConsumer"      # Permission to call Google APIs (IAM, etc)
  ]
}

# Service Account for Strava Sync Job (runs sync)
module "sync_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "strava-sync-job"
  display_name = "Strava Sync Job SA"
  description  = "Service account for strava-sync Cloud Run Job"
  project_id   = var.project_id

  roles = [] # No GCP roles needed, only Turso/Strava access
}

# Allow GitHub Actions WIF to act as scheduler SA
resource "google_service_account_iam_member" "wif_can_act_as_scheduler" {
  service_account_id = module.scheduler_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:github-actions@${var.project_id}.iam.gserviceaccount.com"
}

# Allow scheduler SA to act as itself
resource "google_service_account_iam_member" "scheduler_can_act_as_itself" {
  service_account_id = module.scheduler_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${module.scheduler_service_account.email}"
}

# Allow GitHub Actions WIF to act as sync SA
resource "google_service_account_iam_member" "wif_can_act_as_sync" {
  service_account_id = module.sync_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:github-actions@${var.project_id}.iam.gserviceaccount.com"
}

# Cloud Run Job for Strava sync
module "strava_sync_job" {
  source = "../../../../infrastructure/modules/cloud-run-job"

  job_name              = "strava-sync"
  image                 = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/strava-sync:${var.image_tag}"
  service_account_email = module.sync_service_account.email

  env_vars = merge(
    var.env_vars,
    {
      TURSO_URL                     = var.turso_url
      TURSO_AUTH_TOKEN              = var.turso_auth_token
      TURSO_MONITORING_URL          = var.turso_monitoring_url
      TURSO_MONITORING_AUTH_TOKEN   = var.turso_monitoring_auth_token
      STRAVA_CLIENT_ID              = var.strava_client_id
      STRAVA_CLIENT_SECRET          = var.strava_client_secret
      STRAVA_TOKEN_ENCRYPTION_KEY   = var.strava_token_encryption_key
      SYNC_MAX_CONNECTIONS          = tostring(var.max_connections)
      SYNC_MAX_REQUESTS_15MIN       = tostring(var.max_requests_15min)
      SYNC_MAX_REQUESTS_DAY         = tostring(var.max_requests_day)
      # Alerting configuration
      ADMIN_EMAIL                   = var.admin_email
      NTFY_TOPIC                    = var.ntfy_topic
      RESEND_API_KEY                = var.resend_api_key
    }
  )

  cpu_limit    = "1"
  memory_limit = "512Mi"
  timeout      = "1800s" # 30 minutes
  max_retries  = 2

  labels = {
    environment = var.environment
    service     = "strava-sync"
    managed_by  = "terraform"
  }
}

# Grant scheduler SA permission to invoke job
resource "google_cloud_run_v2_job_iam_member" "scheduler_invoker" {
  name     = module.strava_sync_job.job_name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${module.scheduler_service_account.email}"
}

# Allow scheduler SA to create tokens
resource "google_project_iam_member" "scheduler_token_creator" {
  project = var.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${module.scheduler_service_account.email}"
}

# Allow scheduler SA to be used
resource "google_service_account_iam_member" "scheduler_user" {
  service_account_id = module.scheduler_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${module.scheduler_service_account.email}"
}

# Cloud Scheduler to trigger sync every 3 days at 2am Pacific
module "sync_schedule" {
  source = "../../../../infrastructure/modules/cloud-scheduler"

  job_name    = "strava-sync-trigger"
  description = "Trigger Strava sync job every 3 days at 2am Pacific"
  schedule    = var.sync_schedule
  time_zone   = var.sync_timezone

  http_target = {
    uri         = "https://run.googleapis.com/v2/projects/${var.project_id}/locations/${var.region}/jobs/${module.strava_sync_job.job_name}:run"
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
