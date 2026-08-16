import { readApiErrorResponse } from './api.js';
import { escapeHtml } from './markdown.js';

/** @typedef {import('./app.js').AcademiApp} AcademiApp */



/** Methods mixed onto AcademiApp.prototype — `this` is the app instance. */
export const communityMethods = {
openCommunityCreateModal() {
    const ta = document.getElementById('communityModalPostInput');
    if (ta) ta.value = '';
    const fileEl = document.getElementById('communityPostFile');
    if (fileEl) fileEl.value = '';
    this._communityPendingDocId = null;
    this._communityAnalysisText = '';
    this.updateCommunityPostFileName();
    const preview = document.getElementById('communityAnalysisPreview');
    if (preview) {
        preview.hidden = true;
        preview.innerHTML = '';
    }
    const analyzeToggle = document.getElementById('communityAnalyzeToggle');
    if (analyzeToggle) analyzeToggle.checked = true;
    const m = document.getElementById('communityCreateModal');
    if (m) m.hidden = false;
}
,

updateCommunityPostFileName() {
    const fileEl = document.getElementById('communityPostFile');
    const nameEl = document.getElementById('communityPostFileName');
    if (!nameEl) return;
    const fileName = fileEl?.files?.[0]?.name;
    nameEl.textContent = fileName ? `Selected: ${fileName}` : 'No file selected';
}
,

async uploadCommunityAttachment(file) {
    await this.ensureMockSession();
    const headers = {};
    if (this.authToken) headers['Authorization'] = `Bearer ${this.authToken}`;
    const fd = new FormData();
    fd.append('file', file);
    const res = await fetch(`${this.apiBaseUrl}/docs/upload`, { method: 'POST', headers, body: fd });
    if (!res.ok) throw new Error(await readApiErrorResponse(res));
    return res.json();
}
,

async analyzeCommunityAttachment(docId) {
    const headers = { 'Content-Type': 'application/json', Accept: 'application/json' };
    if (this.authToken) headers['Authorization'] = `Bearer ${this.authToken}`;
    const res = await fetch(`${this.apiBaseUrl}/ai/learn`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
            doc_id: docId,
            disable_research: !this.researchEnabled,
            message: 'Provide a concise community-friendly summary: key points, who it helps, and one practical takeaway.',
        }),
    });
    if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error || 'Analysis failed');
    }
    const data = await res.json();
    return data.response || '';
}
,

async handleCommunityFileSelected(fileList) {
    const file = fileList?.[0];
    if (!file) return;
    this.updateCommunityPostFileName();
    const preview = document.getElementById('communityAnalysisPreview');
    const analyzeOn = document.getElementById('communityAnalyzeToggle')?.checked !== false;

    try {
        this.flashStatus('Uploading attachment…');
        const doc = await this.uploadCommunityAttachment(file);
        this._communityPendingDocId = doc.id;
        this.invalidateDocsCache();

        if (!analyzeOn) {
            if (preview) {
                preview.hidden = false;
                preview.innerHTML = `<p class="form-hint">Attached: <strong>${this.escapeHtml(doc.title || file.name)}</strong> (no AI analysis)</p>`;
            }
            this.flashStatus('File attached');
            return;
        }

        if (preview) {
            preview.hidden = false;
            preview.innerHTML = '<p class="guide-loading" style="margin:0">Analyzing attachment…</p>';
        }
        const analysis = await this.analyzeCommunityAttachment(doc.id);
        this._communityAnalysisText = analysis;
        if (preview) {
            preview.innerHTML = `
                <p class="form-label">AI analysis</p>
                <div class="community-analysis-text">${this.escapeHtml(analysis).replace(/\n/g, '<br>')}</div>`;
        }
        this.flashStatus('Analysis ready — review before posting');
    } catch (e) {
        console.error(e);
        this._communityPendingDocId = null;
        this._communityAnalysisText = '';
        if (preview) {
            preview.hidden = false;
            preview.innerHTML = `<p class="guide-error" style="margin:0">${this.escapeHtml(e.message || 'Upload/analysis failed')}</p>`;
        }
        alert(e.message || 'Could not process file');
    }
}
,

closeCommunityCreateModal() {
    const m = document.getElementById('communityCreateModal');
    if (m) m.hidden = true;
}
,

closeCommunityPostDetailModal() {
    this._communityDetailPostId = null;
    const m = document.getElementById('communityPostDetailModal');
    if (m) m.hidden = true;
}
,

truncatePreview(text, maxLen) {
    const s = (text || '').replace(/\s+/g, ' ').trim();
    if (s.length <= maxLen) return s;
    return s.slice(0, maxLen - 1) + '…';
}
,

renderCommunityPostRow(post) {
    const wrap = document.createElement('button');
    wrap.type = 'button';
    wrap.className = 'community-post-row glass';
    wrap.dataset.postId = post.id;
    const preview = this.truncatePreview(post.content, 160);
    const author = this.escapeHtml(post.author_name || 'Member');
    const when = this.escapeHtml(this.formatSessionTime(post.created_at));
    const replies = post.comments ?? 0;
    const ups = post.upvotes ?? 0;
    wrap.innerHTML = `
        <div class="community-post-row-top">
            <span class="community-post-row-author">${author}</span>
            <span class="community-post-row-time">${when}</span>
        </div>
        <p class="community-post-row-preview">${this.escapeHtml(preview || '(No text)')}</p>
        <div class="community-post-row-meta">
            <span>△ ${ups}</span><span>${replies} repl${replies === 1 ? 'y' : 'ies'}</span>
        </div>
    `;
    return wrap;
}
,

async openCommunityPostDetail(postId) {
    await this.ensureMockSession();
    this._communityDetailPostId = postId;
    const modal = document.getElementById('communityPostDetailModal');
    if (modal) modal.hidden = false;
    await this.refreshCommunityPostDetail();
}
,

async refreshCommunityPostDetail() {
    const postId = this._communityDetailPostId;
    if (!postId) return;
    try {
        const headers = this.authBearerHeaders({ Accept: 'application/json' });
        const postUrl = `${this.apiBaseUrl}/community/posts/${encodeURIComponent(postId)}`;
        const commentsUrl = `${this.apiBaseUrl}/community/posts/${encodeURIComponent(postId)}/comments`;
        const [pres, crs] = await Promise.all([
            fetch(postUrl, { headers }),
            fetch(commentsUrl, { headers }),
        ]);
        if (!pres.ok) throw new Error(await readApiErrorResponse(pres));

        const post = await pres.json();
        let comments = [];
        if (crs.ok) {
            const raw = await crs.json();
            comments = Array.isArray(raw) ? raw : [];
        }

        this.patchCommunityPost(post);
        this.fillCommunityPostDetail(post, comments);
    } catch (e) {
        console.error(e);
        alert(e.message || 'Could not load post');
        this.closeCommunityPostDetailModal();
    }
}
,

patchCommunityPost(post) {
    if (!post?.id) return;
    const idx = this._communityPosts.findIndex((p) => p.id === post.id);
    if (idx >= 0) {
        this._communityPosts[idx] = { ...this._communityPosts[idx], ...post };
    }
}
,

patchCommunityPostVotes(postId, upvotes, downvotes) {
    const post = this._communityPosts.find((p) => p.id === postId);
    if (post) {
        post.upvotes = upvotes;
        post.downvotes = downvotes;
    }
    const row = document.querySelector(`.community-post-row[data-post-id="${CSS.escape(String(postId))}"]`);
    if (row) {
        const meta = row.querySelector('.community-post-row-meta');
        if (meta) {
            const replies = post?.comments ?? 0;
            meta.innerHTML = `<span>△ ${upvotes ?? 0}</span><span>${replies} repl${replies === 1 ? 'y' : 'ies'}</span>`;
        }
    }
}
,

patchCommunityPostCommentCount(postId, delta) {
    const post = this._communityPosts.find((p) => p.id === postId);
    if (post) {
        post.comments = (post.comments ?? 0) + delta;
        this.patchCommunityPostVotes(postId, post.upvotes ?? 0, post.downvotes ?? 0);
    }
}
,

fillCommunityPostDetail(post, comments) {
    const author = post.author_name || 'Member';
    const when = this.formatSessionTime(post.created_at);
    const meta = document.getElementById('communityPostDetailMeta');
    if (meta) meta.textContent = `${author} · ${when}`;

    const contentEl = document.getElementById('communityPostDetailContent');
    if (contentEl) contentEl.textContent = post.content || '';

    const tagsEl = document.getElementById('communityPostDetailTags');
    if (tagsEl) {
        const tags = post.tags || [];
        if (tags.length) {
            tagsEl.innerHTML = tags.map((t) => `<span class="tag-chip">${this.escapeHtml(t)}</span>`).join('');
            tagsEl.hidden = false;
        } else {
            tagsEl.innerHTML = '';
            tagsEl.hidden = true;
        }
    }

    const docWrap = document.getElementById('communityPostDetailDoc');
    if (docWrap) {
        if (post.doc_id) {
            docWrap.innerHTML = `<button type="button" class="action-btn small community-detail-open-doc" data-doc-id="${this.escapeHtml(post.doc_id)}">Open in Docs</button>`;
            docWrap.hidden = false;
        } else {
            docWrap.innerHTML = '';
            docWrap.hidden = true;
        }
    }

    const upEl = document.getElementById('communityDetailUpCount');
    const downEl = document.getElementById('communityDetailDownCount');
    if (upEl) upEl.textContent = String(post.upvotes ?? 0);
    if (downEl) downEl.textContent = String(post.downvotes ?? 0);

    const list = document.getElementById('communityPostDetailComments');
    if (list) {
        if (!comments.length) {
            list.innerHTML = '<p class="guide-hint" style="margin:0;font-size:12px">No replies yet.</p>';
        } else {
            list.innerHTML = comments
                .map((c) => {
                    const a = this.escapeHtml(c.author_name || 'Member');
                    const txt = this.escapeHtml(c.content || '');
                    return `<div class="community-comment-line"><span class="community-comment-author">${a}</span>${txt}</div>`;
                })
                .join('');
        }
    }
}
,

async loadCommunity(force = false) {
    await this.ensureMockSession();
    const feed = document.getElementById('communityFeed');
    if (!feed) return;

    const now = Date.now();
    if (!force && this._communityPosts.length && now - this._communityLoadedAt < this._screenCacheMs) {
        this.renderCommunityFeed(this._communityPosts);
        return;
    }
    if (this._communityLoadPromise) return this._communityLoadPromise;

    this._communityLoadPromise = (async () => {
        feed.innerHTML = '<p class="guide-loading">Loading community…</p>';
        try {
            const res = await fetch(`${this.apiBaseUrl}/community/posts`, {
                headers: this.authBearerHeaders({ Accept: 'application/json' }),
            });
            if (!res.ok) throw new Error(await readApiErrorResponse(res));
            const rawPosts = await res.json();
            this._communityPosts = Array.isArray(rawPosts) ? rawPosts.filter((p) => p && typeof p === 'object' && p.id) : [];
            this._communityLoadedAt = Date.now();
            this.renderCommunityFeed(this._communityPosts);
        } catch (e) {
            console.error(e);
            feed.innerHTML = `<p class="guide-error">${this.escapeHtml(e.message || 'Failed to load')}</p>`;
        }
    })().finally(() => {
        this._communityLoadPromise = null;
    });

    return this._communityLoadPromise;
}
,

renderCommunityFeed(posts) {
    const feed = document.getElementById('communityFeed');
    if (!feed) return;
    feed.innerHTML = '';
    if (!posts.length) {
        feed.innerHTML = '<p class="guide-hint">No posts yet. Tap ＋ to add one.</p>';
        return;
    }
    for (const p of posts) {
        feed.appendChild(this.renderCommunityPostRow(p));
    }
}
,

async submitCommunityPost() {
    await this.ensureMockSession();
    const input = document.getElementById('communityModalPostInput');
    let text = (input?.value || '').trim();
    const docId = this._communityPendingDocId || '';
    const analysis = (this._communityAnalysisText || '').trim();

    if (!text && !docId) {
        alert('Write something or attach a file to post.');
        return;
    }

    if (analysis && text) {
        text += `\n\n---\n**AI analysis:**\n${analysis}`;
    } else if (analysis && !text) {
        text = `**AI analysis of shared file:**\n\n${analysis}`;
    }

    const submitBtn = document.getElementById('communityModalPostSubmit');
    const origLabel = submitBtn?.textContent;
    if (submitBtn) {
        submitBtn.disabled = true;
        submitBtn.textContent = 'Posting…';
    }

    try {
        const body = { content: text, tags: [] };
        if (docId) {
            body.doc_id = docId;
            body.tags = ['#docs', '#ai-analysis'];
        }
        const res = await fetch(`${this.apiBaseUrl}/community/posts`, {
            method: 'POST',
            headers: this.authBearerHeaders({ 'Content-Type': 'application/json', Accept: 'application/json' }),
            body: JSON.stringify(body),
        });
        if (!res.ok) throw new Error(await readApiErrorResponse(res));
        if (input) input.value = '';
        this._communityPendingDocId = null;
        this._communityAnalysisText = '';
        this.closeCommunityCreateModal();
        this.invalidateCommunityCache();
        await this.loadCommunity(true);
        this.flashStatus('Posted to community');
    } catch (e) {
        console.error(e);
        alert(e.message || 'Could not post');
    } finally {
        if (submitBtn) {
            submitBtn.disabled = false;
            submitBtn.textContent = origLabel || 'Post';
        }
    }
}
,

async voteCommunityPost(postId, type) {
    await this.ensureMockSession();
    try {
        const res = await fetch(`${this.apiBaseUrl}/community/posts/${encodeURIComponent(postId)}/vote`, {
            method: 'POST',
            headers: this.authBearerHeaders({ 'Content-Type': 'application/json', Accept: 'application/json' }),
            body: JSON.stringify({ type }),
        });
        if (!res.ok) throw new Error(await readApiErrorResponse(res));
        const data = await res.json();
        this.patchCommunityPostVotes(postId, data.upvotes, data.downvotes);
        if (this._communityDetailPostId === postId) {
            const upEl = document.getElementById('communityDetailUpCount');
            const downEl = document.getElementById('communityDetailDownCount');
            if (upEl) upEl.textContent = String(data.upvotes ?? 0);
            if (downEl) downEl.textContent = String(data.downvotes ?? 0);
        }
    } catch (e) {
        alert(e.message || 'Vote failed');
    }
}
,

async submitCommunityCommentFromDetail() {
    const postId = this._communityDetailPostId;
    if (!postId) return;
    const ta = document.getElementById('communityDetailReplyInput');
    const btn = document.getElementById('communityDetailReplySend');
    const text = (ta?.value || '').trim();
    if (!text) return;
    await this.submitCommunityComment(postId, text, ta, btn);
}
,

async submitCommunityComment(postId, text, textareaEl, buttonEl) {
    await this.ensureMockSession();
    const orig = buttonEl?.textContent;
    if (buttonEl) {
        buttonEl.disabled = true;
        buttonEl.textContent = '…';
    }
    try {
        const res = await fetch(`${this.apiBaseUrl}/community/posts/${encodeURIComponent(postId)}/comments`, {
            method: 'POST',
            headers: this.authBearerHeaders({ 'Content-Type': 'application/json', Accept: 'application/json' }),
            body: JSON.stringify({ content: text }),
        });
        if (!res.ok) throw new Error(await readApiErrorResponse(res));
        if (textareaEl) textareaEl.value = '';
        this.patchCommunityPostCommentCount(postId, 1);
        if (this._communityDetailPostId === postId) {
            await this.refreshCommunityPostDetail();
        }
    } catch (e) {
        alert(e.message || 'Reply failed');
    } finally {
        if (buttonEl) {
            buttonEl.disabled = false;
            buttonEl.textContent = orig || 'Reply';
        }
    }
}

};

export function bindCommunityEvents(app) {
    document.getElementById('communityAddPostBtn')?.addEventListener('click', () => app.openCommunityCreateModal());
    document.getElementById('communityCreateModalClose')?.addEventListener('click', () => app.closeCommunityCreateModal());
    document.getElementById('communityCreateModalBackdrop')?.addEventListener('click', () => app.closeCommunityCreateModal());
    document.getElementById('communityCreateModalCancel')?.addEventListener('click', () => app.closeCommunityCreateModal());
    document.getElementById('communityModalPostSubmit')?.addEventListener('click', () => void app.submitCommunityPost());
    document.getElementById('communityPostFile')?.addEventListener('change', (e) => {
        const files = e.target.files;
        if (!files?.length) return;
        void app.handleCommunityFileSelected(files).finally(() => {
            e.target.value = '';
        });
    });

    document.getElementById('communityPostDetailClose')?.addEventListener('click', () => app.closeCommunityPostDetailModal());
    document.getElementById('communityPostDetailBackdrop')?.addEventListener('click', () => app.closeCommunityPostDetailModal());
    document.getElementById('communityDetailVoteUp')?.addEventListener('click', () => {
        if (app._communityDetailPostId) void app.voteCommunityPost(app._communityDetailPostId, 'up');
    });
    document.getElementById('communityDetailVoteDown')?.addEventListener('click', () => {
        if (app._communityDetailPostId) void app.voteCommunityPost(app._communityDetailPostId, 'down');
    });
    document.getElementById('communityDetailReplySend')?.addEventListener('click', () =>
        void app.submitCommunityCommentFromDetail(),
    );

    document.getElementById('communityFeed')?.addEventListener('click', (e) => {
        const row = e.target.closest('.community-post-row');
        if (row?.dataset?.postId) {
            e.preventDefault();
            void app.openCommunityPostDetail(row.dataset.postId);
        }
    });

    document.getElementById('communityPostDetailModal')?.addEventListener('click', (e) => {
        const od = e.target.closest('.community-detail-open-doc');
        if (od?.dataset?.docId) {
            e.preventDefault();
            void app.openDocFromCommunity(od.dataset.docId);
        }
    });
}
