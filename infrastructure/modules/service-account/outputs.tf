output "email" {
  description = "Email address of the service account"
  value       = google_service_account.account.email
}

output "name" {
  description = "Name of the service account"
  value       = google_service_account.account.name
}

output "account_id" {
  description = "Account ID of the service account"
  value       = google_service_account.account.account_id
}

output "unique_id" {
  description = "Unique ID of the service account"
  value       = google_service_account.account.unique_id
}
