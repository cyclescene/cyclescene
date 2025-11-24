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

# Service Account for API with storage, geocoding, and signBlob permissions
module "api_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "cyclescene-api"
  display_name = "CycleScene API Service Account"
  description  = "Service account for API with storage, geocoding, and signed URL capabilities"
  project_id   = var.project_id

  roles = [
    "roles/iam.serviceAccountTokenCreator",              # Required for signing URLs
    "roles/serviceusage.serviceUsageConsumer",           # Required to call Google APIs
    "roles/eventarc.publisher",                          # Required to publish events to Eventarc
  ]
}

# Allow GitHub Actions WIF service account to act as the API service account
resource "google_service_account_iam_member" "wif_can_act_as_api" {
  service_account_id = module.api_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:github-actions@${var.project_id}.iam.gserviceaccount.com"
}

# Storage bucket for user-submitted media (images)
module "user_media_bucket" {
  source = "../../../../infrastructure/modules/storage-bucket"

  bucket_name                 = "${var.project_id}-user-media-staging"
  location                    = var.region
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
  force_destroy               = false
  versioning_enabled          = false

  # CORS configuration for browser uploads
  cors_rules = [
    {
      origin          = var.allowed_origins
      method          = ["GET", "POST", "PUT", "HEAD"]
      response_header = ["Content-Type", "Content-Length"]
      max_age_seconds = 3600
    }
  ]

  # Lifecycle rule to transition old media to cheaper storage
  lifecycle_rules = [
    {
      action = {
        type          = "SetStorageClass"
        storage_class = "NEARLINE"
      }
      condition = {
        age = 90 # Move to nearline after 90 days
      }
    }
  ]

  labels = {
    environment = var.environment
    purpose     = "user-media"
    managed_by  = "opentofu"
  }

  # Grant the API service account permission to create signed URLs and upload objects
  iam_members = {
    "api-storage-access" = {
      role   = "roles/storage.objectUser"
      member = "serviceAccount:${module.api_service_account.email}"
    }
  }
}

# Cloud Run Service for API Gateway
module "api_service" {
  source = "../../../../infrastructure/modules/cloud-run-service"

  service_name          = "cyclescene-api-gateway"
  image                 = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/api:${var.image_tag}"
  service_account_email = module.api_service_account.email

  env_vars = merge(
    var.env_vars,
    {
      STAGING_BUCKET_NAME   = module.user_media_bucket.bucket_name
      GCP_PROJECT           = var.project_id
      SERVICE_ACCOUNT_EMAIL = module.api_service_account.email
      TURSO_DB_URL          = var.turso_db_url
      TURSO_DB_RW_TOKEN     = var.turso_db_rw_token
      IMAGE_OPTIMIZER_URL   = var.image_optimizer_url
      EVENTARC_CHANNEL_NAME = var.eventarc_channel_name
      RESEND_API_KEY        = var.resend_api_key
      EDIT_LINK_BASE_URL    = var.edit_link_base_url
    }
  )

  # Resource configuration
  cpu_limit              = var.api_cpu_limit
  memory_limit           = var.api_memory_limit
  cpu_always_allocated   = false

  # Scaling configuration
  min_instances = var.api_min_instances
  max_instances = var.api_max_instances

  # Network configuration
  container_port        = 8080
  allow_unauthenticated = var.api_allow_public
  timeout               = "300s"

  labels = {
    environment = var.environment
    service     = "api"
    managed_by  = "opentofu"
  }
}

# Custom domain mapping for API (optional)
module "api_domain" {
  count  = var.api_custom_domain != "" ? 1 : 0
  source = "../../../../infrastructure/modules/cloud-run-domain"

  project_id              = var.project_id
  location                = var.region
  domain_name             = var.api_custom_domain
  cloud_run_service_name  = module.api_service.service_name
}
