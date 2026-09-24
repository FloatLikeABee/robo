# Project API + UI. Build context is the repo root (pkg/morphai-rs).
# No secret build args. Runtime secrets come from the environment only.
# morph-engi/Dockerfile.dockerignore filters this build. The root .dockerignore
# still excludes morph-engi from the Morph image.

FROM node:22-bookworm-slim AS ui
WORKDIR /src
COPY morph-engi/frontend/package.json morph-engi/frontend/package-lock.json morph-engi/frontend/
WORKDIR /src/morph-engi/frontend
RUN npm ci
COPY morph-engi/frontend/ ./
ENV CI=true \
    VITE_API_BASE_URL=
RUN npm run build

FROM rust:1.88-bookworm AS build
RUN apt-get update \
  && apt-get install -y --no-install-recommends pkg-config libsqlite3-dev \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY pkg/morphai-rs/ pkg/morphai-rs/
COPY morph-engi/backend/ morph-engi/backend/
WORKDIR /src/morph-engi/backend
RUN cargo build --release --locked

FROM debian:bookworm-slim
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates wget gosu libsqlite3-0 \
  && rm -rf /var/lib/apt/lists/* \
  && groupadd --gid 65532 morphengi \
  && useradd --uid 65532 --gid 65532 --no-create-home --shell /usr/sbin/nologin morphengi
WORKDIR /app
COPY --from=build /src/morph-engi/backend/target/release/morph-engi-api /app/morph-engi-api
COPY --from=ui /src/morph-engi/frontend/dist /app/frontend/dist
COPY morph-engi/deploy/docker-entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh /app/morph-engi-api
ENV PORT=9096 \
    APP_ENV=production \
    STATIC_DIR=/app/frontend/dist \
    MORPH_ENGI_DATABASE_URL=sqlite:///data/morph_engi.db \
    MORPH_ENGI_UPLOAD_DIR=/data/uploads
EXPOSE 9096
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD wget -q -O /dev/null "http://127.0.0.1:${PORT}/health" || exit 1
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/morph-engi-api"]
