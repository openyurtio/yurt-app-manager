#!/usr/bin/env bash

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

set -o errexit
set -o nounset
set -o pipefail

YURT_MOD="$(head -1 $YURT_ROOT/go.mod | awk '{print $2}')"
YURT_OUTPUT_DIR=${YURT_ROOT}/_output
YURT_LOCAL_BIN_DIR=${YURT_OUTPUT_DIR}/local/bin

YURT_E2E_TARGETS="tests/e2e/yurt-e2e-test"

PROJECT_PREFIX=${PROJECT_PREFIX:-yurt}
LABEL_PREFIX=${LABEL_PREFIX:-openyurt.io}
GIT_COMMIT=$(git rev-parse --short HEAD)
GIT_COMMIT_SHORT=$GIT_COMMIT
GIT_VERSION=${GIT_VERSION:-$(git describe --abbrev=0 --tags)}
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
REPO=${REPO:-openyurt}
TAG=${TAG:-${GIT_COMMIT_SHORT}}
BIN_NAME=yurt-app-manager