resource "google_cloud_run_v2_job" "job" {
  name     = var.job_name
  location = var.region

  template {
    template {
      service_account = var.service_account_email

      max_retries = var.max_retries
      timeout     = var.timeout

      containers {
        image = var.image

        # Environment variables
        dynamic "env" {
          for_each = var.env_vars
          content {
            name  = env.key
            value = env.value
          }
        }

        resources {
          limits = {
            cpu    = var.cpu_limit
            memory = var.memory_limit
          }
        }

        # Args and command are simple fields, not blocks
        args    = length(var.args) > 0 ? var.args : null
        command = length(var.command) > 0 ? var.command : null
      }

      dynamic "vpc_access" {
        for_each = var.vpc_connector != null ? [1] : []
        content {
          connector = var.vpc_connector
          egress    = var.vpc_egress
        }
      }
    }

    parallelism  = var.parallelism
    task_count   = var.task_count
  }

  labels = var.labels
}
