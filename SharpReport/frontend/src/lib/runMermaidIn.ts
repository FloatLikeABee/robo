let mermaidReady: Promise<typeof import('mermaid').default> | null = null;

const MERMAID_INIT = {
	startOnLoad: false,
	theme: 'dark',
	securityLevel: 'strict',
	suppressErrorRendering: true,
};

async function loadMermaid() {
	if (!mermaidReady) {
		mermaidReady = import('mermaid').then((mod) => {
			const mermaid = mod.default;
			mermaid.initialize(MERMAID_INIT);
			return mermaid;
		});
	}
	return mermaidReady;
}

export async function runMermaidIn(root: HTMLElement | null) {
	if (!root) return;
	const codes = root.querySelectorAll('code.language-mermaid');
	for (const el of codes) {
		const pre = document.createElement('pre');
		pre.className = 'mermaid';
		pre.textContent = el.textContent || '';
		el.parentElement?.replaceWith(pre);
	}
	const nodes = [...root.querySelectorAll('pre.mermaid, .mermaid')] as HTMLElement[];
	if (!nodes.length) return;
	const mermaid = await loadMermaid();
	for (const el of nodes) {
		const source = (el.textContent || '').trim();
		if (!source) continue;
		const id = `mmd-${Math.random().toString(36).slice(2, 10)}`;
		try {
			const { svg } = await mermaid.render(id, source);
			el.innerHTML = svg;
			el.classList.add('mermaid-rendered');
		} catch {
			el.textContent = source;
			el.classList.add('mermaid-fallback');
		}
	}
}
