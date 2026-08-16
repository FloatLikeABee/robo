import { browser } from '$app/environment';
import { get, writable } from 'svelte/store';
import { goto } from '$app/navigation';
import { apiJson, getAuthToken, setAuthToken } from '$lib/api';

/** Navigate to login; safe when embedded in a cross-origin parent or after a failed iframe load. */
function redirectToLogin() {
	if (!browser) return;
	if (window.location.pathname.startsWith('/login')) return;

	const loginPath = '/login';
	if (window.self !== window.top) {
		try {
			if (window.top!.location.origin === window.location.origin) {
				window.top!.location.href = `${window.location.origin}${loginPath}`;
				return;
			}
		} catch {
			/* cross-origin parent — navigate this frame only */
		}
		window.location.assign(`${window.location.origin}${loginPath}`);
		return;
	}
	void goto(loginPath);
}

export interface User {
	id: string;
	email: string;
	name: string;
	role: string;
	avatarUrl?: string;
	roles?: string[];
}

export interface AuthState {
	isAuthenticated: boolean;
	user: User | null;
	loading: boolean;
	token?: string;
}

export const auth = writable<AuthState>({
	isAuthenticated: false,
	user: null,
	loading: true
});

/** After session restore finishes, run callback or send user to login. Unsubscribe on cleanup. */
export function whenSessionReady(run: () => void): () => void {
	const snap = get(auth);
	if (!snap.loading) {
		if (!snap.isAuthenticated) redirectToLogin();
		else run();
		return () => {};
	}
	const unsub = auth.subscribe((s) => {
		if (s.loading) return;
		unsub();
		if (!s.isAuthenticated) redirectToLogin();
		else run();
	});
	return () => unsub();
}

type MeResponse = {
	id: string;
	email: string;
	name: string;
	role: string;
	avatar_url?: string;
	roles?: string[];
};

function mapUser(raw: MeResponse): User {
	return {
		id: raw.id,
		email: raw.email,
		name: raw.name,
		role: raw.role,
		avatarUrl: raw.avatar_url,
		roles: raw.roles
	};
}

export async function login(
	email: string,
	password: string,
	options?: { rememberMe?: boolean }
): Promise<AuthState> {
	try {
		const data = await apiJson<{
			token: string;
			user: MeResponse;
		}>('/api/v1/auth/login', {
			method: 'POST',
			body: JSON.stringify({ email, password })
		});

		setAuthToken(data.token, { rememberMe: options?.rememberMe ?? true });

		const authState: AuthState = {
			isAuthenticated: true,
			user: mapUser(data.user),
			loading: false,
			token: data.token
		};

		auth.set(authState);
		return authState;
	} catch (error) {
		console.error('Login error:', error);
		setAuthToken('');
		auth.set({
			isAuthenticated: false,
			user: null,
			loading: false
		});
		throw error;
	}
}

export async function logout(): Promise<void> {
	try {
		const token = getAuthToken();
		if (token) {
			await apiJson('/api/v1/auth/logout', {
				method: 'POST',
				headers: { Authorization: `Bearer ${token}` }
			});
		}
	} finally {
		setAuthToken('');
		auth.set({
			isAuthenticated: false,
			user: null,
			loading: false
		});
	}
}

export async function checkAuth(): Promise<AuthState> {
	const token = getAuthToken();

	if (!token) {
		return {
			isAuthenticated: false,
			user: null,
			loading: false
		};
	}

	try {
		const user = await apiJson<MeResponse>('/api/v1/auth/me');

		return {
			isAuthenticated: true,
			user: mapUser(user),
			loading: false,
			token
		};
	} catch (error) {
		console.error('Auth check error:', error);
		setAuthToken('');
		return {
			isAuthenticated: false,
			user: null,
			loading: false
		};
	}
}
