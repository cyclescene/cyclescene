PROJECT_ID := cyclescene
REGION := us-west1
BUILD_IMAGE := go125
TIMEOUT := 540s
MEMORY := 256M
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

.PHONY: build-local-scraper auth-docker push-image deploy-job setup-repo

setup-repo:
	@echo "--- Creating Artifact Registry Repository: $(REPO_NAME)"
	gcloud artifacts repositories create $(REPO_NAME) \
		--repository-format=docker \
		--location=$(REGION) \
		--description="docker repo for cyclescene" \
		--project=$(PROJECT_ID) || true
	@echo "-- REPOSITORY SETUP COMPLETE ---"

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

deploy-job: push-scraper-image
	@echo "--- Deploying Cloud Run Job ---"
	gcloud run jobs deploy $(SCRAPER_SERVICE_NAME) \
		--image $(SCRAPER_IMAGE_TAG) \
		--region $(REGION) \
		--env-vars-file ./runtime.yml 
	@echo "--- DEPLOYMENT COMPLETE ---"


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

deploy-all: deploy-job deploy-api-service
