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
# gitVersion: "vX.Y" used to indicate the last release version
# gitCommit: the git commit id corresponding to this source code
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
    echo $1/$(go env GOOS)/$(go env GOARCH)
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

    local bin_name=${BIN_NAME:-yurt-app-manager}
    local target_bin_dir=$(get_binary_dir_with_arch ${YURT_LOCAL_BIN_DIR})
    rm -rf ${target_bin_dir}
    mkdir -p ${target_bin_dir}
    
    echo "Building ${bin_name}"
    go build -o ${target_bin_dir}/${bin_name} \
        -ldflags "${goldflags:-}" \
        -gcflags "${gcflags:-}" ${goflags} ${YURT_ROOT}/cmd/yurt-app-manager
}

# gen_yamls generates yaml files for the yurt-app-manager
gen_yamls() {
    local OUT_YAML_DIR=$YURT_ROOT/_output/yamls
    local BUILD_YAML_DIR=${OUT_YAML_DIR}/build
    [ -f $BUILD_YAML_DIR ] || mkdir -p $BUILD_YAML_DIR
    mkdir -p ${BUILD_YAML_DIR}
    (
        rm -rf ${BUILD_YAML_DIR}/yurt-app-manager
        cp -rf $YURT_ROOT/config/yurt-app-manager ${BUILD_YAML_DIR}
        cd ${BUILD_YAML_DIR}/yurt-app-manager/manager
        kustomize edit set image controller=$REPO/yurt-app-manager:${TAG}
	)
    set +x
    echo "==== create yurt-app-manager.yaml in $OUT_YAML_DIR ===="
    kustomize build ${BUILD_YAML_DIR}/yurt-app-manager/default > ${OUT_YAML_DIR}/yurt-app-manager.yaml
    rm -Rf ${BUILD_YAML_DIR}
}

