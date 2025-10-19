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
    prefix = "terraform/state/image-optimizer"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Service Account for Image Optimizer with storage and DB access
module "optimizer_service_account" {
  source = "../../../../infrastructure/modules/service-account"

  account_id   = "cyclescene-image-optimizer"
  display_name = "CycleScene Image Optimizer Service Account"
  description  = "Service account for image optimization with storage and database access"
  project_id   = var.project_id

  roles = [
    "roles/storage.objectAdmin",       # Full access to storage buckets (read/write/delete)
  ]
}

# Optimized media bucket for final processed images
module "optimized_media_bucket" {
  source = "../../../../infrastructure/modules/storage-bucket"

  bucket_name                 = "${var.project_id}-user-media-optimized"
  location                    = var.region
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
  force_destroy               = false
  versioning_enabled          = false

  labels = {
    environment = var.environment
    purpose     = "optimized-media"
    managed_by  = "opentofu"
  }
}

# Cloud Run Service for Image Optimizer
module "image_optimizer_service" {
  source = "../../../../infrastructure/modules/cloud-run-service"

  service_name          = "cyclescene-image-optimizer"
  image                 = "${var.region}-docker.pkg.dev/${var.project_id}/cyclescene/image-optimizer/image-optimizer-image:latest"
  service_account_email = module.optimizer_service_account.email

  env_vars = merge(
    var.env_vars,
    {
      GCP_PROJECT       = var.project_id
      STAGING_BUCKET    = var.staging_bucket_name
      OPTIMIZED_BUCKET  = module.optimized_media_bucket.bucket_name
      TURSO_DB_URL      = var.turso_db_url
      TURSO_DB_RW_TOKEN = var.turso_db_rw_token
    }
  )

  # Resource configuration
  cpu_limit              = var.optimizer_cpu_limit
  memory_limit           = var.optimizer_memory_limit
  cpu_always_allocated   = false

  # Scaling configuration
  min_instances = 0  # Scale to zero when not in use
  max_instances = var.optimizer_max_instances

  # Network configuration
  container_port        = 8080
  allow_unauthenticated = false  # Restrict to internal calls only
  timeout               = "600s"  # 10 minutes for large images

  labels = {
    environment = var.environment
    service     = "image-optimizer"
    managed_by  = "opentofu"
  }
}

# Eventarc channel for image optimization events
module "image_optimization_channel" {
  source = "../../../../infrastructure/modules/eventarc-channel"

  project_id                = var.project_id
  location                  = var.region
  channel_name              = "image-optimization-events"
  trigger_name              = "image-optimizer-trigger"
  trigger_description       = "Routes image optimization events to the image optimizer service"
  event_type                = "com.cyclescene.image.optimization"
  cloud_run_service_name    = module.image_optimizer_service.service_name
  cloud_run_path            = "/optimize"
  trigger_service_account   = google_service_account.eventarc_trigger_sa.email

  labels = {
    environment = var.environment
    service     = "image-optimizer"
    managed_by  = "opentofu"
  }
}

# Service account for Eventarc trigger to invoke Cloud Run
resource "google_service_account" "eventarc_trigger_sa" {
  account_id   = "eventarc-image-optimizer-trigger"
  display_name = "Eventarc Image Optimizer Trigger SA"
  description  = "Service account for Eventarc trigger to invoke image optimizer"
  project      = var.project_id
}

# Grant Eventarc trigger SA permission to invoke the optimizer service
resource "google_cloud_run_service_iam_member" "eventarc_invoker" {
  service  = module.image_optimizer_service.service_name
  location = module.image_optimizer_service.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.eventarc_trigger_sa.email}"
}

# Grant API service account permission to invoke the optimizer service (legacy, for direct calls)
resource "google_cloud_run_service_iam_member" "api_invoker" {
  service  = module.image_optimizer_service.service_name
  location = module.image_optimizer_service.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${var.api_service_account_email}"
}
