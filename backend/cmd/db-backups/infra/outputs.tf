output "backup_bucket_name" {
  description = "Name of the backup storage bucket"
  value       = module.backup_storage_bucket.bucket_name
}

output "backup_bucket_url" {
  description = "URL of the backup storage bucket"
  value       = module.backup_storage_bucket.bucket_url
}

output "service_account_email" {
  description = "Email of the backup service account"
  value       = module.backup_service_account.email
}

output "job_name" {
  description = "Name of the backup job"
  value       = module.db_backup_job.job_name
}

output "job_id" {
  description = "ID of the backup job"
  value       = module.db_backup_job.job_id
}

output "schedule_name" {
  description = "Name of the backup scheduler"
  value       = module.backup_schedule.job_name
}
