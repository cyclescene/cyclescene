variable "project_id" {
  description = "GCP project ID"
  type        = string
  default     = null
}

variable "topic_name" {
  description = "Name of the Pub/Sub topic"
  type        = string
}

variable "message_retention_duration" {
  description = "Message retention duration in seconds (86400s = 1 day, max 2678400s = 31 days)"
  type        = string
  default     = null
}

variable "schema_settings" {
  description = "Schema settings for the topic"
  type = object({
    schema   = string
    encoding = string
  })
  default = null
}

variable "labels" {
  description = "Labels to apply to resources"
  type        = map(string)
  default     = {}
}

# Eventarc trigger settings
variable "create_eventarc_trigger" {
  description = "Whether to create an Eventarc trigger"
  type        = bool
  default     = false
}

variable "trigger_name" {
  description = "Name of the Eventarc trigger (defaults to topic_name-trigger)"
  type        = string
  default     = null
}

variable "trigger_location" {
  description = "Location/region for the Eventarc trigger"
  type        = string
  default     = "us-central1"
}

variable "cloud_run_service_name" {
  description = "Name of the Cloud Run service to trigger"
  type        = string
  default     = null
}

variable "cloud_run_path" {
  description = "Path on the Cloud Run service to send events to"
  type        = string
  default     = "/"
}

variable "eventarc_service_account_email" {
  description = "Service account email for Eventarc to use when invoking Cloud Run"
  type        = string
  default     = null
}

# Subscription settings
variable "create_subscription" {
  description = "Whether to create a subscription"
  type        = bool
  default     = true
}

variable "subscription_name" {
  description = "Name of the subscription (defaults to topic_name-sub)"
  type        = string
  default     = null
}

variable "ack_deadline_seconds" {
  description = "Ack deadline in seconds"
  type        = number
  default     = 10
}

variable "subscription_message_retention" {
  description = "How long to retain unacknowledged messages (default 7 days)"
  type        = string
  default     = "604800s"
}

variable "retain_acked_messages" {
  description = "Whether to retain acknowledged messages"
  type        = bool
  default     = false
}

variable "enable_message_ordering" {
  description = "Enable message ordering"
  type        = bool
  default     = false
}

variable "subscription_expiration_ttl" {
  description = "TTL for the subscription (null = never expires)"
  type        = string
  default     = null
}

variable "retry_policy" {
  description = "Retry policy configuration"
  type = object({
    minimum_backoff = optional(string, "10s")
    maximum_backoff = optional(string, "600s")
  })
  default = null
}

variable "dead_letter_topic" {
  description = "Dead letter topic name"
  type        = string
  default     = null
}

variable "max_delivery_attempts" {
  description = "Maximum delivery attempts before sending to dead letter topic"
  type        = number
  default     = 5
}

# Push subscription settings (for Cloud Run)
variable "push_endpoint" {
  description = "Push endpoint URL (e.g., Cloud Run service URL)"
  type        = string
  default     = null
}

variable "push_attributes" {
  description = "Additional attributes for push subscriptions"
  type        = map(string)
  default     = {}
}

variable "oidc_service_account_email" {
  description = "Service account email for OIDC token authentication"
  type        = string
  default     = null
}

variable "oidc_audience" {
  description = "Audience for OIDC token (defaults to push_endpoint)"
  type        = string
  default     = null
}

# IAM settings
variable "publisher_service_accounts" {
  description = "List of service account emails that can publish to this topic"
  type        = list(string)
  default     = []
}

variable "subscriber_service_accounts" {
  description = "List of service account emails that can subscribe to this topic"
  type        = list(string)
  default     = []
}
