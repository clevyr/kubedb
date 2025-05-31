FROM alpine:3.22.0
LABEL org.opencontainers.image.source="https://github.com/clevyr/kubedb"
WORKDIR /data

COPY kubedb /usr/local/bin

ENV KUBECONFIG /.kube/config
ENTRYPOINT ["kubedb"]
