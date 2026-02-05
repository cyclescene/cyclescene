variable "project_id" {
  description = "GCP project ID"
  type        = string
  default     = "leaguefindr"
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-west1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "image_tag" {
  description = "Docker image tag to deploy"
  type        = string
}

variable "turso_url" {
  description = "Turso database URL"
  type        = string
  sensitive   = true
}

variable "turso_auth_token" {
  description = "Turso database auth token"
  type        = string
  sensitive   = true
}

variable "turso_monitoring_url" {
  description = "Turso monitoring database URL"
  type        = string
  sensitive   = true
}

variable "turso_monitoring_auth_token" {
  description = "Turso monitoring database auth token"
  type        = string
  sensitive   = true
}

variable "strava_client_id" {
  description = "Strava API client ID"
  type        = string
}

variable "strava_client_secret" {
  description = "Strava API client secret"
  type        = string
  sensitive   = true
}

variable "strava_token_encryption_key" {
  description = "Encryption key for Strava refresh tokens"
  type        = string
  sensitive   = true
}

variable "sync_schedule" {
  description = "Cloud Scheduler cron schedule"
  type        = string
  default     = "0 2 */3 * *" # Every 3 days at 2am
}

variable "sync_timezone" {
  description = "Time zone for sync schedule"
  type        = string
  default     = "America/Los_Angeles"
}

variable "max_connections" {
  description = "Maximum connections to sync per run"
  type        = number
  default     = 100
}

variable "max_requests_15min" {
  description = "Maximum API requests per 15 minutes"
  type        = number
  default     = 90
}

variable "max_requests_day" {
  description = "Maximum API requests per day"
  type        = number
  default     = 900
}

variable "admin_email" {
  description = "Admin email for alert notifications"
  type        = string
  default     = ""
}

variable "ntfy_topic" {
  description = "ntfy.sh topic for push notifications"
  type        = string
  default     = "cyclescene-strava-sync"
}

variable "resend_api_key" {
  description = "Resend API key for email alerts"
  type        = string
  sensitive   = true
  default     = ""
}

variable "env_vars" {
  description = "Additional environment variables for the sync job"
  type        = map(string)
  default     = {}
}
