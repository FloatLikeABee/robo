// See https://kit.svelte.dev/docs/types#app

declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}

	interface ImportMetaEnv {
		readonly PUBLIC_API_URL?: string;
		readonly PUBLIC_MORPH_AI_URL?: string;
		readonly VITE_MORPH_AI_URL?: string;
	}
}

export {};
