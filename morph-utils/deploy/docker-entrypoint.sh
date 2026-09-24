#!/bin/sh
# Writes /config.js from public VITE_* env, then listens on PORT.
# Only ${PORT} is substituted. nginx $uri stays in the template.
set -eu

PORT="${PORT:-3040}"
case "$PORT" in
  ''|*[!0-9]*)
    echo "PORT must be a number" >&2
    exit 1
    ;;
esac

json_escape() {
  printf '%s' "$1" | tr -d '\r\n' | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g'
}

emit() {
  key=$1
  val=$2
  if [ -n "$val" ]; then
    printf '  %s: "%s",\n' "$key" "$(json_escape "$val")"
  fi
}

{
  printf '%s\n' 'window.__MORPH_UTILS_CONFIG__ = {'
  emit morphApiUrl "${VITE_MORPH_API_URL-}"
  emit usersPanelApiUrl "${VITE_USERS_PANEL_API_URL-}"
  emit sheetxUrl "${VITE_SHEETX_URL-}"
  emit formsxUrl "${VITE_FORMSX_URL-}"
  emit composerxUrl "${VITE_COMPOSERX_URL-}"
  emit dataxUrl "${VITE_DATAX_URL-}"
  emit projectsUrl "${VITE_PROJECTS_URL-}"
  emit morphEngiUrl "${VITE_MORPH_ENGI_URL-}"
  emit morphAiUrl "${VITE_MORPH_AI_URL-}"
  printf '%s\n' '};'
} > /usr/share/nginx/html/config.js

sed 's/\${PORT}/'"$PORT"'/g' /etc/nginx/templates/morph-utils.conf.template > /etc/nginx/conf.d/default.conf
exec nginx -g 'daemon off;'
