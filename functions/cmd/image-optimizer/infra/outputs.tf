output "optimizer_service_url" {
  description = "URL of the image optimizer Cloud Run service"
  value       = module.image_optimizer_service.service_url
}

output "optimizer_service_name" {
  description = "Name of the image optimizer Cloud Run service"
  value       = module.image_optimizer_service.service_name
}

output "optimized_bucket_name" {
  description = "Name of the optimized media bucket"
  value       = module.optimized_media_bucket.bucket_name
}

output "optimizer_service_account_email" {
  description = "Email of the optimizer service account"
  value       = module.optimizer_service_account.email
}
