
# Image URL to use all building/pushing image targets
IMG ?= ctrlarm-controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Directories
TOOLS_DIR := $(PWD)/hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
BIN_DIR := bin

# Binaries
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/controller-gen
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint
KUSTOMIZE := $(TOOLS_BIN_DIR)/kustomize

all: manager

# Run tests
test: generate lint manifests
	go test ./... -coverprofile cover.out

# Binaries

manager: generate lint-full
	go build -o bin/manager main.go

$(CONTROLLER_GEN): $(TOOLS_DIR)/go.mod # Build controller-gen from tools folder.
	cd $(TOOLS_DIR); go build -tags=tools -o $(BIN_DIR)/controller-gen sigs.k8s.io/controller-tools/cmd/controller-gen

$(GOLANGCI_LINT): $(TOOLS_DIR)/go.mod # Build golangci-lint from tools folder.
	cd $(TOOLS_DIR); go build -tags=tools -o $(BIN_DIR)/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

$(KUSTOMIZE): $(TOOLS_DIR)/go.mod # Build kustomize from tools folder.
	cd $(TOOLS_DIR); go build -tags=tools -o $(BIN_DIR)/kustomize sigs.k8s.io/kustomize/kustomize/v3

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate lint manifests
	go run ./main.go

# Install CRDs into a cluster
install: $(KUSTOMIZE) manifests
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: $(KUSTOMIZE) manifests
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | envsubst | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Linting
.PHONY: lint lint-full
lint: $(GOLANGCI_LINT) ## Lint codebase
	$(GOLANGCI_LINT) run -v

lint-full: $(GOLANGCI_LINT) ## Run slower linters to detect possible issues
	$(GOLANGCI_LINT) run -v --fast=false

# Generate code
generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build: generate manifests
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

docker-ci: generate manifests
	go build -o bin/manager main.go
	docker login ${REGISTRY} -u ${REGISTRY_USERNAME} -p ${REGISTRY_PASSWORD}
	docker build . -t ${IMG}
	docker push ${IMG}

# Release

RELEASE_DIR := out

$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)/

.PHONY: release
release: release-containers release-manifests

.PHONY: release-manifests
release-manifests: $(KUSTOMIZE) $(RELEASE_DIR)
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > $(RELEASE_DIR)/ctrlarm-components.yaml

.PHONY: release-containers
release-containers: $(RELEASE_DIR)
	$(MAKE) docker-build docker-push

