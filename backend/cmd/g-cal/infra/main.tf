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
    prefix = "services/g-cal"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_project_service" "required_apis" {
  for_each = toset([
    "calendar-json.googleapis.com",
    "geocoding-backend.googleapis.com",
    "run.googleapis.com",
    "cloudscheduler.googleapis.com",
  ])

  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}

module "scheduler_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "gcal-scheduler"
  display_name = "Google Calendar Scheduler"
  description  = "Triggers the Google Calendar import job"
  project_id   = var.project_id

  roles = [
    "roles/run.invoker",
    "roles/serviceusage.serviceUsageConsumer",
  ]
}

module "calendar_job_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "gcal-job"
  display_name = "Google Calendar Import Job"
  description  = "Reads configured Google Calendars and geocodes their events"
  project_id   = var.project_id

  roles = [
    "roles/serviceusage.serviceUsageConsumer",
  ]
}

resource "google_service_account_iam_member" "wif_can_act_as_scheduler" {
  service_account_id = module.scheduler_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:github-actions@${var.project_id}.iam.gserviceaccount.com"
}

resource "google_service_account_iam_member" "wif_can_act_as_calendar_job" {
  service_account_id = module.calendar_job_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:github-actions@${var.project_id}.iam.gserviceaccount.com"
}

resource "google_service_account_iam_member" "scheduler_can_act_as_itself" {
  service_account_id = module.scheduler_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${module.scheduler_service_account.email}"
}

resource "google_project_iam_member" "scheduler_token_creator" {
  project = var.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${module.scheduler_service_account.email}"
}

module "calendar_job" {
  source = "../../../../infrastructure/modules/cloud-run-job"

  job_name              = "google-calendar-import"
  image                 = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/g-cal:${var.image_tag}"
  service_account_email = module.calendar_job_service_account.email
  env_vars = merge(var.env_vars, {
    GOOGLE_CALENDARS  = var.google_calendars
    TURSO_DB_URL      = var.turso_db_url
    TURSO_DB_RW_TOKEN = var.turso_db_rw_token
  })

  cpu_limit    = "1"
  memory_limit = "512Mi"
  timeout      = "600s"
  max_retries  = 3

  labels = {
    environment = var.environment
    service     = "g-cal"
    managed_by  = "opentofu"
  }

  depends_on = [google_project_service.required_apis]
}

resource "google_cloud_run_v2_job_iam_member" "scheduler_invoker" {
  name     = module.calendar_job.job_name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${module.scheduler_service_account.email}"
}

module "calendar_schedule" {
  source = "../../../../infrastructure/modules/cloud-scheduler"

  job_name    = "google-calendar-import-twice-daily"
  description = "Trigger the Google Calendar import job at 3 AM and 3 PM"
  schedule    = var.schedule
  time_zone   = var.timezone

  http_target = {
    uri         = "https://run.googleapis.com/v2/projects/${var.project_id}/locations/${var.region}/jobs/${module.calendar_job.job_name}:run"
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
