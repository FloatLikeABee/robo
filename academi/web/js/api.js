/** Backend /api/v1 base. On real devices, localhost is wrong — use page hostname + academi-api-port, or set meta/localStorage. */
export function resolveAcademiApiBaseUrl() {
    try {
        const stored = localStorage.getItem('academi_api_base');
        if (stored && /^https?:\/\//i.test(stored.trim())) {
            return stored.trim().replace(/\/+$/, '');
        }
    } catch (_) {
        /* private mode */
    }

    const meta = document.querySelector('meta[name="academi-api-base"]');
    const metaBase = meta?.getAttribute('content')?.trim();
    if (metaBase && /^https?:\/\//i.test(metaBase)) {
        return metaBase.replace(/\/+$/, '');
    }

    let apiPort = '8978';
    const portMeta = document.querySelector('meta[name="academi-api-port"]');
    if (portMeta?.getAttribute('content')?.trim()) {
        apiPort = portMeta.getAttribute('content').trim();
    }

    const { protocol, hostname } = window.location;
    const isLocal =
        !hostname ||
        hostname === 'localhost' ||
        hostname === '127.0.0.1' ||
        hostname === '[::1]';

    if (isLocal) {
        return `http://localhost:${apiPort}/api/v1`;
    }

    const scheme = protocol === 'https:' ? 'https' : 'http';
    return `${scheme}://${hostname}:${apiPort}/api/v1`;
}

export function debounce(fn, ms) {
    let timer = null;
    return (...args) => {
        clearTimeout(timer);
        timer = setTimeout(() => fn(...args), ms);
    };
}

export async function readApiErrorResponse(res) {
    const text = await res.text();
    try {
        const j = JSON.parse(text);
        if (j && typeof j.error === 'string' && j.error) return j.error;
    } catch (_) {
        /* not JSON */
    }
    const hint = text && text.trim() ? text.trim().slice(0, 180) : '';
    return hint || `Request failed (${res.status})`;
}

export function authBearerHeaders(authToken, extra = {}) {
    const h = { ...extra };
    if (authToken) h['Authorization'] = `Bearer ${authToken}`;
    return h;
}

export function sessionHeaders(authToken) {
    const h = { Accept: 'application/json', 'Content-Type': 'application/json' };
    if (authToken) h['Authorization'] = `Bearer ${authToken}`;
    return h;
}
