# MorphData — React build + Go API (serves SPA)
FROM node:20-bookworm AS frontend
WORKDIR /fe
COPY morph/frontend/package*.json ./
RUN npm ci
COPY morph/frontend/ ./
RUN npm run build

FROM golang:1.22-bookworm AS backend
WORKDIR /src
COPY pkg/morphai /src/pkg/morphai
COPY morph/go.mod morph/go.sum /src/morph/
WORKDIR /src/morph
RUN go mod download
COPY morph/ /src/morph/
COPY --from=frontend /fe/build /src/morph/frontend/build
RUN CGO_ENABLED=1 go build -o /out/morph-server main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=backend /out/morph-server /app/morph-server
COPY --from=backend /src/morph/frontend/build /app/frontend/build
ENV PORT=9090
ENV DB_PATH=/data/badger
RUN mkdir -p /data/badger
EXPOSE 9090
CMD ["sh", "-c", "exec /app/morph-server"]
