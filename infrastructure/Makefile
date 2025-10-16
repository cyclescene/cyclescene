PROJECT_ID := cyclescene
REGION := us-west1
ENVIRONMENT := dev
STATE_BUCKET := $(PROJECT_ID)-terraform-state

.PHONY: help init plan apply deploy destroy clean

help:
	@echo "Infrastructure Deployment Commands:"
	@echo ""
	@echo "  make deploy      - Deploy infrastructure (init + apply)"
	@echo "  make plan        - Preview infrastructure changes"
	@echo "  make apply       - Apply infrastructure changes"
	@echo "  make destroy     - Destroy all infrastructure"
	@echo "  make clean       - Clean local OpenTofu files"
	@echo ""
	@echo "Variables (override with make VAR=value):"
	@echo "  PROJECT_ID=$(PROJECT_ID)"
	@echo "  REGION=$(REGION)"
	@echo "  ENVIRONMENT=$(ENVIRONMENT)"

setup-backend:
	@echo "--- Setting up Terraform state backend ---"
	@gsutil ls -b gs://$(STATE_BUCKET) 2>/dev/null || \
		(gsutil mb -p $(PROJECT_ID) -l $(REGION) gs://$(STATE_BUCKET) && \
		 gsutil versioning set on gs://$(STATE_BUCKET))
	@echo "--- Backend setup complete ---"

init: setup-backend
	@echo "--- Initializing OpenTofu ---"
	tofu init \
		-backend-config="bucket=$(STATE_BUCKET)" \
		-backend-config="prefix=terraform/state/$(ENVIRONMENT)"
	@echo "--- Initialization complete ---"

plan: init
	@echo "--- Planning infrastructure changes ---"
	tofu plan \
		-var="project_id=$(PROJECT_ID)" \
		-var="region=$(REGION)" \
		-var="environment=$(ENVIRONMENT)"

apply: init
	@echo "--- Applying infrastructure changes ---"
	tofu apply \
		-var="project_id=$(PROJECT_ID)" \
		-var="region=$(REGION)" \
		-var="environment=$(ENVIRONMENT)" \
		-auto-approve
	@echo "--- Deployment complete ---"

deploy: apply

destroy: init
	@echo "--- Destroying infrastructure ---"
	tofu destroy \
		-var="project_id=$(PROJECT_ID)" \
		-var="region=$(REGION)" \
		-var="environment=$(ENVIRONMENT)" \
		-auto-approve
	@echo "--- Destruction complete ---"

clean:
	@echo "--- Cleaning local OpenTofu files ---"
	rm -rf .terraform .terraform.lock.hcl tfplan tfplan.*
	@echo "--- Clean complete ---"
