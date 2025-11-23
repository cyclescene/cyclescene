output "job_name" {
  description = "Name of the Cloud Scheduler job"
  value       = google_cloud_scheduler_job.job.name
}

output "job_id" {
  description = "ID of the Cloud Scheduler job"
  value       = google_cloud_scheduler_job.job.id
}

output "service_account_email" {
  description = "Email of the created service account (if created)"
  value       = var.create_service_account ? google_service_account.scheduler_sa[0].email : null
}
