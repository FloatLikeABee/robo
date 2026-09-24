# Data Access API + UI. Build context is the repo root (pkg/morphai-rs).
# No secret build args. Runtime secrets come from the environment only.

FROM node:22-bookworm-slim AS ui
WORKDIR /src
COPY SharpReport/frontend/package.json SharpReport/frontend/package-lock.json SharpReport/frontend/.npmrc SharpReport/frontend/
WORKDIR /src/SharpReport/frontend
RUN npm ci
COPY SharpReport/frontend/ ./
ENV CI=true
RUN npm run build

FROM rust:1.89-bookworm AS build
WORKDIR /src
COPY pkg/morphai-rs pkg/morphai-rs
COPY SharpReport/backend/Cargo.toml SharpReport/backend/Cargo.lock SharpReport/backend/
WORKDIR /src/SharpReport/backend
COPY SharpReport/backend/ ./
RUN cargo build --release --locked

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates libsqlite3-0 libssl3 wget \
  && rm -rf /var/lib/apt/lists/* \
  && groupadd --gid 65532 sharpreport \
  && useradd --uid 65532 --gid 65532 --no-create-home --shell /usr/sbin/nologin sharpreport
WORKDIR /app
COPY --from=build /src/SharpReport/backend/target/release/datapulse /app/datapulse
COPY --from=ui /src/SharpReport/frontend/build /app/ui
COPY SharpReport/backend/config /app/config
COPY SharpReport/backend/migrations /app/migrations
COPY SharpReport/deploy/docker-entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh /app/datapulse
ENV PORT=3050 \
    SHARPREPORT_PORT=3050 \
    SHARPREPORT_DATABASE_URL=sqlite:///data/datapulse.db \
    SHARPREPORT_UI_DIR=/app/ui \
    RUST_LOG=info
EXPOSE 3050
HEALTHCHECK --interval=30s --timeout=3s --start-period=20s --retries=3 \
  CMD wget -q -O /dev/null "http://127.0.0.1:${PORT}/health" || exit 1
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/datapulse"]
