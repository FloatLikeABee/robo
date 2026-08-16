# Stage 1: Build Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Backend
FROM rust:1.84-slim AS backend-builder
WORKDIR /app/backend
COPY backend/Cargo.* ./
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release
COPY backend/ ./
COPY --from=frontend-builder /app/frontend/build ./static
RUN cargo build --release

# Stage 3: Runtime
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y \
    openjdk-17-jre-headless \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy backend
COPY --from=backend-builder /app/backend/target/release/datapulse /app/datapulse

# Create metabase directory
RUN mkdir -p /app/metabase /app/data

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:3050/api/v1/health || exit 1

EXPOSE 3050

CMD ["/app/datapulse"]