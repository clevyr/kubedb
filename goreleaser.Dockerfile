FROM alpine:3.24.1
WORKDIR /data

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/kubedb /usr/local/bin

ENV KUBECONFIG /.kube/config
ENTRYPOINT ["kubedb"]
