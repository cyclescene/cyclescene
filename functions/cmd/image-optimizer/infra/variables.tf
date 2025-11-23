variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region for deploying resources"
  type        = string
  default     = "us-west1"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "staging_bucket_name" {
  description = "Name of the staging bucket where images are initially uploaded"
  type        = string
}

variable "api_service_account_email" {
  description = "Email of the API service account (for IAM permissions)"
  type        = string
}

variable "optimizer_cpu_limit" {
  description = "CPU limit for the optimizer service"
  type        = string
  default     = "2"
}

variable "optimizer_memory_limit" {
  description = "Memory limit for the optimizer service"
  type        = string
  default     = "2Gi"
}

variable "optimizer_max_instances" {
  description = "Maximum number of optimizer instances"
  type        = number
  default     = 10
}

variable "env_vars" {
  description = "Additional environment variables"
  type        = map(string)
  default     = {}
}

variable "turso_db_url" {
  description = "Turso database URL"
  type        = string
  sensitive   = true
}

variable "turso_db_rw_token" {
  description = "Turso database read/write token"
  type        = string
  sensitive   = true
}
