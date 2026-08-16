# SharpReport / DataPulse — thin wrapper pointing at existing multi-stage Dockerfile.
# Prefer SharpReport/deploy/docker-compose.yml on Alibaba ECS (needs Java/Metabase RAM).
FROM debian:bookworm-slim AS hint
RUN echo "Use SharpReport/deploy/Dockerfile with context SharpReport/ for full builds." > /README.txt
# Re-export the project Dockerfile by building from monorepo:
#   docker build -f SharpReport/deploy/Dockerfile -t robo/sharpreport SharpReport

# Actual usable image (context must still be SharpReport/ for that Dockerfile).
# This monorepo Dockerfile builds backend only; Metabase is expected via Compose.
FROM rust:1.84-slim AS backend-builder
WORKDIR /app/backend
COPY SharpReport/backend/Cargo.toml SharpReport/backend/Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release && rm -rf src
COPY SharpReport/backend/ ./
RUN touch src/main.rs && cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    openjdk-17-jre-headless ca-certificates curl \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend-builder /app/backend/target/release/datapulse /app/datapulse
RUN mkdir -p /app/metabase /app/data
ENV PORT=3050
EXPOSE 3050
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://127.0.0.1:${PORT}/api/v1/health || exit 1
CMD ["sh", "-c", "exec /app/datapulse"]
