# syntax=docker/dockerfile:1.7

ARG GO_IMAGE=golang:1.26.5-alpine3.23
ARG RUNTIME_IMAGE=alpine:3.23.3

FROM ${GO_IMAGE} AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
ARG VERSION=dev
ARG VCS_REF=unknown
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath \
      -ldflags="-s -w -X go-admin/common/global.Version=${VERSION}" \
      -o /out/go-admin ./cmd/go-admin && \
    go version -m /out/go-admin

FROM ${RUNTIME_IMAGE} AS runtime
ARG VERSION=dev
ARG VCS_REF=unknown

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S -g 10001 goadmin && \
    adduser -S -D -H -u 10001 -G goadmin goadmin && \
    install -d -o goadmin -g goadmin -m 0750 \
      /etc/go-admin \
      /srv/go-admin/static/uploadfile \
      /tmp/go-admin/logs

COPY --from=build --chown=10001:10001 /out/go-admin /usr/local/bin/go-admin

LABEL org.opencontainers.image.title="go-admin-plus API" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.source="https://github.com/NAMEWTA/go-admin-plus"

USER 10001:10001
WORKDIR /srv/go-admin
EXPOSE 8000

ENTRYPOINT ["/usr/local/bin/go-admin"]
CMD ["server", "--profile", "server", "--config", "/etc/go-admin/settings.yml", "--files", "/srv/go-admin/static/uploadfile"]
