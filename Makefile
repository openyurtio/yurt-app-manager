# Copyright 2020 The OpenYurt Authors.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

CRD_OPTIONS ?= "crd:crdVersions=v1"
TARGET_PLATFORMS ?= linux/arm64
IMAGE_REPO ?= openyurt
IMAGE_TAG ?= $(shell git describe --abbrev=0 --tags)
GIT_VERSION = $(shell git describe --abbrev=0 --tags)
GIT_COMMIT = $(shell git rev-parse --short HEAD)

DOCKER_BUILD_ARGS = --build-arg GIT_VERSION=${GIT_VERSION}

ifeq (${REGION}, cn)
DOCKER_BUILD_ARGS += --build-arg GOPROXY=https://goproxy.cn --build-arg MIRROR_REPO=mirrors.aliyun.com
endif

ifneq (${http_proxy},)
DOCKER_BUILD_ARGS += --build-arg http_proxy='${http_proxy}'
endif

ifneq (${https_proxy},)
DOCKER_BUILD_ARGS += --build-arg https_proxy='${https_proxy}'
endif

.PHONY: clean all release build

all: test build

# Build binaries in the host environment
build: 
	bash hack/make-rules/build.sh

# Run test
test: fmt vet
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

docker-build:
	docker buildx build --no-cache --load ${DOCKER_BUILD_ARGS} --platform ${TARGET_PLATFORMS} -f hack/dockerfiles/Dockerfile . -t ${IMAGE_REPO}/yurt-app-manager:${IMAGE_TAG}

docker-push:
	docker buildx rm container-builder || true
	docker buildx create --use --name=yurt-app-manager-container-builder
	# enable qemu for arm64 build
	# https://github.com/docker/buildx/issues/464#issuecomment-741507760
	docker run --privileged --rm tonistiigi/binfmt --uninstall qemu-aarch64
	docker run --rm --privileged tonistiigi/binfmt --install all
	docker buildx build --no-cache --push ${DOCKER_BUILD_ARGS} --platform ${TARGET_PLATFORMS} -f hack/dockerfiles/Dockerfile . -t ${IMAGE_REPO}/yurt-app-manager:${IMAGE_TAG}

clean: 
	-rm -Rf _output
	-rm -Rf dockerbuild

generate: controller-gen manifests generate-goclient

# Generate manifests, e.g., CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." paths="./pkg/yurtappmanager/..." output:crd:artifacts:config=config/yurt-app-manager/crd/bases  output:rbac:artifacts:config=config/yurt-app-manager/rbac output:webhook:artifacts:config=config/yurt-app-manager/webhook

# Generate go codes.
generate-goclient: controller-gen
	hack/make-rules/generate_client.sh
	$(CONTROLLER_GEN) object:headerFile="./pkg/yurtappmanager/hack/boilerplate.go.txt" paths="./pkg/yurtappmanager/apis/..."

gen-all-in-one:
	helm template --include-crds yurt-app-manager charts/yurt-app-manager > config/setup/all_in_one.yaml
# generate deploy yaml.
#
# ARGS:
#   REPO: image repo.
#   TAG:  image tag 
#   REGION: in which region this rule is executed, if in mainland China,
#   	set it as cn.
#
# Examples:
#   # compile yurt-app-manager
#   make generate-deploy-yaml REGION=cn REPO= TAG=
#   or
#   make generate-deploy-yaml 

generate-deploy-yaml: 
	hack/make-rules/genyaml.sh

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
ifeq ("$(shell $(CONTROLLER_GEN) --version 2> /dev/null)", "Version: v0.7.0")
else
	rm -rf $(CONTROLLER_GEN)
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0)
endif

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary.
	$(call go-get-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1)

GINKGO = $(shell pwd)/bin/ginkgo
ginkgo: ## Download ginkgo locally if necessary.
	$(call go-get-tool,$(GINKGO),github.com/onsi/ginkgo/ginkgo@v1.16.4)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef