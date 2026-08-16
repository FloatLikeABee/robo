import { readApiErrorResponse, debounce } from './api.js';
import { escapeHtml, markdownToHtml, renderLearnModalBody } from './markdown.js';

/** @typedef {import('./app.js').AcademiApp} AcademiApp */



/** Methods mixed onto AcademiApp.prototype — `this` is the app instance. */
export const docsMethods = {
isDocLearnable(doc) {
    const t = (doc.type || '').toLowerCase();
    return ['markdown', 'text', 'pdf', 'image'].includes(t);
}
,

openCreateDocModal() {
    document.getElementById('createDocTitle').value = '';
    document.getElementById('createDocContent').value = '';
    const f = document.getElementById('createDocFile');
    if (f) f.value = '';
    this.updateCreateDocFileName();
    document.getElementById('createDocModal').hidden = false;
}
,

closeCreateDocModal() {
    document.getElementById('createDocModal').hidden = true;
    this.updateCreateDocFileName();
}
,

updateCreateDocFileName() {
    const fileEl = document.getElementById('createDocFile');
    const nameEl = document.getElementById('createDocFileName');
    if (!nameEl) return;
    const fileName = fileEl?.files?.[0]?.name;
    nameEl.textContent = fileName ? `Selected: ${fileName}` : 'No file selected';
}
,

async submitCreateDoc(runLearnAfter) {
    const title = document.getElementById('createDocTitle').value.trim();
    const content = document.getElementById('createDocContent').value.trim();
    const fileEl = document.getElementById('createDocFile');
    const file = fileEl?.files?.[0];

    if (file) {
        const fd = new FormData();
        fd.append('file', file);
        if (title) fd.append('title', title);
        try {
            const res = await fetch(`${this.apiBaseUrl}/docs/upload`, {
                method: 'POST',
                headers: this.authBearerHeaders(),
                body: fd,
            });
            if (!res.ok) throw new Error(await readApiErrorResponse(res));
            const doc = await res.json();
            this.closeCreateDocModal();
            if (fileEl) fileEl.value = '';
            this.updateCreateDocFileName();
            await this.loadDocs(true);
            this.flashStatus(`Uploaded: ${doc.title || 'Document'}`);
            if (runLearnAfter && doc.id) {
                await this.runLearnForDoc(doc.id);
            }
        } catch (e) {
            console.error(e);
            alert(e.message || 'Upload failed');
        }
        return;
    }

    if (!title || !content) {
        alert('Add a file, or fill in both title and content.');
        return;
    }
    const headers = this.authBearerHeaders({
        'Content-Type': 'application/json',
        Accept: 'application/json',
    });
    try {
        const res = await fetch(`${this.apiBaseUrl}/docs`, {
            method: 'POST',
            headers,
            body: JSON.stringify({
                title,
                type: 'markdown',
                content,
                tags: ['#created', '#notes'],
            }),
        });
        if (!res.ok) throw new Error(await readApiErrorResponse(res));
        const doc = await res.json();
        this.closeCreateDocModal();
        await this.loadDocs(true);
        this.flashStatus(`Saved: ${doc.title}`);
        if (runLearnAfter && doc.id) {
            await this.runLearnForDoc(doc.id);
        }
    } catch (e) {
        console.error(e);
        alert(e.message || 'Could not save document');
    }
}
,

requestLearnNotificationFromGesture() {
    try {
        if (typeof Notification === 'undefined') return;
        if (Notification.permission === 'default') {
            void Notification.requestPermission();
        }
    } catch (_) {
        /* ignore */
    }
}
,

setLearnModalAnalyzingUI(analyzing) {
    const bgBtn = document.getElementById('learnModalBackground');
    const saveBtn = document.getElementById('learnModalSaveDoc');
    if (bgBtn) bgBtn.hidden = !analyzing;
    if (saveBtn) {
        const showSave = !analyzing && this._learnLastAnalysisOk !== false;
        saveBtn.hidden = !showSave;
        saveBtn.disabled = analyzing || !showSave;
    }
}
,

reopenLearnResultsModal() {
    const learnModal = document.getElementById('learnModal');
    const bodyEl = document.getElementById('learnModalBody');
    const heading = document.getElementById('learnModalHeading');
    if (heading) heading.textContent = this._learnAnalysisTitle || 'Help you learn';
    const bodyMode = this._learnLastAnalysisOk !== false ? 'markdown' : 'plain';
    renderLearnModalBody(bodyEl, this._learnAnalysisText || '', bodyMode);
    this.setLearnModalAnalyzingUI(false);
    if (learnModal) learnModal.hidden = false;
}
,

maybeNotifyLearnFinished(ok, titleLine, errMsg) {
    const docLabel = titleLine || 'Document';
    const canNotify = typeof Notification !== 'undefined' && Notification.permission === 'granted';
    if (canNotify) {
        try {
            const n = new Notification(ok ? 'Help you learn — ready' : 'Help you learn — error', {
                body: ok ? `Analysis finished for “${docLabel}”. Tap to view.` : (errMsg || 'Analysis failed.'),
                tag: 'academi-learn-finished',
            });
            n.onclick = () => {
                try {
                    n.close();
                } catch (_) {
                    /* ignore */
                }
                window.focus();
                this.reopenLearnResultsModal();
            };
        } catch (e) {
            console.warn('Notification failed', e);
            this.flashStatus(
                ok ? `Learning analysis ready — open Help you learn for “${docLabel}”.` : (errMsg || 'Analysis failed.'),
            );
        }
        return;
    }
    this.flashStatus(
        ok
            ? `Learning analysis ready for “${docLabel}” — open Docs and tap Help you learn to view.`
            : errMsg || 'Analysis failed.',
    );
}
,

async runLearnForDoc(docId) {
    const doc = this.docsList.find((d) => d.id === docId);
    const jobId = ++this._learnJobSeq;
    this._learnFetchPending = true;
    this._learnWantNotifyForJobId = null;

    const headers = { 'Content-Type': 'application/json', Accept: 'application/json' };
    if (this.authToken) headers['Authorization'] = `Bearer ${this.authToken}`;
    const learnModal = document.getElementById('learnModal');
    const bodyEl = document.getElementById('learnModalBody');
    const heading = document.getElementById('learnModalHeading');
    const headingDocTitle = doc ? doc.title : 'Document';
    if (heading) heading.textContent = doc ? `Help you learn · ${doc.title}` : 'Help you learn';
    renderLearnModalBody(bodyEl, 'Analyzing…', 'loading');
    this.setLearnModalAnalyzingUI(true);
    if (learnModal) learnModal.hidden = false;

    const finish = (ok, textOrErr) => {
        if (jobId !== this._learnJobSeq) return;

        this._learnFetchPending = false;
        this._learnLastAnalysisOk = ok;
        const modalHidden = !learnModal || learnModal.hidden;
        const tabHidden = document.visibilityState === 'hidden';
        const dismissedThisJob = this._learnWantNotifyForJobId === jobId;
        const shouldNotify = tabHidden || dismissedThisJob || modalHidden;

        if (ok) {
            this._learnAnalysisText = textOrErr || '';
            this._learnAnalysisTitle = `Learning: ${headingDocTitle}`;
            renderLearnModalBody(bodyEl, this._learnAnalysisText, 'markdown');
        } else {
            const err = textOrErr || 'Error';
            this._learnAnalysisText = err;
            this._learnAnalysisTitle = `Help you learn · ${headingDocTitle}`;
            renderLearnModalBody(bodyEl, err, 'plain');
        }

        this.setLearnModalAnalyzingUI(false);
        if (this._learnWantNotifyForJobId === jobId) {
            this._learnWantNotifyForJobId = null;
        }

        if (shouldNotify) {
            this.maybeNotifyLearnFinished(ok, headingDocTitle, ok ? null : String(textOrErr || 'Error'));
        }
    };

    try {
        const res = await fetch(`${this.apiBaseUrl}/ai/learn`, {
            method: 'POST',
            headers,
            body: JSON.stringify({
                doc_id: docId,
                disable_research: !this.researchEnabled,
            }),
        });
        if (jobId !== this._learnJobSeq) return;

        if (!res.ok) {
            const err = await res.json().catch(() => ({}));
            throw new Error(err.error || 'Analysis failed');
        }
        const data = await res.json();
        finish(true, data.response || '');
    } catch (e) {
        console.error(e);
        if (jobId !== this._learnJobSeq) return;
        finish(false, e.message || 'Error');
    }
}
,

closeLearnModal() {
    if (this._learnFetchPending) {
        const id = this._learnJobSeq;
        if (id > 0) {
            this._learnWantNotifyForJobId = id;
            this.requestLearnNotificationFromGesture();
        }
    }
    document.getElementById('learnModal').hidden = true;
}
,

async saveLearnAnalysisToDocs() {
    const content = this._learnAnalysisText;
    if (!content || !content.trim()) return;
    const headers = { 'Content-Type': 'application/json', Accept: 'application/json' };
    if (this.authToken) headers['Authorization'] = `Bearer ${this.authToken}`;
    try {
        const res = await fetch(`${this.apiBaseUrl}/docs`, {
            method: 'POST',
            headers,
            body: JSON.stringify({
                title: this._learnAnalysisTitle || 'Learning analysis',
                type: 'markdown',
                content,
                tags: ['#analysis', '#help-you-learn'],
            }),
        });
        if (!res.ok) throw new Error(await readApiErrorResponse(res));
        this.closeLearnModal();
        await this.loadDocs(true);
        this.flashStatus('Analysis saved to Docs');
    } catch (e) {
        console.error(e);
        alert(e.message || 'Could not save analysis');
    }
}
,

async loadDocs(force = false) {
    const grid = document.getElementById('docsGrid');
    if (!grid) return;

    const now = Date.now();
    if (!force && this.docsList.length && now - this._docsLoadedAt < this._screenCacheMs) {
        const q = document.getElementById('docSearch')?.value?.trim() || '';
        this.renderDocsGrid(q);
        return;
    }
    if (this._docsLoadPromise) return this._docsLoadPromise;

    this._docsLoadPromise = (async () => {
        try {
            const headers = { Accept: 'application/json' };
            if (this.authToken) headers['Authorization'] = `Bearer ${this.authToken}`;
            const res = await fetch(`${this.apiBaseUrl}/docs?brief=1`, { headers });
            if (!res.ok) throw new Error('Failed to load docs');
            this.docsList = await res.json();
            this._docsLoadedAt = Date.now();
            const q = document.getElementById('docSearch')?.value?.trim() || '';
            this.renderDocsGrid(q);
        } catch (e) {
            console.error(e);
            this.docsList = [];
            this.renderDocsGrid('');
        }
    })().finally(() => {
        this._docsLoadPromise = null;
    });

    return this._docsLoadPromise;
}
,

renderDocsGrid(query) {
    const grid = document.getElementById('docsGrid');
    const emptyEl = document.getElementById('docsEmpty');
    if (!grid) return;
    grid.querySelectorAll('.doc-card').forEach((n) => n.remove());

    const q = (query || '').toLowerCase();
    let list = this.docsList;
    if (q) {
        list = this.docsList.filter((d) => {
            const blob = `${d.title || ''} ${d.ai_summary || ''} ${d.content || ''} ${(d.tags || []).join(' ')}`.toLowerCase();
            return blob.includes(q);
        });
    }

    if (!list.length) {
        if (emptyEl) emptyEl.hidden = false;
        return;
    }
    if (emptyEl) emptyEl.hidden = true;

    const sorted = [...list].sort((a, b) => (b.created_at || 0) - (a.created_at || 0));
    for (const doc of sorted) {
        const card = document.createElement('div');
        card.className = 'doc-card glass';
        card.setAttribute('role', 'button');
        card.dataset.docId = doc.id;
        const tags = (doc.tags || []).slice(0, 6).map((t) => `<span class="tag-chip">${this.escapeHtml(t)}</span>`).join('');
        const rawSum = doc.ai_summary || (doc.content ? String(doc.content).slice(0, 220) : '') || '';
        const summary = this.escapeHtml(rawSum);
        const truncated = (doc.content && doc.content.length > 220) || (doc.ai_summary && doc.ai_summary.length > 220);
        const title = this.escapeHtml(doc.title || 'Untitled');
        const learnBtn = this.isDocLearnable(doc)
            ? `<button type="button" class="doc-learn-btn" data-doc-id="${this.escapeHtml(doc.id)}">Help you learn</button>`
            : '';
        const publishBtn = `<button type="button" class="doc-publish-btn" data-doc-id="${this.escapeHtml(doc.id)}">Share to Community</button>`;
        card.innerHTML = `
            <div class="doc-card-head">
                <div class="doc-preview"><div class="doc-icon">⧉</div></div>
                <div class="doc-info">
                    <h4>${title}</h4>
                    <p>${summary}${truncated ? '…' : ''}</p>
                </div>
            </div>
            <div class="doc-tags truncated">${tags}</div>
            <div class="doc-card-actions">${publishBtn}${learnBtn}</div>
        `;
        grid.appendChild(card);
    }
}
,

mergeDocIntoList(doc) {
    if (!doc || !doc.id) return;
    const id = String(doc.id);
    const i = this.docsList.findIndex((d) => String(d.id) === id);
    if (i >= 0) {
        this.docsList[i] = { ...this.docsList[i], ...doc };
    } else {
        this.docsList.push(doc);
    }
}
,

async fetchDocFromApiForViewer(docId) {
    const headers = { Accept: 'application/json' };
    if (this.authToken) headers['Authorization'] = `Bearer ${this.authToken}`;
    const res = await fetch(`${this.apiBaseUrl}/docs/${encodeURIComponent(docId)}`, { headers });
    if (!res.ok) throw new Error(await readApiErrorResponse(res));
    return res.json();
}
,

async openDocFromCommunity(docId) {
    const id = String(docId || '').trim();
    if (!id) return;
    this.closeCommunityPostDetailModal();
    this.switchScreen('docs');
    this.flashStatus('Downloading document…');
    try {
        await this.ensureMockSession();
        const full = await this.fetchDocFromApiForViewer(id);
        this.mergeDocIntoList(full);
        this.invalidateDocsCache();
        await this.loadDocs(true);
        const q = document.getElementById('docSearch')?.value?.trim() || '';
        this.renderDocsGrid(q);
        await this.openDocModal(id);
        this.flashStatus('Opened in Docs');
    } catch (e) {
        console.error(e);
        alert(e.message || 'Could not download document');
        this.updateUI();
    }
}
,

async openDocModal(docId) {
    let doc = this.docsList.find((d) => String(d.id) === String(docId));
    const modal = document.getElementById('docModal');
    const titleEl = document.getElementById('docModalTitle');
    const bodyEl = document.getElementById('docModalBody');
    if (!modal || !titleEl || !bodyEl) return;

    const typ = (doc?.type || '').toLowerCase();
    const needsFetch =
        !doc ||
        (['markdown', 'text'].includes(typ) && !String(doc.content || '').trim()) ||
        (typ === 'pdf' && !doc.stored_filename && !String(doc.content || '').trim());

    if (needsFetch) {
        try {
            doc = await this.fetchDocFromApiForViewer(docId);
            this.mergeDocIntoList(doc);
        } catch (e) {
            console.error(e);
            alert(e.message || 'Could not load document');
            return;
        }
    }
    if (!doc) {
        alert('Document not found');
        return;
    }

    const t = (doc.type || '').toLowerCase();
    titleEl.textContent = doc.title || 'Document';
    if (t === 'image' && doc.id) {
        const url = `${this.apiBaseUrl}/docs/${encodeURIComponent(doc.id)}/file`;
        bodyEl.classList.remove('doc-modal-body--md');
        const summary = this.escapeHtml(doc.ai_summary || 'Image document.');
        const safeUrl = this.escapeHtml(url);
        bodyEl.innerHTML = `<p class="doc-file-summary">${summary}</p>
            <div class="doc-open-external-wrap">
                <a href="${safeUrl}" target="_blank" rel="noopener noreferrer" class="btn-primary">Open in browser</a>
            </div>`;
    } else if (t === 'pdf' && doc.id) {
        const url = `${this.apiBaseUrl}/docs/${encodeURIComponent(doc.id)}/file`;
        bodyEl.classList.remove('doc-modal-body--md');
        const summary = this.escapeHtml(doc.ai_summary || 'PDF document.');
        const safeUrl = this.escapeHtml(url);
        bodyEl.innerHTML = `<p class="doc-file-summary">${summary}</p>
            <div class="doc-open-external-wrap">
                <a href="${safeUrl}" target="_blank" rel="noopener noreferrer" class="btn-primary">Open PDF in browser</a>
            </div>`;
    } else {
        bodyEl.classList.add('doc-modal-body--md');
        const raw = doc.content || doc.ai_summary || '(No body)';
        bodyEl.innerHTML = markdownToHtml(raw);
    }
    modal.hidden = false;
}
,

closeDocModal() {
    const modal = document.getElementById('docModal');
    if (modal) modal.hidden = true;
}
,

async publishDocToCommunity(docId) {
    await this.ensureMockSession();
    const doc = this.docsList.find((d) => d.id === docId);
    const title = doc?.title || 'Document';
    const note = window.prompt(`Share “${title}” to Community. Add an optional note:`, '') ?? '';
    try {
        const res = await fetch(`${this.apiBaseUrl}/community/posts`, {
            method: 'POST',
            headers: this.authBearerHeaders({ 'Content-Type': 'application/json', Accept: 'application/json' }),
            body: JSON.stringify({
                content: note.trim(),
                doc_id: docId,
                doc_title: title,
                tags: ['#docs', '#Published'],
            }),
        });
        if (!res.ok) throw new Error(await readApiErrorResponse(res));
        this.flashStatus('Published to Community');
        this.switchScreen('community');
        this.invalidateCommunityCache();
        await this.loadCommunity(true);
    } catch (e) {
        alert(e.message || 'Could not publish');
    }
}

};

export function bindDocsEvents(app) {
    document.getElementById('createDocBtn')?.addEventListener('click', () => app.openCreateDocModal());
    document.getElementById('createDocModalClose')?.addEventListener('click', () => app.closeCreateDocModal());
    document.getElementById('createDocModalBackdrop')?.addEventListener('click', () => app.closeCreateDocModal());
    document.getElementById('createDocCancel')?.addEventListener('click', () => app.closeCreateDocModal());
    document.getElementById('createDocSaveBtn')?.addEventListener('click', () => app.submitCreateDoc(false));
    document.getElementById('createDocSaveAnalyzeBtn')?.addEventListener('click', () => app.submitCreateDoc(true));
    document.getElementById('createDocFile')?.addEventListener('change', () => app.updateCreateDocFileName());

    document.getElementById('learnModalClose')?.addEventListener('click', () => app.closeLearnModal());
    document.getElementById('learnModalBackdrop')?.addEventListener('click', () => app.closeLearnModal());
    document.getElementById('learnModalDismiss')?.addEventListener('click', () => app.closeLearnModal());
    document.getElementById('learnModalBackground')?.addEventListener('click', () => {
        app.requestLearnNotificationFromGesture();
        app.closeLearnModal();
    });
    document.getElementById('learnModalSaveDoc')?.addEventListener('click', () => app.saveLearnAnalysisToDocs());

    const docSearch = document.getElementById('docSearch');
    if (docSearch) {
        const debouncedDocSearch = debounce((q) => app.renderDocsGrid(q), 200);
        docSearch.addEventListener('input', () => debouncedDocSearch(docSearch.value.trim()));
    }

    document.getElementById('docModalClose')?.addEventListener('click', () => app.closeDocModal());
    document.getElementById('docModalBackdrop')?.addEventListener('click', () => app.closeDocModal());

    document.getElementById('docsGrid')?.addEventListener('click', (e) => {
        const pub = e.target.closest('.doc-publish-btn');
        if (pub?.dataset?.docId) {
            e.stopPropagation();
            void app.publishDocToCommunity(pub.dataset.docId);
            return;
        }
        const learn = e.target.closest('.doc-learn-btn');
        if (learn?.dataset?.docId) {
            e.stopPropagation();
            void app.runLearnForDoc(learn.dataset.docId);
            return;
        }
        const card = e.target.closest('.doc-card');
        if (card?.dataset?.docId) {
            app.openDocModal(card.dataset.docId);
        }
    });
}
