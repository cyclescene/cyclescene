output "job_name" {
  description = "Cloud Run Job name"
  value       = module.calendar_job.job_name
}

output "calendar_job_service_account_email" {
  description = "Share each Google Calendar with this service account as a Reader"
  value       = module.calendar_job_service_account.email
}

output "schedule_name" {
  description = "Cloud Scheduler job name"
  value       = module.calendar_schedule.job_name
}
