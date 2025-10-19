# API Gateway Configuration
project_id                = "cyclescene"
region                    = "us-west1"
environment               = "production"
api_custom_domain         = "api.cyclescene.cc"
api_service_account_email = "cyclescene-api@cyclescene.iam.gserviceaccount.com"

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

# Database & API credentials
turso_db_url       = "libsql://cyclescene-spacesedan.aws-us-west-2.turso.io"
turso_db_rw_token  = "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJhIjoicnciLCJpYXQiOjE3NTk2MDI3NjYsImlkIjoiYTk3MWQ4NjYtNmYyMC00MTliLTg3NzItNjUxOGQwNjFiMWViIiwicmlkIjoiNTM1YzdmY2MtOTFkNi00ZWUzLTlkOGUtMjJhMGNiY2QzNWU0In0.FsnvHvKo6mo15dLttjm1ljUnzjs0XHOHQ0leynLMD_Vj9X4sqwyq_Ve3CA3hOL3BbHlK8nEM226JYLA0ZADFCw"
staging_bucket_name = "cyclescene-user-media-staging"

# Image Optimizer
optimizer_service_account_email = "cyclescene-image-optimizer@cyclescene.iam.gserviceaccount.com"
optimizer_cpu_limit             = "2"
optimizer_memory_limit          = "2Gi"
optimizer_max_instances         = 5
image_optimizer_url             = "https://cyclescene-image-optimizer-556687167657.us-west1.run.app"
