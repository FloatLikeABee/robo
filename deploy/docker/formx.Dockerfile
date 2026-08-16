# FormsX API (Go) — monorepo context for pkg/* replace
FROM golang:1.22-bookworm AS builder
WORKDIR /src
COPY pkg/morphai /src/pkg/morphai
COPY pkg/assistmd /src/pkg/assistmd
COPY pkg/webresearch /src/pkg/webresearch
COPY formx/backend/go.mod formx/backend/go.sum /src/formx/backend/
WORKDIR /src/formx/backend
RUN go mod download
COPY formx/backend/ /src/formx/backend/
RUN CGO_ENABLED=0 go build -o /out/formsx-server ./cmd/server

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /out/formsx-server /app/formsx-server
ENV SERVER_PORT=29909
ENV PORT=29909
ENV UPLOAD_DIR=/uploads
RUN mkdir -p /uploads
EXPOSE 29909
CMD ["sh", "-c", "if [ -n \"$PORT\" ]; then export SERVER_PORT=\"$PORT\"; fi; exec /app/formsx-server"]
