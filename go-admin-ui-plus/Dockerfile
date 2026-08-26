# syntax=docker/dockerfile:1.7

ARG NODE_IMAGE=node:22.22.3-alpine3.23
ARG NGINX_IMAGE=nginx:1.29.8-alpine3.23

FROM ${NODE_IMAGE} AS build
WORKDIR /workspace

RUN corepack enable && corepack install --global pnpm@11.1.3
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/admin/package.json apps/admin/package.json
COPY domains/demo/package.json domains/demo/package.json
COPY domains/jobs/package.json domains/jobs/package.json
COPY domains/monitor/package.json domains/monitor/package.json
COPY domains/system/package.json domains/system/package.json
COPY domains/tools/package.json domains/tools/package.json
COPY packages/api-client/package.json packages/api-client/package.json
COPY packages/app-core/package.json packages/app-core/package.json
COPY packages/config/package.json packages/config/package.json
COPY packages/contracts/package.json packages/contracts/package.json
COPY packages/runtime/package.json packages/runtime/package.json
COPY packages/ui/package.json packages/ui/package.json
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile

COPY . .
RUN pnpm build:prod

FROM ${NGINX_IMAGE} AS runtime
ARG VERSION=dev
ARG VCS_REF=unknown

RUN rm -f /etc/nginx/conf.d/default.conf && \
    install -d -o 101 -g 101 -m 0755 \
      /var/cache/nginx/client_temp \
      /var/cache/nginx/proxy_temp \
      /var/cache/nginx/fastcgi_temp \
      /var/cache/nginx/uwsgi_temp \
      /var/cache/nginx/scgi_temp \
      /var/run/nginx

COPY --from=build --chown=101:101 /workspace/dist /usr/share/nginx/html
COPY --chown=101:101 deploy/nginx/nginx.conf /etc/nginx/nginx.conf

LABEL org.opencontainers.image.title="go-admin-plus Admin Web" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.source="https://github.com/NAMEWTA/go-admin-plus"

USER 101:101
EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
