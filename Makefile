#Arches can be: amd64 s390x arm64 ppc64le
ARCH ?= ppc64le

REGISTRY ?= quay.io/powercloud
REPOSITORY ?= power-dev-plugin
TAG ?= latest

CONTAINER_RUNTIME ?= $(shell command -v podman 2> /dev/null || echo docker)

########################################################################
# Go Targets

.PHONY: build
build: fmt vet
	GOOS=linux GOARCH=$(ARCH) go build -o bin/power-dev-plugin cmd/power-dev-plugin/main.go
	GOOS=linux GOARCH=$(ARCH) go build -o bin/power-dev-webhook cmd/webhook/main.go

.PHONY: build-scanner
build-scanner: fmt vet
	GOOS=linux GOARCH=$(ARCH) go build -o bin/devices-scanner cmd/scanner/main.go

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: clean
clean:
	rm -f ./bin/power-dev-plugin
	rm -rf vendor

########################################################################
# Container Targets

.PHONY: image
image:
	$(CONTAINER_RUNTIME) buildx build \
		-t $(REGISTRY)/$(REPOSITORY):$(TAG) \
		--platform linux/$(ARCH) -f build/Containerfile .

.PHONY: push
push:
	$(info push Container image...)
	$(CONTAINER_RUNTIME) push $(REGISTRY)/$(REPOSITORY):$(TAG)

########################################################################
# Deployment Targets

.PHONY: dep-plugin
dep-plugin:
	kustomize build manifests | oc apply -f -

.PHONY: dep-examples
dep-examples:
	kustomize build examples | oc apply -f -