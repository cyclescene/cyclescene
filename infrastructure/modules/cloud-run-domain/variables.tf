variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "location" {
  description = "GCP region where the Cloud Run service is deployed"
  type        = string
  default     = "us-west1"
}

variable "domain_name" {
  description = "Custom domain name to map to the Cloud Run service (e.g., api.example.com)"
  type        = string
}

variable "cloud_run_service_name" {
  description = "Name of the Cloud Run service to map the domain to"
  type        = string
}
