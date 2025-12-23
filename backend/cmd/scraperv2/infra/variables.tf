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

variable "scraper_timezone" {
  description = "Timezone for scraper schedule"
  type        = string
  default     = "UTC"
}

variable "turso_db_url" {
  description = "Turso database URL"
  type        = string
  sensitive   = true
}

variable "turso_db_rw_token" {
  description = "Turso database read-write token"
  type        = string
  sensitive   = true
}

variable "turso_db_ro_token" {
  description = "Turso database read-only token"
  type        = string
  sensitive   = true
}

variable "strava_access_token" {
  description = "Strava API access token"
  type        = string
  sensitive   = true
  default     = ""
}

variable "rwgps_auth_token" {
  description = "RideWithGPS authentication token"
  type        = string
  sensitive   = true
  default     = ""
}

variable "rwgps_api_key" {
  description = "RideWithGPS API key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "env_vars" {
  description = "Environment variables for the scraper"
  type        = map(string)
  default = {
    NODE_ENV = "production"
  }
}

variable "image_tag" {
  description = "Docker image tag for the scraper service"
  type        = string
  default     = "latest"
}
