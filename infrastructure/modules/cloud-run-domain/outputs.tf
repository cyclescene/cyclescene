output "domain_name" {
  description = "The custom domain name"
  value       = google_cloud_run_domain_mapping.domain.name
}

output "domain_status" {
  description = "Status of the domain mapping"
  value       = try(google_cloud_run_domain_mapping.domain.status[0].conditions[0].message, "Pending verification")
}

output "mapped_route_name" {
  description = "The Cloud Run service this domain is mapped to"
  value       = try(google_cloud_run_domain_mapping.domain.spec[0].route_name, null)
}
