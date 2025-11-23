variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "location" {
  description = "GCP region/location"
  type        = string
  default     = "us-west1"
}

variable "channel_name" {
  description = "Name of the Eventarc channel"
  type        = string
}

variable "trigger_name" {
  description = "Name of the Eventarc trigger"
  type        = string
}

variable "trigger_description" {
  description = "Description of the Eventarc trigger"
  type        = string
  default     = ""
}

variable "event_type" {
  description = "Custom event type to listen for (e.g., com.cyclescene.image.optimization)"
  type        = string
}

variable "cloud_run_service_name" {
  description = "Name of the Cloud Run service to invoke"
  type        = string
}

variable "cloud_run_path" {
  description = "Path on the Cloud Run service to send events to"
  type        = string
  default     = "/"
}

variable "trigger_service_account" {
  description = "Service account email for the trigger to use when invoking Cloud Run"
  type        = string
}

variable "publisher_service_account_email" {
  description = "Email of the service account that will publish events to this channel"
  type        = string
  default     = ""
}

variable "labels" {
  description = "Labels to apply to resources"
  type        = map(string)
  default     = {}
}
