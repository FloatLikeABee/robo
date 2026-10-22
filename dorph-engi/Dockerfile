# Project API + UI. Build context is the repo root (pkg/dorphai-rs).
# No secret build args. Runtime secrets come from the environment only.
# dorph-engi/Dockerfile.dockerignore filters this build. The root .dockerignore
# still excludes dorph-engi from the Dorph image.

FROM node:22-bookworm-slim AS ui
WORKDIR /src
COPY dorph-engi/frontend/package.json dorph-engi/frontend/package-lock.json dorph-engi/frontend/
WORKDIR /src/dorph-engi/frontend
RUN npm ci
COPY dorph-engi/frontend/ ./
ENV CI=true \
    VITE_API_BASE_URL=
RUN npm run build

FROM rust:1.88-bookworm AS build
RUN apt-get update \
  && apt-get install -y --no-install-recommends pkg-config libsqlite3-dev \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY pkg/dorphai-rs/ pkg/dorphai-rs/
COPY dorph-engi/backend/ dorph-engi/backend/
WORKDIR /src/dorph-engi/backend
RUN cargo build --release --locked

FROM debian:bookworm-slim
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates wget gosu libsqlite3-0 \
  && rm -rf /var/lib/apt/lists/* \
  && groupadd --gid 65532 dorphengi \
  && useradd --uid 65532 --gid 65532 --no-create-home --shell /usr/sbin/nologin dorphengi
WORKDIR /app
COPY --from=build /src/dorph-engi/backend/target/release/dorph-engi-api /app/dorph-engi-api
COPY --from=ui /src/dorph-engi/frontend/dist /app/frontend/dist
COPY dorph-engi/deploy/docker-entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh /app/dorph-engi-api
ENV PORT=9096 \
    APP_ENV=production \
    STATIC_DIR=/app/frontend/dist \
    DORPH_ENGI_DATABASE_URL=sqlite:///data/dorph_engi.db \
    DORPH_ENGI_UPLOAD_DIR=/data/uploads
EXPOSE 9096
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD wget -q -O /dev/null "http://127.0.0.1:${PORT}/health" || exit 1
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/dorph-engi-api"]
