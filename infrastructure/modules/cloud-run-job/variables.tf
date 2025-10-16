variable "job_name" {
  description = "Name of the Cloud Run job"
  type        = string
}

variable "region" {
  description = "GCP region where the job will be deployed"
  type        = string
  default     = "us-west1"
}

variable "image" {
  description = "Container image to deploy"
  type        = string
}

variable "service_account_email" {
  description = "Service account email for the Cloud Run job"
  type        = string
  default     = null
}

variable "env_vars" {
  description = "Map of environment variables"
  type        = map(string)
  default     = {}
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

variable "timeout" {
  description = "Maximum execution time for the job in seconds"
  type        = string
  default     = "600s"
}

variable "max_retries" {
  description = "Maximum number of retries for failed tasks"
  type        = number
  default     = 3
}

variable "task_count" {
  description = "Number of tasks to run"
  type        = number
  default     = 1
}

variable "parallelism" {
  description = "Number of tasks to run in parallel"
  type        = number
  default     = 1
}

variable "command" {
  description = "Override the container's ENTRYPOINT"
  type        = list(string)
  default     = []
}

variable "args" {
  description = "Override the container's CMD"
  type        = list(string)
  default     = []
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

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = false
}

variable "labels" {
  description = "Labels to apply to the job"
  type        = map(string)
  default     = {}
}
