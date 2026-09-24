# Event Logs API + UI. Build context is the repo root (pkg/ replace directives).
# No secret build args. Runtime secrets come from the environment only.
# formx/Dockerfile.dockerignore filters this build. The root .dockerignore
# still excludes formx from the Morph image.

FROM node:22-bookworm-slim AS ui
WORKDIR /src
COPY formx/frontend/package.json formx/frontend/package-lock.json formx/frontend/
WORKDIR /src/formx/frontend
RUN npm ci
COPY formx/frontend/ ./
ENV CI=true \
    VITE_API_URL=
RUN npm run build

FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY pkg/ pkg/
COPY formx/backend/go.mod formx/backend/go.sum formx/backend/
WORKDIR /src/formx/backend
RUN go mod download
COPY formx/backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/formx-server ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates su-exec \
  && addgroup -g 65532 formx \
  && adduser -D -H -u 65532 -G formx formx
WORKDIR /app
COPY --from=build /out/formx-server /app/formx-server
COPY --from=ui /src/formx/frontend/dist /app/frontend/dist
COPY formx/deploy/docker-entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh /app/formx-server
ENV PORT=29909 \
    SERVER_PORT=29909 \
    FORMSX_SQLITE_PATH=/data/formsx.sqlite \
    FORMSX_BADGER_PATH=/data/formsx_badger \
    UPLOAD_DIR=/data/uploads \
    GIN_MODE=release
EXPOSE 29909
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD wget -q -O /dev/null "http://127.0.0.1:${PORT}/health" || exit 1
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/formx-server"]
