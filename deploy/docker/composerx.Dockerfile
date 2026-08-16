# ComposerX API (Go)
FROM golang:1.24-bookworm AS builder
WORKDIR /src
COPY pkg/morphai /src/pkg/morphai
COPY pkg/assistmd /src/pkg/assistmd
COPY pkg/webresearch /src/pkg/webresearch
COPY composerx/backend/go.mod composerx/backend/go.sum /src/composerx/backend/
WORKDIR /src/composerx/backend
RUN go mod download
COPY composerx/backend/ /src/composerx/backend/
RUN CGO_ENABLED=0 go build -o /out/composerx-server .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /out/composerx-server /app/composerx-server
ENV PORT=8043
ENV TRAN_FILE_STORAGE_PATH=/storage
ENV TRAN_MONGO_DB=alterathena
RUN mkdir -p /storage
EXPOSE 8043
CMD ["sh", "-c", "exec /app/composerx-server"]
