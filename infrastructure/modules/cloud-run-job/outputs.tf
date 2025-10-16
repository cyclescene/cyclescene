output "job_name" {
  description = "Name of the Cloud Run job"
  value       = google_cloud_run_v2_job.job.name
}

output "job_id" {
  description = "ID of the Cloud Run job"
  value       = google_cloud_run_v2_job.job.id
}

output "job_location" {
  description = "Location of the Cloud Run job"
  value       = google_cloud_run_v2_job.job.location
}
