# Content Maker API + UI. Build context is the repo root (pkg/ replace directives).
# No secret build args. Runtime secrets come from the environment only.

FROM node:22-bookworm-slim AS ui
WORKDIR /src
COPY composerx/frontend/package.json composerx/frontend/package-lock.json composerx/frontend/
WORKDIR /src/composerx/frontend
RUN npm ci
COPY composerx/frontend/ ./
ENV CI=true
# Empty keeps same-origin API calls. Unset would bake http://localhost:8043.
ARG VITE_API_BASE=
ENV VITE_API_BASE=${VITE_API_BASE}
RUN npm run build

FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY pkg/ pkg/
COPY composerx/backend/go.mod composerx/backend/go.sum composerx/backend/
WORKDIR /src/composerx/backend
RUN go mod download
COPY composerx/backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/composerx-server .

FROM alpine:3.22
RUN apk add --no-cache ca-certificates su-exec \
  && addgroup -g 65532 composerx \
  && adduser -D -H -u 65532 -G composerx composerx
WORKDIR /app
COPY --from=build /out/composerx-server /app/composerx-server
COPY --from=ui /src/composerx/frontend/dist /app/frontend/dist
COPY composerx/deploy/docker-entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh /app/composerx-server
ENV PORT=8043 \
    COMPOSERX_PORT=8043 \
    COMPOSERX_SQLITE_PATH=/data/composerx.sqlite \
    COMPOSERX_BADGER_PATH=/data/composerx_badger \
    TRAN_FILE_STORAGE_PATH=/data/storage \
    COMPOSERX_UI_DIR=/app/frontend/dist \
    GIN_MODE=release
EXPOSE 8043
HEALTHCHECK --interval=30s --timeout=3s --start-period=20s --retries=3 \
  CMD wget -q -O /dev/null "http://127.0.0.1:${PORT}/health" || exit 1
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/composerx-server"]
