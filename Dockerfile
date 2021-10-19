# Build the manager binary
FROM golang:1.16 as builder

ARG ssh_prv_key
ARG ssh_pub_key

# Add the keys and set permissions
RUN mkdir -p /root/.ssh && chmod 0700 /root/.ssh
RUN echo "$ssh_prv_key" | tr -d '\r' > /root/.ssh/id_rsa && \
    echo "$ssh_pub_key" | tr -d '\r' > /root/.ssh/id_rsa.pub && \
    chmod 600 /root/.ssh/id_rsa && \
    chmod 600 /root/.ssh/id_rsa.pub

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download


# Copy the go source
COPY main.go main.go
COPY apis/ apis/
COPY pkg/ pkg/


# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o mgr main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:3.9

ARG APK_MIRROR=mirrors.aliyun.com
ARG CodeSource=
ARG CodeBranches=
ARG CodeVersion=

ENV CODE_SOURCE=$CodeSource
ENV CODE_BRANCHES=$CodeBranches
ENV CODE_VERSION=$CodeVersion

LABEL CodeSource=$CodeSource CodeBranches=$CodeBranches CodeVersion=$CodeVersion
LABEL PERF_BUSINESS_TYPE=polar-common-domain

WORKDIR /
COPY --from=builder /workspace/mgr .
COPY --from=builder /workspace/pkg/workflow /usr/local/polar-as/k8s/workflow

ENTRYPOINT ["./mgr"]
