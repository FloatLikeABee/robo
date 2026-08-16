# Production env template for FormsX / Morph / ComposerX style apps.
# Real values: copy deploy/env.production.example → deploy/.env.production

# Multi-stage: UsersPanel Rust API
FROM rust:1.84-bookworm AS builder
WORKDIR /app
COPY UsersPanel/backend/Cargo.toml UsersPanel/backend/Cargo.lock ./
# Cache deps
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release && rm -rf src
COPY UsersPanel/backend/src ./src
COPY UsersPanel/backend/migrations ./migrations
RUN touch src/main.rs && cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/target/release/users-panel-api /app/users-panel-api
ENV HOST=0.0.0.0
ENV PORT=5001
EXPOSE 5001
CMD ["sh", "-c", "exec /app/users-panel-api"]
