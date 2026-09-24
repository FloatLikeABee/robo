# Morph API + Morph AI UI. Build context is the repo root (pkg/ replace directives).
# No secret build args. Runtime secrets come from the environment only.

FROM node:22-bookworm-slim AS ui
WORKDIR /src
COPY scripts/with-root-env.cjs scripts/with-root-env.cjs
COPY morph/frontend/package.json morph/frontend/package-lock.json morph/frontend/
WORKDIR /src/morph/frontend
RUN npm ci
COPY morph/frontend/ ./
ENV CI=true
# Public MorphUtils origin, inlined by CRA. Render passes service env vars as
# build args. Empty or loopback omits the header link (see headerAppLinks.js).
ARG REACT_APP_MORPH_UTILS_URL
ENV REACT_APP_MORPH_UTILS_URL=${REACT_APP_MORPH_UTILS_URL}
RUN npm run build

FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY pkg/ pkg/
COPY morph/go.mod morph/go.sum morph/
WORKDIR /src/morph
RUN go mod download
COPY morph/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/morph-server .

FROM alpine:3.22
RUN apk add --no-cache ca-certificates su-exec \
  && addgroup -g 65532 morph \
  && adduser -D -H -u 65532 -G morph morph
WORKDIR /app
COPY --from=build /out/morph-server /app/morph-server
COPY --from=ui /src/morph/frontend/build /app/frontend/build
COPY deploy/docker-entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh /app/morph-server
ENV PORT=9090 \
    DB_PATH=/data/badger \
    TRAN_SQLITE_PATH=/data/tran.sqlite \
    ENTITY_DETAILS_BADGER=/data/entity_details \
    MORPH_KNOWLEDGE_DIR=/data/knowledge \
    TRAN_ENTITY_ATTACHMENT_DIR=/data/uploads/entity_attachments \
    GIN_MODE=release
EXPOSE 9090
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD wget -q -O /dev/null "http://127.0.0.1:${PORT}/health" || exit 1
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/morph-server"]
