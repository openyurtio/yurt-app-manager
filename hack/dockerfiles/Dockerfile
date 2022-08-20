# multi-arch image building for yurthub

FROM --platform=${BUILDPLATFORM} golang:1.17.1 as builder
ADD . /build
ARG TARGETOS TARGETARCH GIT_VERSION GOPROXY MIRROR_REPO
WORKDIR /build/
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GIT_VERSION=${GIT_VERSION:-undefined} make build

FROM --platform=${TARGETPLATFORM} alpine:3.14
ARG TARGETOS TARGETARCH MIRROR_REPO
RUN if [ ! -z "${MIRROR_REPO+x}" ]; then sed -i "s/dl-cdn.alpinelinux.org/${MIRROR_REPO}/g" /etc/apk/repositories; fi && \
    apk add ca-certificates bash libc6-compat && update-ca-certificates && rm /var/cache/apk/*
COPY --from=builder /build/_output/local/bin/${TARGETOS}/${TARGETARCH}/yurt-app-manager /usr/local/bin/yurt-app-manager
ENTRYPOINT ["/usr/local/bin/yurt-app-manager"]