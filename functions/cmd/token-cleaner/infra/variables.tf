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

variable "env_vars" {
  description = "Environment variables for the token cleaner"
  type        = map(string)
  default = {
    NODE_ENV = "production"
  }
}

variable "image_tag" {
  description = "Docker image tag for the token cleaner job"
  type        = string
  default     = "latest"
}
