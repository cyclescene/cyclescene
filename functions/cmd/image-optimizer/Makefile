PROJECT_ID := cyclescene-479119
REGION := us-west1
REPO_NAME := cyclescene
SERVICE_NAME := image-optimizer
IMAGE_NAME := image-optimizer-image
LOCAL_TAG := cyclescene-image-optimizer:latest
IMAGE_TAG := $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPO_NAME)/$(SERVICE_NAME)/$(IMAGE_NAME):latest

.PHONY: help build push deploy deploy-all clean

help:
	@echo "CycleScene Image Optimizer Deployment Commands:"
	@echo ""
	@echo "  make build       - Build Docker image"
	@echo "  make push        - Push Docker image to Artifact Registry"
	@echo "  make deploy      - Deploy infrastructure with OpenTofu"
	@echo "  make deploy-all  - Build, push, and deploy everything"
	@echo "  make clean       - Clean local Docker images and Terraform files"

# Build Docker image (context is functions root)
build:
	@echo "--- Building Docker image ---"
	docker build -t $(LOCAL_TAG) -f Dockerfile ../..

# Push image to Artifact Registry
push: build
	@echo "--- Configuring Docker for Artifact Registry ---"
	gcloud auth configure-docker $(REGION)-docker.pkg.dev --quiet
	@echo "--- Tagging and pushing image ---"
	docker tag $(LOCAL_TAG) $(IMAGE_TAG)
	docker push $(IMAGE_TAG)

# Deploy infrastructure with OpenTofu
deploy:
	@echo "--- Deploying infrastructure ---"
	cd infra && tofu init && tofu apply -auto-approve

# Full deployment pipeline
deploy-all: push deploy

# Clean up
clean:
	@echo "--- Cleaning up ---"
	docker rmi $(LOCAL_TAG) $(IMAGE_TAG) 2>/dev/null || true
	cd infra && rm -rf .terraform .terraform.lock.hcl tfplan
