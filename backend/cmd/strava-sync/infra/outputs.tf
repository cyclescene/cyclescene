output "job_name" {
  description = "Cloud Run Job name"
  value       = module.strava_sync_job.job_name
}

output "job_id" {
  description = "Cloud Run Job ID"
  value       = module.strava_sync_job.job_id
}

output "scheduler_job_name" {
  description = "Cloud Scheduler job name"
  value       = module.sync_schedule.job_name
}

output "scheduler_sa_email" {
  description = "Scheduler service account email"
  value       = module.scheduler_service_account.email
}

output "sync_sa_email" {
  description = "Sync job service account email"
  value       = module.sync_service_account.email
}

output "schedule" {
  description = "Sync schedule (cron format)"
  value       = var.sync_schedule
}

output "timezone" {
  description = "Sync schedule timezone"
  value       = var.sync_timezone
}
