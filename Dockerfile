ARG GO_VERSION=1.11.5
FROM golang:${GO_VERSION}-alpine

RUN apk --no-cache add ca-certificates make git
WORKDIR /src
COPY go.mod go.sum Makefile ./
RUN make vendor
COPY . .
RUN make test all


FROM alpine:latest
VOLUME /output
VOLUME /config
VOLUME /manifests

ENV KUBECTL_VERSION=1.8.9 \
    KUBECTL_CHECKSUM=dc49a95b460585c3910264b4a44d717bb5b7a3e1c5f18315cb15662e99a0d231

# install kubectl
RUN apk update && \
    apk add curl && \
    curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/bin/kubectl && \
    apk del curl

COPY --from=0 /src/bin/k8s-config-projector /bin/k8s-config-projector
ENTRYPOINT [ \
  "k8s-config-projector" \
]

CMD [ \
  "--manifests=/manifests", \
  "--config-repo=/config", \
  "--output=/output" \
]
