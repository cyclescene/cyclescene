# Eventarc Channel for custom application events
resource "google_eventarc_channel" "channel" {
  name       = var.channel_name
  location   = var.location
  project    = var.project_id

  labels = var.labels
}

# Eventarc Trigger that listens to channel and routes to Cloud Run
resource "google_eventarc_trigger" "trigger" {
  name        = var.trigger_name
  location    = var.location
  project     = var.project_id
  description = var.trigger_description

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

  # Service account for invoking Cloud Run
  service_account = var.trigger_service_account

  # Use the Eventarc channel
  transport {
    pubsub {
      topic = google_eventarc_channel.channel.pubsub_topic
    }
  }

  labels = var.labels
}
