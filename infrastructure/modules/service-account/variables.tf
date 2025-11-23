variable "account_id" {
  description = "The account ID for the service account (must be 6-30 characters)"
  type        = string
}

variable "display_name" {
  description = "The display name for the service account"
  type        = string
  default     = null
}

variable "description" {
  description = "A description of the service account"
  type        = string
  default     = null
}

variable "project_id" {
  description = "The GCP project ID"
  type        = string
}

variable "roles" {
  description = "List of IAM roles to assign to the service account (e.g., ['roles/cloudsql.client', 'roles/storage.objectViewer'])"
  type        = list(string)
  default     = []
}
