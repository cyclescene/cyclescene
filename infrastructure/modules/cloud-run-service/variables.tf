variable "service_name" {
  description = "Name of the Cloud Run service"
  type        = string
}

variable "region" {
  description = "GCP region where the service will be deployed"
  type        = string
  default     = "us-west1"
}

variable "image" {
  description = "Container image to deploy"
  type        = string
}

variable "service_account_email" {
  description = "Service account email for the Cloud Run service"
  type        = string
  default     = null
}

variable "env_vars" {
  description = "Map of environment variables"
  type        = map(string)
  default     = {}
}

variable "min_instances" {
  description = "Minimum number of instances"
  type        = number
  default     = 0
}

variable "max_instances" {
  description = "Maximum number of instances"
  type        = number
  default     = 10
}

variable "cpu_limit" {
  description = "CPU limit (e.g., '1', '2', '4')"
  type        = string
  default     = "1"
}

variable "memory_limit" {
  description = "Memory limit (e.g., '512Mi', '1Gi', '2Gi')"
  type        = string
  default     = "512Mi"
}

variable "cpu_always_allocated" {
  description = "Whether CPU should be always allocated (true) or only during request processing (false)"
  type        = bool
  default     = false
}

variable "container_port" {
  description = "Container port to expose"
  type        = number
  default     = 8080
}

variable "timeout" {
  description = "Request timeout in seconds"
  type        = string
  default     = "300s"
}

variable "vpc_connector" {
  description = "VPC connector for VPC access"
  type        = string
  default     = null
}

variable "vpc_egress" {
  description = "VPC egress setting (PRIVATE_RANGES_ONLY or ALL_TRAFFIC)"
  type        = string
  default     = "PRIVATE_RANGES_ONLY"
}

variable "allow_unauthenticated" {
  description = "Allow unauthenticated access to the service"
  type        = bool
  default     = false
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = false
}

variable "labels" {
  description = "Labels to apply to the service"
  type        = map(string)
  default     = {}
}
