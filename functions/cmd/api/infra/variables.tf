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
