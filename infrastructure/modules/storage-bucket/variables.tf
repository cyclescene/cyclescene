variable "bucket_name" {
  description = "Name of the storage bucket (must be globally unique)"
  type        = string
}

variable "location" {
  description = "Location of the bucket (e.g., 'US', 'EU', 'us-central1')"
  type        = string
  default     = "US"
}

variable "storage_class" {
  description = "Storage class (STANDARD, NEARLINE, COLDLINE, ARCHIVE)"
  type        = string
  default     = "STANDARD"
}

variable "uniform_bucket_level_access" {
  description = "Enable uniform bucket-level access"
  type        = bool
  default     = true
}

variable "force_destroy" {
  description = "Allow deletion of non-empty bucket"
  type        = bool
  default     = false
}

variable "versioning_enabled" {
  description = "Enable object versioning"
  type        = bool
  default     = false
}

variable "lifecycle_rules" {
  description = "List of lifecycle rules for the bucket"
  type = list(object({
    action = object({
      type          = string
      storage_class = optional(string)
    })
    condition = object({
      age                   = optional(number)
      created_before        = optional(string)
      with_state            = optional(string)
      matches_storage_class = optional(list(string))
      num_newer_versions    = optional(number)
    })
  }))
  default = []
}

variable "cors_rules" {
  description = "CORS rules for the bucket"
  type = list(object({
    origin          = list(string)
    method          = list(string)
    response_header = optional(list(string))
    max_age_seconds = optional(number)
  }))
  default = []
}

variable "iam_members" {
  description = "IAM members to grant access to the bucket. Key is a unique identifier, value contains role and member."
  type = map(object({
    role   = string
    member = string
  }))
  default = {}
}

variable "labels" {
  description = "Labels to apply to the bucket"
  type        = map(string)
  default     = {}
}
