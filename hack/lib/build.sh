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

#!/usr/bin/env bash

set -x

# project_info generates the project information and the corresponding valuse 
# for 'ldflags -X' option
project_info() {
    PROJECT_INFO_PKG=${YURT_MOD}/pkg/projectinfo
    echo "-X ${PROJECT_INFO_PKG}.projectPrefix=${PROJECT_PREFIX}"
    echo "-X ${PROJECT_INFO_PKG}.labelPrefix=${LABEL_PREFIX}"
    echo "-X ${PROJECT_INFO_PKG}.gitVersion=${GIT_VERSION}"
    echo "-X ${PROJECT_INFO_PKG}.gitCommit=${GIT_COMMIT}"
    echo "-X ${PROJECT_INFO_PKG}.buildDate=${BUILD_DATE}"
}

# get_binary_dir_with_arch generated the binary's directory with GOOS and GOARCH.
# eg: ./_output/bin/darwin/arm64/
get_binary_dir_with_arch(){
    echo $1/$(go env GOOS)/$(go env GOARCH)/
}


build_binary() {
    local goflags goldflags gcflags
    goldflags="${GOLDFLAGS:--s -w $(project_info)}"
    gcflags="${GOGCFLAGS:-}"
    goflags=${GOFLAGS:-}

    local arg
    for arg; do
      if [[ "${arg}" == -* ]]; then
        # Assume arguments starting with a dash are flags to pass to go.
        goflags+=("${arg}")
      fi
    done

    local bin_dir=${YURT_BIN_DIR}
    local bin_name=${BIN_NAME:-yurt-app-manager}
    mkdir -p ${bin_dir}
    
    echo "Building ${bin_name}"
    go build -o ${bin_dir}/${bin_name} \
        -ldflags "${goldflags:-}" \
        -gcflags "${gcflags:-}" ${goflags} 
}


