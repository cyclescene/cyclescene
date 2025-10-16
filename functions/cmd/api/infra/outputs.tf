output "api_url" {
  description = "URL of the API service"
  value       = module.api_service.service_url
}

output "service_name" {
  description = "Name of the API service"
  value       = module.api_service.service_name
}

output "service_account_email" {
  description = "Email of the API service account"
  value       = module.api_service_account.email
}

output "media_bucket_name" {
  description = "Name of the user media bucket"
  value       = module.user_media_bucket.bucket_name
}

output "media_bucket_url" {
  description = "URL of the user media bucket"
  value       = module.user_media_bucket.bucket_url
}
