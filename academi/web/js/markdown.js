/** HTML escaping and lightweight markdown rendering for chat/docs. */

export function escapeHtml(text) {
    const d = document.createElement('div');
    d.textContent = text;
    return d.innerHTML;
}


export function inlineMarkdown(text) {
    let out = text;
    out = out.replace(/`([^`]+)`/g, '<code class="md-inline-code">$1</code>');
    out = out.replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer" class="source-link">$1</a>');
    out = out.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    out = out.replace(/\*([^*]+)\*/g, '<em>$1</em>');
    return out;
}


export function markdownToHtml(text) {
    const escaped = escapeHtml(String(text || '')).replace(/\r\n/g, '\n');
    const codeBlocks = [];
    const withCodeTokens = escaped.replace(/```([a-zA-Z0-9_-]+)?\n?([\s\S]*?)```/g, (_, lang, code) => {
        const token = `@@CODE_BLOCK_${codeBlocks.length}@@`;
        const languageClass = lang ? ` lang-${lang.toLowerCase()}` : '';
        codeBlocks.push(`<pre class="md-code-block"><code class="${languageClass}">${code.trim()}</code></pre>`);
        return token;
    });

    const lines = withCodeTokens.split('\n');
    const html = [];
    const paragraph = [];
    let inUl = false;
    let inOl = false;

    const isTableLine = (line) => {
        const trimmed = line.trim();
        return trimmed.includes('|') && /^\|?.+\|.+\|?$/.test(trimmed);
    };
    const isTableSeparator = (line) => {
        const trimmed = line.trim();
        return /^\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)+\|?$/.test(trimmed);
    };
    const parseTableCells = (line) => {
        const trimmed = line.trim().replace(/^\|/, '').replace(/\|$/, '');
        return trimmed.split('|').map((cell) => inlineMarkdown(cell.trim()));
    };
    const emitTable = (headerLine, separatorLine, bodyLines) => {
        const headerCells = parseTableCells(headerLine);
        const separators = separatorLine
            .trim()
            .replace(/^\|/, '')
            .replace(/\|$/, '')
            .split('|')
            .map((s) => s.trim());
        const aligns = separators.map((s) => {
            const left = s.startsWith(':');
            const right = s.endsWith(':');
            if (left && right) return 'center';
            if (right) return 'right';
            return 'left';
        });
        const rows = bodyLines.map(parseTableCells);
        const headHtml = headerCells
            .map((cell, i) => `<th style="text-align:${aligns[i] || 'left'}">${cell}</th>`)
            .join('');
        const bodyHtml = rows
            .map((cells) => {
                const row = cells
                    .map((cell, i) => `<td style="text-align:${aligns[i] || 'left'}">${cell}</td>`)
                    .join('');
                return `<tr>${row}</tr>`;
            })
            .join('');
        return `<div class="md-table-wrap"><table class="md-table"><thead><tr>${headHtml}</tr></thead><tbody>${bodyHtml}</tbody></table></div>`;
    };

    const flushParagraph = () => {
        if (!paragraph.length) return;
        html.push(`<p>${paragraph.join('<br>')}</p>`);
        paragraph.length = 0;
    };
    const closeLists = () => {
        if (inUl) {
            html.push('</ul>');
            inUl = false;
        }
        if (inOl) {
            html.push('</ol>');
            inOl = false;
        }
    };

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trimEnd();
        const codeTokenMatch = line.match(/^@@CODE_BLOCK_(\d+)@@$/);
        if (codeTokenMatch) {
            flushParagraph();
            closeLists();
            html.push(codeBlocks[Number(codeTokenMatch[1])] || '');
            continue;
        }

        if (!line.trim()) {
            flushParagraph();
            closeLists();
            continue;
        }

        // GFM-style pipe table:
        // | h1 | h2 |
        // | --- | ---: |
        // | v1 | v2 |
        if (isTableLine(line)) {
            const next = lines[i + 1] || '';
            if (next && isTableSeparator(next)) {
                flushParagraph();
                closeLists();
                const body = [];
                let j = i + 2;
                while (j < lines.length && isTableLine(lines[j]) && !isTableSeparator(lines[j])) {
                    body.push(lines[j]);
                    j++;
                }
                html.push(emitTable(line, next, body));
                i = j - 1;
                continue;
            }
        }

        const heading = line.match(/^(#{1,6})\s+(.+)$/);
        if (heading) {
            flushParagraph();
            closeLists();
            const level = heading[1].length;
            html.push(`<h${level}>${inlineMarkdown(heading[2])}</h${level}>`);
            continue;
        }

        const ul = line.match(/^[-*]\s+(.+)$/);
        if (ul) {
            flushParagraph();
            if (inOl) {
                html.push('</ol>');
                inOl = false;
            }
            if (!inUl) {
                html.push('<ul>');
                inUl = true;
            }
            html.push(`<li>${inlineMarkdown(ul[1])}</li>`);
            continue;
        }

        const ol = line.match(/^\d+\.\s+(.+)$/);
        if (ol) {
            flushParagraph();
            if (inUl) {
                html.push('</ul>');
                inUl = false;
            }
            if (!inOl) {
                html.push('<ol>');
                inOl = true;
            }
            html.push(`<li>${inlineMarkdown(ol[1])}</li>`);
            continue;
        }

        paragraph.push(inlineMarkdown(line));
    }

    flushParagraph();
    closeLists();
    return html.join('');
}


export function renderLearnModalBody(el, text, mode) {
    if (!el) return;
    if (mode === 'loading') {
        el.classList.remove('doc-modal-body--md');
        el.removeAttribute('style');
        el.innerHTML = '';
        el.textContent = text || '';
        return;
    }
    if (mode === 'plain') {
        el.classList.remove('doc-modal-body--md');
        el.style.whiteSpace = 'pre-wrap';
        el.innerHTML = '';
        el.textContent = text || '';
        return;
    }
    el.removeAttribute('style');
    el.classList.add('doc-modal-body--md');
    const raw = text || '';
    if (typeof marked === 'undefined' || typeof DOMPurify === 'undefined') {
        el.textContent = raw;
        return;
    }
    const html = marked.parse(raw, { gfm: true, breaks: true });
    el.innerHTML = DOMPurify.sanitize(html);
}
