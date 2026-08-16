const configured = import.meta.env.PUBLIC_API_URL as string | undefined;
const usersPanelConfigured = import.meta.env.PUBLIC_USERS_PANEL_URL as string | undefined;
const usersPanelViteConfigured = import.meta.env.VITE_USERS_PANEL_API_URL as string | undefined;
const AUTH_TOKEN_KEY = 'datax_auth_token';
const AUTH_REMEMBER_KEY = 'datax_auth_remember';
const LEGACY_AUTH_TOKEN_KEY = 'sharpreport_auth_token';
const LEGACY_AUTH_REMEMBER_KEY = 'sharpreport_auth_remember';
const SHARED_SESSION_COOKIE_KEY = 'userspanel_session_token';
/** Match UsersPanel default JWT expiry (see JWT_EXPIRY_HOURS). */
const SESSION_COOKIE_MAX_AGE_SECONDS = 48 * 3600;

/**
 * API origin without trailing slash. Empty means same-origin (relative `/api/...` paths).
 * In dev, defaults to same-origin so Vite proxies `/api` to the backend (avoids CORS).
 */
export function getApiOrigin(): string {
	if (configured !== undefined && configured !== '') {
		return configured.replace(/\/$/, '');
	}
	return '';
}

/** Full URL for `fetch()` — use for all `/api/v1/...` calls. */
export function apiUrl(path: string): string {
	const origin = getApiOrigin();
	const p = path.startsWith('/') ? path : `/${path}`;
	if (!origin) return p;
	return `${origin}${p}`;
}

/**
 * UsersPanel API origin. Empty = same-origin — Vite proxies `/api/messages` to UsersPanel (no CORS).
 */
export function usersPanelOrigin(): string {
	if (usersPanelViteConfigured !== undefined && usersPanelViteConfigured !== '') {
		return usersPanelViteConfigured.replace(/\/$/, '');
	}
	if (usersPanelConfigured !== undefined && usersPanelConfigured !== '') {
		return usersPanelConfigured.replace(/\/$/, '');
	}
	return '';
}

function readCookie(name: string): string {
	if (typeof document === 'undefined') return '';
	const prefix = `${name}=`;
	const row = document.cookie
		.split(';')
		.map((v) => v.trim())
		.find((v) => v.startsWith(prefix));
	if (!row) return '';
	return decodeURIComponent(row.slice(prefix.length));
}

function writeCookie(name: string, value: string, maxAgeSeconds: number): void {
	if (typeof document === 'undefined') return;
	document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`;
}

function writeSessionCookie(name: string, value: string): void {
	if (typeof document === 'undefined') return;
	document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; SameSite=Lax`;
}

export function isJwtExpired(token: string, skewMs = 60_000): boolean {
	const parts = token.split('.');
	if (parts.length !== 3) return false;
	try {
		let b64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
		const pad = b64.length % 4;
		if (pad) b64 += '='.repeat(4 - pad);
		const payload = JSON.parse(atob(b64)) as { exp?: number };
		if (typeof payload.exp !== 'number') return false;
		return Date.now() >= payload.exp * 1000 - skewMs;
	} catch {
		return false;
	}
}

function clearTokenStores(): void {
	if (typeof localStorage !== 'undefined') {
		localStorage.removeItem(AUTH_TOKEN_KEY);
		localStorage.removeItem(LEGACY_AUTH_TOKEN_KEY);
		localStorage.removeItem('authToken');
	}
	if (typeof sessionStorage !== 'undefined') {
		sessionStorage.removeItem(AUTH_TOKEN_KEY);
		sessionStorage.removeItem(LEGACY_AUTH_TOKEN_KEY);
	}
	writeCookie(SHARED_SESSION_COOKIE_KEY, '', 0);
}

function syncSharedCookie(token: string, remember: boolean): void {
	if (!token) return;
	if (remember) {
		writeCookie(SHARED_SESSION_COOKIE_KEY, token, SESSION_COOKIE_MAX_AGE_SECONDS);
	} else {
		writeSessionCookie(SHARED_SESSION_COOKIE_KEY, token);
	}
}

function acceptTokenOrClear(raw: string): string {
	if (!raw) return '';
	if (isJwtExpired(raw)) {
		clearTokenStores();
		return '';
	}
	return raw;
}

/** UsersPanel JWT shared across Morph / Tran* apps. */
export function getAuthToken(): string {
	let shared = acceptTokenOrClear(readCookie(SHARED_SESSION_COOKIE_KEY));
	if (shared) return shared;

	const remember =
		typeof localStorage !== 'undefined' &&
		(localStorage.getItem(AUTH_REMEMBER_KEY) ?? localStorage.getItem(LEGACY_AUTH_REMEMBER_KEY)) !==
			'0';
	const fromStore = acceptTokenOrClear(
		remember
			? (typeof localStorage !== 'undefined'
					? (localStorage.getItem(AUTH_TOKEN_KEY) ??
						localStorage.getItem(LEGACY_AUTH_TOKEN_KEY) ??
						'')
					: '')
			: (typeof sessionStorage !== 'undefined'
					? (sessionStorage.getItem(AUTH_TOKEN_KEY) ??
						sessionStorage.getItem(LEGACY_AUTH_TOKEN_KEY) ??
						'')
					: '')
	);
	if (!fromStore) return '';
	syncSharedCookie(fromStore, remember);
	return fromStore;
}

export function setAuthToken(token: string, options?: { rememberMe?: boolean }): void {
	if (!token) {
		clearTokenStores();
		return;
	}
	const remember = options?.rememberMe ?? true;
	if (typeof localStorage !== 'undefined') {
		localStorage.setItem(AUTH_REMEMBER_KEY, remember ? '1' : '0');
	}
	if (remember) {
		if (typeof localStorage !== 'undefined') localStorage.setItem(AUTH_TOKEN_KEY, token);
		if (typeof sessionStorage !== 'undefined') sessionStorage.removeItem(AUTH_TOKEN_KEY);
	} else {
		if (typeof sessionStorage !== 'undefined') sessionStorage.setItem(AUTH_TOKEN_KEY, token);
		if (typeof localStorage !== 'undefined') localStorage.removeItem(AUTH_TOKEN_KEY);
	}
	syncSharedCookie(token, remember);
}

export function authHeaders(): Record<string, string> {
	const token = getAuthToken();
	return token ? { Authorization: `Bearer ${token}` } : {};
}

function isLoginPost(path: string, method: string): boolean {
	return method.toUpperCase() === 'POST' && path === '/api/v1/auth/login';
}

/** Authenticated JSON request to the DataX API (proxies UsersPanel auth). */
export async function apiJson<T>(path: string, init: RequestInit = {}): Promise<T> {
	const method = (init.method ?? 'GET').toString().toUpperCase();
	const url = apiUrl(path);
	const sendAuth = !isLoginPost(path, method);

	let res: Response;
	try {
		res = await fetch(url, {
			...init,
			headers: {
				'Content-Type': 'application/json',
				...(sendAuth ? authHeaders() : {}),
				...(init.headers as Record<string, string>)
			}
		});
	} catch {
		const { backendStatus, backendLastError } = await import('$lib/backendHealth');
		backendStatus.set('offline');
		backendLastError.set(
			'DataX API is not reachable. The backend may have stopped after idle — use Retry in the banner or run `./start-all.sh restart sharpreport-api`.'
		);
		throw new Error(
			'DataX API is offline. Restart sharpreport-api or click Retry in the banner above.'
		);
	}

	const data = await res.json().catch(() => ({}));
	if (!res.ok) {
		if (res.status === 401 && sendAuth) clearTokenStores();
		if (res.status === 502 || res.status === 503 || res.status === 504) {
			const { backendStatus, backendLastError } = await import('$lib/backendHealth');
			backendStatus.set('offline');
			backendLastError.set((data as { error?: string }).error || res.statusText || 'API unavailable');
		}
		throw new Error((data as { error?: string }).error || res.statusText);
	}

	const { backendStatus } = await import('$lib/backendHealth');
	backendStatus.set('online');
	return data as T;
}

export async function usersPanelJson<T>(path: string, init: RequestInit = {}): Promise<T> {
	const token = getAuthToken();
	const origin = usersPanelOrigin();
	const url = origin ? `${origin}${path.startsWith('/') ? path : `/${path}`}` : path;
	const res = await fetch(url, {
		...init,
		headers: {
			'Content-Type': 'application/json',
			...(token ? { Authorization: `Bearer ${token}` } : {}),
			...(init.headers as Record<string, string>)
		}
	});
	const data = await res.json().catch(() => ({}));
	if (!res.ok) {
		throw new Error((data as { error?: string }).error || res.statusText);
	}
	return data as T;
}
