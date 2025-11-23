# Create Pub/Sub topic
resource "google_pubsub_topic" "topic" {
  name = var.topic_name

  labels = var.labels

  # message_retention_duration is a simple field, not a block
  message_retention_duration = var.message_retention_duration

  dynamic "schema_settings" {
    for_each = var.schema_settings != null ? [var.schema_settings] : []
    content {
      schema   = schema_settings.value.schema
      encoding = schema_settings.value.encoding
    }
  }
}

# Create Eventarc trigger for Pub/Sub
resource "google_eventarc_trigger" "pubsub_trigger" {
  count = var.create_eventarc_trigger ? 1 : 0

  name     = var.trigger_name != null ? var.trigger_name : "${var.topic_name}-trigger"
  location = var.trigger_location
  project  = var.project_id

  # Match Pub/Sub messages from this topic
  matching_criteria {
    attribute = "type"
    value     = "google.cloud.pubsub.topic.v1.messagePublished"
  }

  # Destination Cloud Run service
  destination {
    cloud_run_service {
      service = var.cloud_run_service_name
      region  = var.trigger_location
      path    = var.cloud_run_path
    }
  }

  # Service account for invoking Cloud Run
  service_account = var.eventarc_service_account_email

  # Transport topic
  transport {
    pubsub {
      topic = google_pubsub_topic.topic.id
    }
  }

  labels = var.labels
}

# Create Pub/Sub subscription (legacy support)
resource "google_pubsub_subscription" "subscription" {
  count = var.create_subscription ? 1 : 0

  name  = var.subscription_name != null ? var.subscription_name : "${var.topic_name}-sub"
  topic = google_pubsub_topic.topic.name

  ack_deadline_seconds       = var.ack_deadline_seconds
  message_retention_duration = var.subscription_message_retention
  retain_acked_messages      = var.retain_acked_messages
  enable_message_ordering    = var.enable_message_ordering

  dynamic "expiration_policy" {
    for_each = var.subscription_expiration_ttl != null ? [1] : []
    content {
      ttl = var.subscription_expiration_ttl
    }
  }

  dynamic "retry_policy" {
    for_each = var.retry_policy != null ? [var.retry_policy] : []
    content {
      minimum_backoff = lookup(retry_policy.value, "minimum_backoff", "10s")
      maximum_backoff = lookup(retry_policy.value, "maximum_backoff", "600s")
    }
  }

  dynamic "dead_letter_policy" {
    for_each = var.dead_letter_topic != null ? [1] : []
    content {
      dead_letter_topic     = var.dead_letter_topic
      max_delivery_attempts = var.max_delivery_attempts
    }
  }

  # Push subscription for Cloud Run (legacy)
  dynamic "push_config" {
    for_each = var.push_endpoint != null ? [1] : []
    content {
      push_endpoint = var.push_endpoint

      attributes = var.push_attributes

      dynamic "oidc_token" {
        for_each = var.oidc_service_account_email != null ? [1] : []
        content {
          service_account_email = var.oidc_service_account_email
          audience              = var.oidc_audience
        }
      }
    }
  }

  labels = var.labels
}

# Grant Pub/Sub Publisher role to service accounts
resource "google_pubsub_topic_iam_member" "publishers" {
  for_each = toset(var.publisher_service_accounts)

  topic  = google_pubsub_topic.topic.name
  role   = "roles/pubsub.publisher"
  member = "serviceAccount:${each.value}"
}

# Grant Pub/Sub Subscriber role to service accounts
resource "google_pubsub_subscription_iam_member" "subscribers" {
  for_each = var.create_subscription ? toset(var.subscriber_service_accounts) : []

  subscription = google_pubsub_subscription.subscription[0].name
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:${each.value}"
}
