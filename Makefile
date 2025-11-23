.PHONY: help bootstrap bootstrap-infra deploy-api deploy-all clean

help:
	@echo "CycleScene Root Makefile Commands:"
	@echo ""
	@echo "  make bootstrap          - Set up Terraform backend bucket and bootstrap infrastructure"
	@echo "  make bootstrap-infra    - Create Terraform state bucket in GCS"
	@echo "  make deploy-api         - Deploy API service"
	@echo "  make deploy-all         - Deploy all services"
	@echo "  make clean              - Clean up all local Terraform state"
	@echo ""

# Bootstrap infrastructure (create GCS bucket for Terraform state)
bootstrap-infra:
	@echo "--- Bootstrapping Terraform Backend ---"
	cd infrastructure/bootstrap && tofu init && tofu apply -auto-approve
	@echo "✓ Terraform backend bucket created"

# Full bootstrap (includes infrastructure setup)
bootstrap: bootstrap-infra
	@echo "✓ Bootstrap complete - ready to deploy services"

# Deploy individual services
deploy-api:
	@echo "--- Deploying API Service ---"
	cd functions/cmd/api && make deploy-all

deploy-image-optimizer:
	@echo "--- Deploying Image Optimizer ---"
	cd functions/cmd/image-optimizer && make deploy-all

deploy-scraper:
	@echo "--- Deploying Scraper ---"
	cd functions/cmd/scraperv2 && make deploy-all

deploy-token-cleaner:
	@echo "--- Deploying Token Cleaner ---"
	cd functions/cmd/token-cleaner && make deploy-all

deploy-db-backups:
	@echo "--- Deploying DB Backups ---"
	cd functions/cmd/db-backups && make deploy-all

# Deploy all services in sequence
deploy-all: deploy-api deploy-image-optimizer deploy-scraper deploy-token-cleaner deploy-db-backups
	@echo "✓ All services deployed"

# Clean up local Terraform state across all services
clean:
	@echo "--- Cleaning up Terraform state ---"
	find . -path "./functions/cmd/*/infra/.terraform*" -type d -exec rm -rf {} + 2>/dev/null || true
	find . -path "./infrastructure/bootstrap/.terraform*" -type d -exec rm -rf {} + 2>/dev/null || true
	find . -name "terraform.tfstate*" -delete
	@echo "✓ Cleanup complete"
