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
