output "topic_name" {
  description = "Name of the Pub/Sub topic"
  value       = google_pubsub_topic.topic.name
}

output "topic_id" {
  description = "ID of the Pub/Sub topic"
  value       = google_pubsub_topic.topic.id
}

output "subscription_name" {
  description = "Name of the subscription (if created)"
  value       = var.create_subscription ? google_pubsub_subscription.subscription[0].name : null
}

output "subscription_id" {
  description = "ID of the subscription (if created)"
  value       = var.create_subscription ? google_pubsub_subscription.subscription[0].id : null
}

output "eventarc_trigger_name" {
  description = "Name of the Eventarc trigger (if created)"
  value       = var.create_eventarc_trigger ? google_eventarc_trigger.pubsub_trigger[0].name : null
}

output "eventarc_trigger_id" {
  description = "ID of the Eventarc trigger (if created)"
  value       = var.create_eventarc_trigger ? google_eventarc_trigger.pubsub_trigger[0].id : null
}

output "eventarc_trigger_uid" {
  description = "UID of the Eventarc trigger (if created)"
  value       = var.create_eventarc_trigger ? google_eventarc_trigger.pubsub_trigger[0].uid : null
}
