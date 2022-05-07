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
readonly yurt_component_image="${REPO}/${bin_target}:${TAG}"

build_multi_arch_binaries() {
    local docker_yurt_root="/opt/src"
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

    [[ -n ${http_proxy+x} ]] && docker_run_opts+=("--env http_proxy=${http_proxy}")
    [[ -n ${https_proxy+x} ]] && docker_run_opts+=("--env https_proxy=${https_proxy}")

    local docker_run_cmd=(
        "/bin/sh"
        "-xe"
        "-c"
    )

    local sub_commands="sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories; \
        apk --no-cache add bash git; \
        cd ${docker_yurt_root}; umask 0022; \
        rm -rf ${YURT_BIN_DIR}/* ; \
        git config --global --add safe.directory ${docker_yurt_root};"
    sub_commands+="GOARCH=amd64 bash ./hack/make-rules/build.sh ${bin_target}; "
    sub_commands+="chown -R $(id -u):$(id -g) ${docker_yurt_root}/_output"

    docker run ${docker_run_opts[@]} ${YURT_BUILD_IMAGE} ${docker_run_cmd[@]} "${sub_commands}"
}

build_docker_image() {
    local binary_path=${YURT_BIN_DIR}/${bin_target}
    if [ -f ${binary_path} ]; then
        local docker_build_path=${DOCKER_BUILD_BASE_IDR}/
        local docker_file_path=${docker_build_path}/Dockerfile.${bin_target}
        mkdir -p ${docker_build_path}
 
        local base_image="k8s.gcr.io/debian-base-amd64:v1.0.0"
        cat <<EOF > "${docker_file_path}"
FROM ${base_image}
COPY ${bin_target} /usr/local/bin/${bin_target}
ENTRYPOINT ["/usr/local/bin/${bin_target}"]
EOF

        ln "${binary_path}" "${docker_build_path}/${bin_target}"
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
    gen_yamls
}

push_images() {
    build_images
    docker push ${yurt_component_image}
}

kindload_images() {
    build_images
    kind load docker-image ${yurt_component_image} || { echo >&2 "kind not installed or error loading image: $(yurt_component_image)"; exit 1; }
}

# gen_yamls generates yaml files for the yurt-app-manager 
gen_yamls() {
    local OUT_YAML_DIR=$YURT_ROOT/_output/yamls/
    local BUILD_YAML_DIR=${OUT_YAML_DIR}/build/
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
