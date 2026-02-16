#syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.26.0-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache \
  CGO_ENABLED=0 GOOS="$TARGETOS" GOARCH="$TARGETARCH" \
  go build -ldflags='-w -s' -trimpath -tags disable_grpc_modules,grpcnotrace


FROM alpine:3.23.3
LABEL org.opencontainers.image.source="https://github.com/clevyr/kubedb"
WORKDIR /data

COPY --from=builder /app/kubedb /usr/local/bin/

ENV KUBECONFIG /.kube/config
ENTRYPOINT ["kubedb"]
