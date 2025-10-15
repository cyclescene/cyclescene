PROJECT_ID := cyclescene
REGION := us-west1
REPO_NAME := cyclescene


SCRAPER_SERVICE_NAME := pdx-scraper
SCRAPER_LOCAL_TAG := pdx-scraper:latest
SCRAPER_IMAGE_NAME := scraper-image
SCRAPER_IMAGE_TAG := $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPO_NAME)/$(SCRAPER_SERVICE_NAME)/$(SCRAPER_IMAGE_NAME):latest
SCRAPER_DOCKERFILE := Dockerfile.Scraper

API_SERVICE_NAME := cyclescene-api-gateway
API_LOCAL_TAG := cyclescene-api
API_IMAGE_NAME := cyclescene-api-image
API_IMAGE_TAG := $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPO_NAME)/$(API_SERVICE_NAME)/$(API_IMAGE_NAME):latest
API_DOCKERFILE := ./Dockerfile.API

CLEANER_SERVICE_NAME := submission-token-cleaner
CLEANER_LOCAL_TAG := cs-token-cleaner:latest
CLEANER_IMAGE_NAME := token-cleaner-image
CLEANER_IMAGE_TAG := $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPO_NAME)/$(CLEANER_SERVICE_NAME)/$(CLEANER_IMAGE_NAME):latest
CLEANER_DOCKERFILE := ./Dockerfile.ExpTokenCleaner


BACKUP_SERVICE_NAME := turso-backup-job
BACKUP_SERVICE_ACCOUNT := cyclescene-backup-uploader@$(PROJECT_ID).iam.gserviceaccount.com
BACKUP_LOCAL_TAG := turso-backup:latest
BACKUP_IMAGE_NAME := turso-backup-image
BACKUP_IMAGE_TAG := $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPO_NAME)/$(BACKUP_SERVICE_NAME)/$(BACKUP_IMAGE_NAME):latest
BACKUP_DOCKERFILE := Dockerfile.Backup

.PHONY: build-local-scraper auth-docker push-image deploy-job setup-repo

# GCP Docker Repo Setup
setup-repo:
	@echo "--- Creating Artifact Registry Repository: $(REPO_NAME)"
	gcloud artifacts repositories create $(REPO_NAME) \
		--repository-format=docker \
		--location=$(REGION) \
		--description="docker repo for cyclescene" \
		--project=$(PROJECT_ID) || true
	@echo "-- REPOSITORY SETUP COMPLETE ---"

# Scraper build and deploy commands
build-local-scraper: setup-repo
	@echo "--- Building Local Docker Image ---"
	docker build -t $(SCRAPER_LOCAL_TAG) . -f $(SCRAPER_DOCKERFILE)

auth-docker:
	@echo "--- Configuring Docker for Artifact Registry ---"
	gcloud auth configure-docker $(REGION)-docker.pkg.dev --quiet

push-scraper-image: auth-docker build-local-scraper
	@echo "-- Tagging and Pushing image to Artifact Registry ---"
	docker tag $(SCRAPER_LOCAL_TAG) $(SCRAPER_IMAGE_TAG)
	docker push $(SCRAPER_IMAGE_TAG)

deploy-scraper: push-scraper-image
	@echo "--- Deploying Cloud Run Job ---"
	gcloud run jobs deploy $(SCRAPER_SERVICE_NAME) \
		--image $(SCRAPER_IMAGE_TAG) \
		--region $(REGION) \
		--env-vars-file ./runtime.yml 
	@echo "--- DEPLOYMENT COMPLETE ---"

# API commands to build and deploy
build-local-api: setup-repo
	@echo "--- Building Local Docker Image ---"
	docker build -t $(API_LOCAL_TAG) . -f $(API_DOCKERFILE)

push-api-image: auth-docker build-local-api
	@echo "-- Tagging and Pushing image to Artifact Registry ---"
	docker tag $(API_LOCAL_TAG) $(API_IMAGE_TAG)
	docker push $(API_IMAGE_TAG)

deploy-api-service: push-api-image
	@echo "--- Deploying API to Cloud Run Services ---"
	gcloud run deploy $(API_SERVICE_NAME) \
		--image $(API_IMAGE_TAG) \
		--region $(REGION) \
		--env-vars-file ./runtime.yml

# Expired Submission Token Cleaner build and deploy
build-local-cleaner: setup-repo
	@echo "--- Building Local Docker Image ---"
	docker build -t $(CLEANER_LOCAL_TAG) . -f $(CLEANER_DOCKERFILE)

push-cleaner-image: auth-docker build-local-cleaner
	@echo "-- Tagging and Pushing image to Artifact Registry ---"
	docker tag $(CLEANER_LOCAL_TAG) $(CLEANER_IMAGE_TAG)
	docker push $(CLEANER_IMAGE_TAG)

deploy-cleaner: push-cleaner-image
	@echo "--- Deploying Cloud Run Job ---"
	gcloud run jobs deploy $(CLEANER_SERVICE_NAME) \
		--image $(CLEANER_IMAGE_TAG) \
		--region $(REGION) \
		--env-vars-file ./runtime.yml 
	@echo "--- DEPLOYMENT COMPLETE ---"


build-local-backup: setup-repo # Assumes setup-repo is defined earlier
	@echo "--- Building Local Backup Docker Image ---"
	docker build -t $(BACKUP_LOCAL_TAG) . -f $(BACKUP_DOCKERFILE)

push-backup-image: auth-docker build-local-backup
	@echo "-- Tagging and Pushing Backup image to Artifact Registry ---"
	docker tag $(BACKUP_LOCAL_TAG) $(BACKUP_IMAGE_TAG)
	docker push $(BACKUP_IMAGE_TAG)

deploy-backup: push-backup-image
	@echo "--- Deploying Turso Backup Cloud Run Job ---"
	gcloud run jobs deploy $(BACKUP_SERVICE_NAME) \
		--image $(BACKUP_IMAGE_TAG) \
		--region $(REGION) \
		--service-account $(BACKUP_SERVICE_ACCOUNT) \
		--env-vars-file ./runtime.yml 
	@echo "--- BACKUP DEPLOYMENT COMPLETE ---"

deploy-all: deploy-scraper deploy-api-service deploy-cleaner
