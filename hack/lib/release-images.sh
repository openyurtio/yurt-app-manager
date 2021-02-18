# Copyright 2021 The OpenYurt Authors.
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

YURT_IMAGE_DIR=${YURT_OUTPUT_DIR}/images
YURTCTL_SERVANT_DIR=${YURT_ROOT}/config/yurtctl-servant
DOCKER_BUILD_BASE_IDR=$YURT_ROOT/dockerbuild
YURT_BUILD_IMAGE="golang:1.15-alpine"

readonly SUPPORTED_OS=linux
readonly bin_target=yurt-app-manager
readonly region=${REGION:-us}

build_multi_arch_binaries() {
    local docker_run_opts=(
        "-i"
        "--rm"
        "--network host"
        "-v ${YURT_ROOT}:/opt/src"
        "--env CGO_ENABLED=0"
        "--env GOOS=${SUPPORTED_OS}"
        "--env PROJECT_PREFIX=${PROJECT_PREFIX}"
        "--env LABEL_PREFIX=${LABEL_PREFIX}"
        "--env GIT_VERSION=${GIT_VERSION}"
        "--env GIT_COMMIT=${GIT_COMMIT}"
        "--env BUILD_DATE=${BUILD_DATE}"
    )
    # use goproxy if build from inside mainland China
    [[ $region == "cn" ]] && docker_run_opts+=("--env GOPROXY=https://goproxy.cn")

    local docker_run_cmd=(
        "/bin/sh"
        "-xe"
        "-c"
    )

    local sub_commands="sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories; \
        apk --no-cache add bash git; \
        cd /opt/src; umask 0022; \
        rm -rf ${YURT_BIN_DIR}/* ;"
    sub_commands+="GOARCH=amd64 bash ./hack/make-rules/build.sh ${bin_target}; "
    sub_commands+="chown -R $(id -u):$(id -g) /opt/src/_output"

    docker run ${docker_run_opts[@]} ${YURT_BUILD_IMAGE} ${docker_run_cmd[@]} "${sub_commands}"
}

build_docker_image() {
    local binary_name=$(get_output_name $bin_target)
    local binary_path=${YURT_BIN_DIR}/${binary_name}
    if [ -f ${binary_path} ]; then
        local docker_build_path=${DOCKER_BUILD_BASE_IDR}/
        local docker_file_path=${docker_build_path}/Dockerfile.${binary_name}
        mkdir -p ${docker_build_path}
 
        local yurt_component_image="${REPO}/${binary_name}:${TAG}"
        local base_image="k8s.gcr.io/debian-base-amd64:v1.0.0"
        cat <<EOF > "${docker_file_path}"
FROM ${base_image}
COPY ${binary_name} /usr/local/bin/${binary_name}
ENTRYPOINT ["/usr/local/bin/${binary_name}"]
EOF

        ln "${binary_path}" "${docker_build_path}/${binary_name}"
        docker build --no-cache -t "${yurt_component_image}" -f "${docker_file_path}" ${docker_build_path}
        rm -rf ${docker_build_path}
    fi
}

build_images() {
    # Always clean up before generating the image
    rm -Rf ${YURT_OUTPUT_DIR}
    rm -Rf ${DOCKER_BUILD_BASE_IDR}
    mkdir -p ${YURT_BIN_DIR}
    mkdir -p ${YURT_IMAGE_DIR}
    mkdir -p ${DOCKER_BUILD_BASE_IDR}
    
    build_multi_arch_binaries
    build_docker_image
}
