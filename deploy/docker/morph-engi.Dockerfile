# Morph Engi API (Rust)
FROM rust:1.84-bookworm AS builder
WORKDIR /app
COPY morph-engi/backend/Cargo.toml morph-engi/backend/Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release && rm -rf src
COPY morph-engi/backend/src ./src
RUN touch src/main.rs && cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/target/release/morph-engi-api /app/morph-engi-api
ENV HOST=0.0.0.0
ENV APP_PORT=9096
ENV PORT=9096
EXPOSE 9096
CMD ["sh", "-c", "if [ -n \"$PORT\" ]; then export APP_PORT=\"$PORT\"; fi; exec /app/morph-engi-api"]
