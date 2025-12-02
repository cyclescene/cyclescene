resource "google_cloud_scheduler_job" "job" {
  name             = var.job_name
  description      = var.description
  schedule         = var.schedule
  time_zone        = var.time_zone
  region           = var.region
  attempt_deadline = var.attempt_deadline

  retry_config {
    retry_count          = var.retry_count
    max_retry_duration   = var.max_retry_duration
    min_backoff_duration = var.min_backoff_duration
    max_backoff_duration = var.max_backoff_duration
    max_doublings        = var.max_doublings
  }

  # HTTP Target (for Cloud Run services)
  dynamic "http_target" {
    for_each = var.http_target != null ? [var.http_target] : []
    content {
      uri         = http_target.value.uri
      http_method = http_target.value.http_method

      # Headers is a simple map, not a block
      headers = lookup(http_target.value, "headers", {})

      body = lookup(http_target.value, "body", null)

      dynamic "oidc_token" {
        for_each = lookup(http_target.value, "oidc_token", null) != null ? [http_target.value.oidc_token] : []
        content {
          service_account_email = oidc_token.value.service_account_email
          audience              = lookup(oidc_token.value, "audience", null)
        }
      }

      dynamic "oauth_token" {
        for_each = lookup(http_target.value, "oauth_token", null) != null ? [http_target.value.oauth_token] : []
        content {
          service_account_email = oauth_token.value.service_account_email
          scope                 = lookup(oauth_token.value, "scope", null)
        }
      }
    }
  }

  # Pub/Sub Target
  dynamic "pubsub_target" {
    for_each = var.pubsub_target != null ? [var.pubsub_target] : []
    content {
      topic_name = pubsub_target.value.topic_name
      data       = lookup(pubsub_target.value, "data", null)
      attributes = lookup(pubsub_target.value, "attributes", {})
    }
  }
}

# Create service account for Cloud Scheduler if needed
resource "google_service_account" "scheduler_sa" {
  count = var.create_service_account ? 1 : 0

  account_id   = "${var.job_name}-scheduler-sa"
  display_name = "Cloud Scheduler SA for ${var.job_name}"
  description  = "Service account used by Cloud Scheduler to invoke ${var.job_name}"
}

# Grant Cloud Run Invoker role to the service account
resource "google_cloud_run_service_iam_member" "invoker" {
  count = var.create_service_account && var.cloud_run_service_name != null ? 1 : 0

  service  = var.cloud_run_service_name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.scheduler_sa[0].email}"
}
