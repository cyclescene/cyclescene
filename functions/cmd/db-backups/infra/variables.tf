variable "project_id" {
  description = "GCP Project ID"
  type        = string
  default     = "cyclescene"
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-west1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "backup_schedule" {
  description = "Cron schedule for backups"
  type        = string
  default     = "0 2 * * *" # 2 AM daily
}

variable "backup_timezone" {
  description = "Timezone for backup schedule"
  type        = string
  default     = "UTC"
}

variable "env_vars" {
  description = "Environment variables for the backup job"
  type        = map(string)
  default = {
    NODE_ENV = "production"
  }
}
