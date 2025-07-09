#syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.6.1 AS xx

FROM --platform=$BUILDPLATFORM golang:1.24.5-alpine AS builder
WORKDIR /app

COPY --from=xx / /

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Set Golang build envs based on Docker platform string
ARG TARGETPLATFORM
RUN --mount=type=cache,target=/root/.cache \
  CGO_ENABLED=0 xx-go build -ldflags='-w -s' -trimpath -tags disable_grpc_modules,grpcnotrace


FROM alpine:3.22.0
LABEL org.opencontainers.image.source="https://github.com/clevyr/kubedb"
WORKDIR /data

COPY --from=builder /app/kubedb /usr/local/bin/

ENV KUBECONFIG /.kube/config
ENTRYPOINT ["kubedb"]
