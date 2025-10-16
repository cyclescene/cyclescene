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
