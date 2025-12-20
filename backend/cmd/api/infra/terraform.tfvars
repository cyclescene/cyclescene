# API Gateway Configuration
project_id = "cyclescene-479119"
region     = "us-west1"
environment = "production"
# api_custom_domain         = "api.cyclescene.cc"  # Commented out - domain already mapped in old project

# API Service Resources
api_cpu_limit    = "2"
api_memory_limit = "1Gi"

# API Scaling
api_min_instances = 0
api_max_instances = 10

# Public Access
api_allow_public = true

# CORS - Update with your actual frontend domains
allowed_origins = [
  "https://cyclescene.cc",
  "https://www.cyclescene.cc",
  "https://form.cyclescene.cc",
  "https://pdx.cyclescene.cc",
  "https://slc.cyclescene.cc",
  "http://localhost:5173",
  "http://localhost:5174",
]

# Database & API credentials (passed via environment variables from GitHub secrets)
# turso_db_url is passed via TF_VAR_turso_db_url
# turso_db_rw_token is passed via TF_VAR_turso_db_rw_token

# Email service (Resend)
# resend_api_key is passed via TF_VAR_resend_api_key environment variable from GitHub secrets

# Edit link base URL for magic link emails
edit_link_base_url = "https://form.cyclescene.cc/rides/edit"

# Image Optimizer
optimizer_cpu_limit             = "2"
optimizer_memory_limit          = "2Gi"
optimizer_max_instances         = 5
image_optimizer_url             = "https://cyclescene-image-optimizer-556687167657.us-west1.run.app"
