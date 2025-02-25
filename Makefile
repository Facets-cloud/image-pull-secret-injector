SHELL := /bin/bash

# Image URL to use all building/pushing image targets
IMG ?= vishnukvfacets/image-pull-secrets:1.0.9

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: all
all: docker deploy

.PHONY: docker
docker:
	docker buildx build --push --platform linux/amd64 -t ${IMG} -f Dockerfile .

.PHONY: deploy
deploy:
	kubectl apply -f config/certmanager/certificate.yaml
	kubectl apply -f config/rbac/role.yaml
	kubectl apply -f config/rbac/role_binding.yaml
	kubectl apply -f config/rbac/service_account.yaml
	kubectl apply -f config/webhook/deployment.yaml
	kubectl apply -f config/webhook/service.yaml
	kubectl apply -f config/webhook/manifests.yaml
	kubectl apply -f config/default/webhookcainjection_patch.yaml