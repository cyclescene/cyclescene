# Eventarc Channel for custom application events
resource "google_eventarc_channel" "channel" {
  name       = var.channel_name
  location   = var.location
  project    = var.project_id
}

# Eventarc Trigger that listens to channel and routes to Cloud Run
resource "google_eventarc_trigger" "trigger" {
  name            = var.trigger_name
  location        = var.location
  project         = var.project_id
  channel         = google_eventarc_channel.channel.id
  service_account = var.trigger_service_account

  # Listen to custom events from the channel
  matching_criteria {
    attribute = "type"
    value     = var.event_type
  }

  # Route to Cloud Run service
  destination {
    cloud_run_service {
      service = var.cloud_run_service_name
      region  = var.location
      path    = var.cloud_run_path
    }
  }

  labels = var.labels
}

# Grant publisher service account permission to publish events to the channel
# The channel uses a Pub/Sub topic under the hood, so we grant pubsub.publisher on that topic
resource "google_pubsub_topic_iam_member" "channel_publisher" {
  count   = var.publisher_service_account_email != "" ? 1 : 0
  topic   = google_eventarc_channel.channel.pubsub_topic
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${var.publisher_service_account_email}"
}
