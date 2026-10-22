# DorphUtils shell. Build context is dorph-utils/.
# Public URL build-args only. No secret build args.
# A non-empty runtime env value is written to /config.js and wins over the build pin.

FROM node:22-bookworm-slim AS build
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
ARG VITE_DORPH_API_URL=
ARG VITE_USERS_PANEL_API_URL=
ARG VITE_SHEETX_URL=
ARG VITE_FORMSX_URL=
ARG VITE_COMPOSERX_URL=
ARG VITE_DATAX_URL=
ARG VITE_PROJECTS_URL=
ARG VITE_DORPH_ENGI_URL=
ARG VITE_DORPH_AI_URL=
ENV VITE_DORPH_API_URL=$VITE_DORPH_API_URL \
    VITE_USERS_PANEL_API_URL=$VITE_USERS_PANEL_API_URL \
    VITE_SHEETX_URL=$VITE_SHEETX_URL \
    VITE_FORMSX_URL=$VITE_FORMSX_URL \
    VITE_COMPOSERX_URL=$VITE_COMPOSERX_URL \
    VITE_DATAX_URL=$VITE_DATAX_URL \
    VITE_PROJECTS_URL=$VITE_PROJECTS_URL \
    VITE_DORPH_ENGI_URL=$VITE_DORPH_ENGI_URL \
    VITE_DORPH_AI_URL=$VITE_DORPH_AI_URL \
    CI=true
RUN npm run build

FROM nginx:1.27-alpine
COPY deploy/nginx.conf.template /etc/nginx/templates/dorph-utils.conf.template
COPY deploy/docker-entrypoint.sh /entrypoint.sh
COPY --from=build /src/dist /usr/share/nginx/html
RUN rm -f /etc/nginx/conf.d/default.conf \
  && chmod 755 /entrypoint.sh
ENV PORT=3040
EXPOSE 3040
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -fsS "http://127.0.0.1:${PORT}/health" || exit 1
ENTRYPOINT ["/entrypoint.sh"]
