# Cloud Run Domain Mapping Module
# Manages custom domain mappings for Cloud Run services

resource "google_cloud_run_domain_mapping" "domain" {
  location = var.location
  name     = var.domain_name

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = var.cloud_run_service_name
  }
}

# Output the DNS records needed for verification
output "dns_records" {
  description = "DNS records to add to your domain registrar"
  value       = try(google_cloud_run_domain_mapping.domain.status[0].resource_records, [])
  sensitive   = false
}
