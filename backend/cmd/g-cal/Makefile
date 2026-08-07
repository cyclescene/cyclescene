PROJECT_ID := cyclescene-479119
REGION := us-west1
REPOSITORY := cyclescene
IMAGE_NAME := g-cal
LOCAL_TAG := $(IMAGE_NAME):latest
REMOTE_TAG := $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)/$(IMAGE_NAME):latest

.PHONY: build push deploy deploy-all

build:
	docker build -t $(LOCAL_TAG) -f Dockerfile ../..

push: build
	gcloud auth configure-docker $(REGION)-docker.pkg.dev --quiet
	docker tag $(LOCAL_TAG) $(REMOTE_TAG)
	docker push $(REMOTE_TAG)

deploy:
	cd infra && tofu init && tofu apply

deploy-all: push deploy
