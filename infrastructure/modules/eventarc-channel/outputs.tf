output "channel_name" {
  description = "Name of the Eventarc channel"
  value       = google_eventarc_channel.channel.name
}

output "channel_id" {
  description = "ID of the Eventarc channel"
  value       = google_eventarc_channel.channel.id
}

output "pubsub_topic" {
  description = "Pub/Sub topic associated with the Eventarc channel"
  value       = google_eventarc_channel.channel.pubsub_topic
}

output "trigger_name" {
  description = "Name of the Eventarc trigger"
  value       = google_eventarc_trigger.trigger.name
}

output "trigger_id" {
  description = "ID of the Eventarc trigger"
  value       = google_eventarc_trigger.trigger.id
}
