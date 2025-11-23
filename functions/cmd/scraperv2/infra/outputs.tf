output "job_name" {
  description = "Name of the scraper job"
  value       = module.scraper_job.job_name
}

output "job_id" {
  description = "ID of the scraper job"
  value       = module.scraper_job.job_id
}

output "schedule_name" {
  description = "Name of the scheduler job"
  value       = module.scraper_schedule.job_name
}
