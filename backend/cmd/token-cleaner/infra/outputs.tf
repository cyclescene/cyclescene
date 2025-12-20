output "job_name" {
  description = "Name of the token cleaner job"
  value       = module.token_cleaner_job.job_name
}

output "job_id" {
  description = "ID of the token cleaner job"
  value       = module.token_cleaner_job.job_id
}

output "schedule_name" {
  description = "Name of the scheduler job"
  value       = module.token_cleaner_schedule.job_name
}
