#Arches can be: amd64 s390x arm64 ppc64le
ARCH ?= arm64

REGISTRY ?= quay.io/powercloud
REPOSITORY ?= power-dev-plugin

build-image:
	+@podman build --platform linux/${ARCH} -t ${REGISTRY}:latest -f Containerfile
.PHONY: build-image

build: fmt vet
	go build -o bin/power-dev-plugin cmd/webhook/main.go
.PHONY: build

fmt:
	go fmt ./...
.PHONY: fmt

vet:
	go vet ./...
.PHONY: vet

clean:
	$(RM) ./bin/power-dev-plugin
.PHONY: clean