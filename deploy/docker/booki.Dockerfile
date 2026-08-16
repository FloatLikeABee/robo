# Booki API (Go)
FROM golang:1.24-bookworm AS builder
WORKDIR /src
COPY pkg/morphai /src/pkg/morphai
COPY pkg/assistmd /src/pkg/assistmd
COPY booki/backend/go.mod booki/backend/go.sum /src/booki/backend/
WORKDIR /src/booki/backend
RUN go mod download
COPY booki/backend/ /src/booki/backend/
RUN CGO_ENABLED=0 go build -o /out/booki-server ./cmd/server

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /out/booki-server /app/booki-server
ENV APP_PORT=9095
ENV PORT=9095
EXPOSE 9095
CMD ["sh", "-c", "if [ -n \"$PORT\" ]; then export APP_PORT=\"$PORT\"; fi; exec /app/booki-server"]
