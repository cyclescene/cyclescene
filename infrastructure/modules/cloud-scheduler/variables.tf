variable "job_name" {
  description = "Name of the Cloud Scheduler job"
  type        = string
}

variable "description" {
  description = "Description of the job"
  type        = string
  default     = null
}

variable "schedule" {
  description = "Cron schedule (e.g., '0 */6 * * *' for every 6 hours)"
  type        = string
}

variable "time_zone" {
  description = "Time zone for the schedule (e.g., 'America/New_York')"
  type        = string
  default     = "UTC"
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-west1"
}

variable "attempt_deadline" {
  description = "Deadline for job attempts (e.g., '320s')"
  type        = string
  default     = "320s"
}

variable "retry_count" {
  description = "Number of retry attempts"
  type        = number
  default     = 3
}

variable "max_retry_duration" {
  description = "Maximum retry duration (e.g., '0s' for unlimited)"
  type        = string
  default     = "0s"
}

variable "min_backoff_duration" {
  description = "Minimum backoff duration between retries"
  type        = string
  default     = "5s"
}

variable "max_backoff_duration" {
  description = "Maximum backoff duration between retries"
  type        = string
  default     = "3600s"
}

variable "max_doublings" {
  description = "Maximum number of times the interval between retries will be doubled"
  type        = number
  default     = 5
}

variable "http_target" {
  description = "HTTP target configuration for Cloud Run or HTTP endpoints"
  type = object({
    uri         = string
    http_method = optional(string, "POST")
    headers     = optional(map(string), {})
    body        = optional(string)
    oidc_token = optional(object({
      service_account_email = string
      audience              = optional(string)
    }))
    oauth_token = optional(object({
      service_account_email = string
      scope                 = optional(string)
    }))
  })
  default = null
}

variable "pubsub_target" {
  description = "Pub/Sub target configuration"
  type = object({
    topic_name = string
    data       = optional(string)
    attributes = optional(map(string), {})
  })
  default = null
}

variable "create_service_account" {
  description = "Create a service account for the scheduler job"
  type        = bool
  default     = false
}

variable "cloud_run_service_name" {
  description = "Cloud Run service name to grant invoker permissions (required if create_service_account is true)"
  type        = string
  default     = null
}
