variable "project_id" {
  description = "GCP project ID"
  type        = string
  default     = "cyclescene-479119"
}

variable "region" {
  description = "GCP region for the Cloud Run Job"
  type        = string
  default     = "us-west1"
}

variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "production"
}

variable "image_tag" {
  description = "Container image tag"
  type        = string
  default     = "latest"
}

variable "google_calendars" {
  description = "JSON array mapping Google Calendar IDs to Cycle Scene city codes"
  type        = string
}

variable "calendar_lookahead_days" {
  description = "Number of future days to import from Google Calendar"
  type        = number
  default     = 14

  validation {
    condition     = var.calendar_lookahead_days >= 1 && var.calendar_lookahead_days <= 365
    error_message = "calendar_lookahead_days must be between 1 and 365."
  }
}

variable "turso_db_url" {
  description = "Turso database URL for the shared rides database"
  type        = string
  sensitive   = true
}

variable "turso_db_rw_token" {
  description = "Turso read-write token for the shared rides database"
  type        = string
  sensitive   = true
}

variable "schedule" {
  description = "Cloud Scheduler cron expression"
  type        = string
  default     = "0 3,15 * * *"
}

variable "timezone" {
  description = "Timezone for the scheduler"
  type        = string
  default     = "America/Los_Angeles"
}

variable "env_vars" {
  description = "Additional environment variables for the import job"
  type        = map(string)
  default     = {}
}
