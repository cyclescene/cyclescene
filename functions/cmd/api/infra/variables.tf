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

variable "api_cpu_limit" {
  description = "CPU limit for API service"
  type        = string
  default     = "2"
}

variable "api_memory_limit" {
  description = "Memory limit for API service"
  type        = string
  default     = "1Gi"
}

variable "api_min_instances" {
  description = "Minimum number of API instances"
  type        = number
  default     = 0
}

variable "api_max_instances" {
  description = "Maximum number of API instances"
  type        = number
  default     = 10
}

variable "api_allow_public" {
  description = "Allow unauthenticated access to API"
  type        = bool
  default     = true
}

variable "allowed_origins" {
  description = "List of allowed origins for CORS"
  type        = list(string)
  default     = ["*"]
}

variable "env_vars" {
  description = "Environment variables for the API"
  type        = map(string)
  default = {
    NODE_ENV = "production"
  }
}

variable "api_custom_domain" {
  description = "Custom domain name for the API (e.g., api.cyclescene.cc)"
  type        = string
  default     = ""
}

variable "optimizer_service_account_email" {
  description = "Email of the image optimizer service account for IAM permissions"
  type        = string
  default     = ""
}

variable "staging_bucket_name" {
  description = "Name of the staging bucket for image uploads"
  type        = string
  default     = ""
}

variable "turso_db_url" {
  description = "Turso database URL"
  type        = string
  default     = ""
  sensitive   = true
}

variable "turso_db_rw_token" {
  description = "Turso database read-write token"
  type        = string
  default     = ""
  sensitive   = true
}

variable "optimizer_cpu_limit" {
  description = "CPU limit for image optimizer service"
  type        = string
  default     = "2"
}

variable "optimizer_memory_limit" {
  description = "Memory limit for image optimizer service"
  type        = string
  default     = "2Gi"
}

variable "optimizer_max_instances" {
  description = "Maximum number of optimizer instances"
  type        = number
  default     = 5
}

variable "image_optimizer_url" {
  description = "URL for the image optimizer service"
  type        = string
  default     = ""
  sensitive   = false
}

variable "eventarc_channel_name" {
  description = "Name of the Eventarc channel for image optimization events"
  type        = string
  default     = "image-optimization-events"
}
