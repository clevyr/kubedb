FROM alpine:3.23.2
LABEL org.opencontainers.image.source="https://github.com/clevyr/kubedb"
WORKDIR /data

COPY kubedb /usr/local/bin

ENV KUBECONFIG /.kube/config
ENTRYPOINT ["kubedb"]
