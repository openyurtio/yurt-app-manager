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

# Build yurt-app-manager and load image to kind
kind-load:
	bash hack/make-rules/kindload-images.sh

e2e-build:
	hack/make-rules/build-e2e.sh

e2e-test:
	bash hack/run-e2e-tests.sh

# Build binaries and docker images.  
# NOTE: this rule can take time, as we build binaries inside containers
#
# ARGS:
#   REPO: image repo.
#   TAG:  image tag 
#   REGION: in which region this rule is executed, if in mainland China,
#   	set it as cn.
#
# Examples:
#   # compile yurt-app-manager
#   make release REGION=cn REPO= TAG=
#   or
#   make release 
release:
	bash hack/make-rules/release-images.sh

# Push docker images.  
#
# ARGS:
#   REPO: image repo.
#   TAG:  image tag 
#   REGION: in which region this rule is executed, if in mainland China,
#   	set it as cn.
#
# Examples:
#   # compile yurt-app-manager
#   make push REGION=cn REPO= TAG=
#   or
#   make push 

push: 
	bash hack/make-rules/push-images.sh

clean: 
	-rm -Rf _output
	-rm -Rf dockerbuild

generate: controller-gen generate-manifests generate-goclient

# Generate manifests, e.g., CRD, RBAC etc.
generate-manifests: controller-gen
	$(CONTROLLER_GEN) crd:trivialVersions=true rbac:roleName=manager-role webhook paths="./pkg/yurtappmanager/..." output:crd:artifacts:config=config/yurt-app-manager/crd/bases  output:rbac:artifacts:config=config/yurt-app-manager/rbac output:webhook:artifacts:config=config/yurt-app-manager/webhook

# Generate go codes.
generate-goclient: controller-gen
	hack/make-rules/generate_client.sh
	$(CONTROLLER_GEN) object:headerFile="./pkg/yurtappmanager/hack/boilerplate.go.txt" paths="./pkg/yurtappmanager/apis/..."


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

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen-openyurt))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	echo "replace sigs.k8s.io/controller-tools => github.com/openkruise/controller-tools v0.2.9-kruise" >> go.mod ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.9 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	mv $(GOPATH)/bin/controller-gen $(GOPATH)/bin/controller-gen-openyurt ;\
	}
CONTROLLER_GEN=$(GOPATH)/bin/controller-gen-openyurt
else
CONTROLLER_GEN=$(shell which controller-gen-openyurt)
endif

install-golint: ## check golint if not exist install golint tools
ifeq (, $(shell which golangci-lint))
	@{ \
	set -e ;\
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.31.0 ;\
	}
GOLINT_BIN=$(shell go env GOPATH)/bin/golangci-lint
else
GOLINT_BIN=$(shell which golangci-lint)
endif

lint: install-golint ## Run go lint against code.
	$(GOLINT_BIN) run -v