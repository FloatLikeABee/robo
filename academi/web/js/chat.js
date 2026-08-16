import { readApiErrorResponse } from './api.js';
import { escapeHtml, markdownToHtml } from './markdown.js';

/** @typedef {import('./app.js').AcademiApp} AcademiApp */



/** Methods mixed onto AcademiApp.prototype — `this` is the app instance. */
export const chatMethods = {
updateSendButton() {
    const messageInput = document.getElementById('messageInput');
    const sendButton = document.getElementById('sendButton');
    const hasText = messageInput.value.trim().length > 0;
    const hasAttach = this.pendingDocIds.length > 0;

    sendButton.disabled = !hasText && !hasAttach;
    sendButton.style.opacity = hasText || hasAttach ? '1' : '0.5';
    const badge = document.getElementById('attachBadge');
    if (badge) {
        if (this.pendingDocIds.length > 0) {
            badge.hidden = false;
            badge.textContent = `${this.pendingDocIds.length} file(s)`;
        } else {
            badge.hidden = true;
        }
    }
}
,

adjustTextareaHeight() {
    const textarea = document.getElementById('messageInput');
    textarea.style.height = 'auto';
    textarea.style.height = Math.max(36, Math.min(textarea.scrollHeight, 100)) + 'px';
}
,

closeSuggestionsPanel() {
    const panel = document.getElementById('suggestionsPanel');
    const toggleBtn = document.getElementById('suggestionsToggle');
    if (!panel || !toggleBtn) return;
    panel.classList.remove('active');
    toggleBtn.classList.remove('active');
}
,

toggleSuggestionsPanel() {
    const panel = document.getElementById('suggestionsPanel');
    const toggleBtn = document.getElementById('suggestionsToggle');
    if (!panel || !toggleBtn) return;
    panel.classList.toggle('active');
    toggleBtn.classList.toggle('active');
}
,

async ensureSession() {
    if (this.currentSessionId) return;
    const res = await fetch(`${this.apiBaseUrl}/chat-sessions`, {
        method: 'POST',
        headers: this.sessionHeaders(),
    });
    if (!res.ok) throw new Error('Could not create session');
    const sess = await res.json();
    this.currentSessionId = sess.id;
    localStorage.setItem('academi_chat_session_id', sess.id);
    void this.renderChatHistoryList();
}
,

async sendMessage(text = null) {
    try {
        await this.ensureSession();
    } catch (e) {
        console.error(e);
    }

    const messageInput = document.getElementById('messageInput');
    let messageText = text != null ? text : messageInput.value.trim();

    if (!messageText && this.pendingDocIds.length === 0) return;

    if (!messageText && this.pendingDocIds.length > 0) {
        messageText = this.helpYouLearnMode
            ? 'Help me learn from the attached material with research context where useful.'
            : 'Here are attached documents — please read them in context of our conversation.';
    }

    this.addMessage(messageText, true);
    messageInput.value = '';
    this.updateSendButton();
    this.adjustTextareaHeight();

    this.showTyping(true);

    const docIdsSnapshot = [...this.pendingDocIds];
    try {
        const response = await this.callAIChat();
        this.pendingDocIds = [];
        this.updateSendButton();
        this.addMessage(response.text, false, response.sources, {
            offerRagSave: response.offerSaveAnalysis,
            pendingSave: response.pendingSave || null,
        });
        if (response.savedDocument) {
            this.flashStatus(`Saved to Docs: ${response.savedDocument.title}`);
            void this.loadDocs(true);
        }
    } catch (error) {
        console.error('AI chat error:', error);
        this.pendingDocIds = docIdsSnapshot;
        this.updateSendButton();
        const detail = error?.message ? String(error.message) : '';
        const userMsg = detail
            ? `Sorry, something went wrong: ${detail}`
            : 'Sorry, I encountered an error. Please try again.';
        this.addMessage(userMsg, false, []);
        if (detail) this.flashStatus(detail.slice(0, 120));
    }

    this.persistSession();
    this.showTyping(false);
}
,

addMessage(text, isUser, sources = [], opts = {}) {
    this.messageId++;
    const message = {
        id: this.messageId,
        text: text,
        isUser: isUser,
        hasSource: !isUser && sources.length > 0,
        sources: sources,
        offerRagSave: !isUser && !!opts.offerRagSave,
        pendingSave: opts.pendingSave || null,
    };

    this.messages.push(message);
    this.appendMessageElement(message);
}
,

createMessageElement(message) {
    const messageElement = document.createElement('div');
    messageElement.className = `message ${message.isUser ? 'user-message' : 'ai-message'} fade-in`;
    messageElement.dataset.messageId = String(message.id);

    let sourcesHtml = '';
    if (message.hasSource && message.sources && message.sources.length > 0) {
        const iconFor = (t) => {
            if (t === 'wiki') return '⌁';
            if (t === 'paper') return '⧗';
            if (t === 'internal') return '⟲';
            return '⧉';
        };
        sourcesHtml = `
            <div class="message-sources">
                ${message.sources.map((source) => {
                    const title = this.escapeHtml(source.title || '');
                    const url = source.url ? String(source.url) : '';
                    const inner = url
                        ? `<a href="${this.escapeHtml(url)}" target="_blank" rel="noopener noreferrer" class="source-link">${title}</a>`
                        : `<span>${title}</span>`;
                    return `
                    <div class="source-item">
                        <span class="source-icon">${iconFor(source.type)}</span>
                        ${inner}
                    </div>`;
                }).join('')}
            </div>
        `;
    }

    messageElement.innerHTML = `
        <div class="message-content">
            <div class="message-text">${this.formatMessage(message.text, message.id)}</div>
            ${sourcesHtml}
            ${!message.isUser && message.offerRagSave ? `
                <div class="rag-save-prompt">
                    <span class="rag-save-label">Save this analysis to Docs for later RAG retrieval?</span>
                    <button class="btn-primary rag-save-btn" data-message-id="${message.id}" type="button">Save to Docs</button>
                    <button class="btn-secondary rag-dismiss-btn" data-message-id="${message.id}" type="button">Not now</button>
                </div>
            ` : ''}
            ${!message.isUser ? `
                <div class="message-actions">
                    <button class="action-btn small save-ai-btn" data-message-id="${message.id}" type="button">⍟ Save</button>
                    <button class="action-btn small">⧉ Share</button>
                    <button class="action-btn small copy-ai-btn" data-message-id="${message.id}" type="button">⎘ Copy</button>
                </div>
            ` : ''}
        </div>
    `;
    return messageElement;
}
,

appendMessageElement(message) {
    const messagesContainer = document.getElementById('messages');
    if (!messagesContainer) return;
    const el = this.createMessageElement(message);
    messagesContainer.appendChild(el);
    this._messageDomById.set(message.id, el);
    this.scrollMessagesToBottom();
}
,

scrollMessagesToBottom() {
    const scrollRoot = document.getElementById('messagesContainer');
    const scrollToBottom = () => {
        if (scrollRoot) {
            scrollRoot.scrollTop = scrollRoot.scrollHeight;
        }
    };
    requestAnimationFrame(() => {
        requestAnimationFrame(scrollToBottom);
    });
}
,

renderMessages() {
    const messagesContainer = document.getElementById('messages');
    if (!messagesContainer) return;
    messagesContainer.innerHTML = '';
    this._messageDomById.clear();
    this._markdownCache.clear();

    this.messages.forEach((message) => {
        const messageElement = this.createMessageElement(message);
        messagesContainer.appendChild(messageElement);
        this._messageDomById.set(message.id, messageElement);
    });

    this.scrollMessagesToBottom();
}
,

formatMessage(text, cacheKey) {
    if (cacheKey != null && this._markdownCache.has(cacheKey)) {
        return this._markdownCache.get(cacheKey);
    }
    const html = markdownToHtml(text);
    if (cacheKey != null) {
        this._markdownCache.set(cacheKey, html);
    }
    return html;
}
,

buildApiMessages() {
    return this.messages.map((m) => ({
        role: m.isUser ? 'user' : 'assistant',
        content: m.text,
    }));
}
,

async callAIChat() {
    const headers = {
        'Content-Type': 'application/json',
    };

    if (this.authToken) {
        headers['Authorization'] = `Bearer ${this.authToken}`;
    }

    const response = await fetch(`${this.apiBaseUrl}/ai/chat`, {
        method: 'POST',
        headers: headers,
        body: JSON.stringify({
            messages: this.buildApiMessages(),
            context: {},
            document_mode: this.documentMode,
                disable_research:
                    this.documentMode || this.helpYouLearnMode || this.pendingDocIds.length > 0
                        ? !this.researchEnabled
                        : true,
                doc_ids: [...this.pendingDocIds],
            help_you_learn: this.helpYouLearnMode,
        }),
    });

    if (!response.ok) {
        const errBody = await response.json().catch(() => ({}));
        throw new Error(errBody.error || 'Failed to get AI response');
    }

    const data = await response.json();
    return {
        text: data.response,
        sources: data.sources || [],
        savedDocument: data.saved_document || null,
        offerSaveAnalysis: !!data.offer_save_analysis,
        pendingSave: data.pending_save || null,
    };
}
,

async saveRagAnalysisFromMessage(messageId, buttonEl) {
    const message = this.messages.find((m) => m.id === messageId && !m.isUser);
    if (!message) return;

    const originalText = buttonEl?.textContent || 'Save to Docs';
    if (buttonEl) {
        buttonEl.disabled = true;
        buttonEl.textContent = 'Saving…';
    }

    const pending = message.pendingSave;
    const contentToSave = pending?.content || this.stripIntroAndAsk(message.text) || message.text;
    const title = pending?.title || this.generateDocTitleFromContent(contentToSave);
    const tags = pending?.tags?.length ? pending.tags : ['#analysis', '#rag'];

    const headers = { 'Content-Type': 'application/json', Accept: 'application/json' };
    if (this.authToken) headers['Authorization'] = `Bearer ${this.authToken}`;

    try {
        const res = await fetch(`${this.apiBaseUrl}/docs`, {
            method: 'POST',
            headers,
            body: JSON.stringify({
                title,
                type: 'markdown',
                content: contentToSave,
                tags,
            }),
        });
        if (!res.ok) throw new Error(await readApiErrorResponse(res));
        const doc = await res.json();
        message.offerRagSave = false;
        message.pendingSave = null;
        this._learnAnalysisText = contentToSave;
        this._learnAnalysisTitle = title;
        const el = this._messageDomById.get(messageId);
        el?.querySelector('.rag-save-prompt')?.remove();
        this.flashStatus(`Saved to Docs for RAG: ${doc.title || title}`);
        await this.loadDocs(true);
        if (buttonEl) buttonEl.textContent = 'Saved';
    } catch (e) {
        console.error(e);
        alert(e.message || 'Could not save to Docs');
        if (buttonEl) {
            buttonEl.disabled = false;
            buttonEl.textContent = originalText;
        }
    }
}
,

async uploadChatFiles(fileList) {
    await this.ensureMockSession();
    const headers = {};
    if (this.authToken) headers['Authorization'] = `Bearer ${this.authToken}`;
    const files = [...fileList];
    const results = await Promise.allSettled(
        files.map(async (file) => {
            const fd = new FormData();
            fd.append('file', file);
            const res = await fetch(`${this.apiBaseUrl}/docs/upload`, { method: 'POST', headers, body: fd });
            if (!res.ok) throw new Error(await readApiErrorResponse(res));
            return res.json();
        }),
    );
    let uploaded = 0;
    for (const result of results) {
        if (result.status === 'fulfilled' && result.value?.id) {
            this.pendingDocIds.push(result.value.id);
            uploaded++;
        } else if (result.status === 'rejected') {
            console.error(result.reason);
            alert(result.reason?.message || 'Upload failed');
        }
    }
    this.updateSendButton();
    if (uploaded > 0) {
        this.flashStatus(`${uploaded} file(s) ready to send`);
    }
}
,

async saveAIMessageToDocs(messageId, buttonEl) {
    const message = this.messages.find((m) => m.id === messageId && !m.isUser);
    if (!message) return;

    const cleanedContent = this.stripIntroAndAsk(message.text);
    const contentToSave = cleanedContent || message.text;
    const title = this.generateDocTitleFromContent(contentToSave);

    const originalText = buttonEl?.textContent || '⍟ Save';
    if (buttonEl) {
        buttonEl.disabled = true;
        buttonEl.textContent = 'Saving...';
    }

    const headers = { 'Content-Type': 'application/json', Accept: 'application/json' };
    if (this.authToken) headers['Authorization'] = `Bearer ${this.authToken}`;

    try {
        const res = await fetch(`${this.apiBaseUrl}/docs`, {
            method: 'POST',
            headers,
            body: JSON.stringify({
                title,
                type: 'markdown',
                content: contentToSave,
                tags: ['#saved-from-chat'],
            }),
        });
        if (!res.ok) throw new Error(await readApiErrorResponse(res));
        const doc = await res.json();
        this.flashStatus(`Saved to Docs: ${doc.title || title}`);
        await this.loadDocs(true);
        if (buttonEl) buttonEl.textContent = 'Saved';
    } catch (e) {
        console.error(e);
        alert(e.message || 'Could not save this response');
        if (buttonEl) buttonEl.textContent = originalText;
    } finally {
        if (buttonEl) {
            setTimeout(() => {
                buttonEl.disabled = false;
                if (buttonEl.textContent === 'Saved') buttonEl.textContent = originalText;
            }, 1200);
        }
    }
}
,

async copyAIMessageToClipboard(messageId, buttonEl) {
    const message = this.messages.find((m) => m.id === messageId && !m.isUser);
    if (!message?.text) return;

    const originalText = buttonEl?.textContent || '⎘ Copy';
    const text = message.text;

    const showCopied = () => {
        this.flashStatus('Copied to clipboard');
        if (buttonEl) {
            buttonEl.textContent = 'Copied';
            setTimeout(() => {
                buttonEl.textContent = originalText;
            }, 1200);
        }
    };

    try {
        if (navigator.clipboard?.writeText) {
            await navigator.clipboard.writeText(text);
            showCopied();
            return;
        }
    } catch (e) {
        console.warn('Clipboard API failed', e);
    }

    const ta = document.createElement('textarea');
    ta.value = text;
    ta.setAttribute('readonly', '');
    ta.style.position = 'fixed';
    ta.style.left = '-9999px';
    document.body.appendChild(ta);
    ta.select();
    try {
        if (document.execCommand('copy')) {
            showCopied();
        } else {
            alert('Could not copy — select the message text manually.');
        }
    } finally {
        document.body.removeChild(ta);
    }
}
,

openChatHistoryDrawer() {
    const d = document.getElementById('chatHistoryDrawer');
    if (d) d.hidden = false;
    void this.renderChatHistoryList();
}
,

closeChatHistoryDrawer() {
    const d = document.getElementById('chatHistoryDrawer');
    if (d) d.hidden = true;
}
,

messagesToRows() {
    return this.messages.map((m) => ({
        role: m.isUser ? 'user' : 'assistant',
        content: m.text,
        sources: (m.sources || []).map((s) => ({
            title: s.title || '',
            type: s.type || 'internal',
            url: s.url || '',
        })),
    }));
}
,

persistSession() {
    if (!this.currentSessionId) return;
    clearTimeout(this._persistTimer);
    this._persistTimer = setTimeout(() => void this.flushSession(), 450);
}
,

async flushSession() {
    const id = this.currentSessionId;
    if (!id) return;
    const rows = this.messagesToRows();
    try {
        await fetch(`${this.apiBaseUrl}/chat-sessions/${encodeURIComponent(id)}`, {
            method: 'PATCH',
            headers: this.sessionHeaders(),
            body: JSON.stringify({ messages: rows }),
        });
        const drawer = document.getElementById('chatHistoryDrawer');
        if (drawer && !drawer.hidden) {
            void this.renderChatHistoryList();
        }
    } catch (e) {
        console.warn('Session save failed', e);
    }
}
,

resetMessagesToGreeting() {
    this.messageId = 1;
    this.messages = [
        {
            id: 1,
            text: "Hello! I'm Docs, your AI study assistant. Ask me anything or use the ⚡ button for quick actions!",
            isUser: false,
            hasSource: false,
            sources: [],
        },
    ];
}
,

applySessionPayload(sess) {
    this.currentSessionId = sess.id;
    localStorage.setItem('academi_chat_session_id', sess.id);
    if (sess.messages && sess.messages.length > 0) {
        this.messages = sess.messages.map((m, i) => ({
            id: i + 1,
            text: m.content,
            isUser: m.role === 'user',
            hasSource: !!(m.sources && m.sources.length),
            sources: m.sources || [],
        }));
        this.messageId = this.messages.length;
    } else {
        this.resetMessagesToGreeting();
    }
    this.pendingDocIds = [];
    this.updateSendButton();
    this.renderMessages();
}
,

async bootstrapChatSession() {
    try {
        const savedId = localStorage.getItem('academi_chat_session_id');
        if (savedId) {
            const res = await fetch(`${this.apiBaseUrl}/chat-sessions/${encodeURIComponent(savedId)}`, {
                headers: this.sessionHeaders(),
            });
            if (res.ok) {
                const sess = await res.json();
                this.applySessionPayload(sess);
                return;
            }
        }
        await this.createNewChatSession();
    } catch (e) {
        console.warn('Chat session bootstrap', e);
        this.resetMessagesToGreeting();
        this.currentSessionId = null;
        this.renderMessages();
    }
}
,

async createNewChatSession() {
    try {
        const res = await fetch(`${this.apiBaseUrl}/chat-sessions`, {
            method: 'POST',
            headers: this.sessionHeaders(),
        });
        if (!res.ok) throw new Error('Could not start chat');
        const sess = await res.json();
        this.applySessionPayload(sess);
        this.closeChatHistoryDrawer();
        await this.flushSession();
        void this.renderChatHistoryList();
    } catch (e) {
        console.error(e);
        alert(
            'Could not create a new chat session. Open this app using your computer\'s LAN address (not localhost), or set localStorage key academi_api_base to your API URL.',
        );
    }
}
,

async loadChatSession(id) {
    try {
        const res = await fetch(`${this.apiBaseUrl}/chat-sessions/${encodeURIComponent(id)}`, {
            headers: this.sessionHeaders(),
        });
        if (!res.ok) throw new Error('Not found');
        const sess = await res.json();
        this.applySessionPayload(sess);
        this.closeChatHistoryDrawer();
        void this.renderChatHistoryList();
    } catch (e) {
        console.error(e);
        alert('Could not open that chat.');
    }
}
,

async deleteChatSession(id) {
    if (!window.confirm('Delete this chat from history?')) return;
    try {
        await fetch(`${this.apiBaseUrl}/chat-sessions/${encodeURIComponent(id)}`, {
            method: 'DELETE',
            headers: this.sessionHeaders(),
        });
        if (this.currentSessionId === id) {
            localStorage.removeItem('academi_chat_session_id');
            await this.createNewChatSession();
        }
        void this.renderChatHistoryList();
    } catch (e) {
        console.error(e);
    }
}
,

formatSessionTime(ts) {
    if (!ts) return '';
    const d = new Date(ts * 1000);
    const now = new Date();
    const sameDay = d.toDateString() === now.toDateString();
    if (sameDay) return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    return d.toLocaleDateString([], { month: 'short', day: 'numeric' });
}
,

async renderChatHistoryList() {
    const ul = document.getElementById('chatHistoryList');
    if (!ul) return;
    try {
        const res = await fetch(`${this.apiBaseUrl}/chat-sessions`, { headers: this.sessionHeaders() });
        if (!res.ok) throw new Error('list failed');
        const items = await res.json();
        ul.innerHTML = '';
        for (const row of items) {
            const li = document.createElement('li');
            li.className = 'chat-history-row';
            const isActive = row.id === this.currentSessionId;
            const t = this.escapeHtml(row.title || 'Chat');
            const preview = this.escapeHtml(row.preview || '');
            const when = this.escapeHtml(this.formatSessionTime(row.updated_at));
            li.innerHTML = `
                <button type="button" class="chat-history-item${isActive ? ' active' : ''}" data-id="${this.escapeHtml(row.id)}">
                    <span class="chat-history-item-title">${t}</span>
                    <span class="chat-history-item-meta">${preview}<br><span style="opacity:0.8">${when}</span></span>
                </button>
                <button type="button" class="chat-history-del" data-id="${this.escapeHtml(row.id)}" title="Delete">✕</button>
            `;
            ul.appendChild(li);
        }
    } catch (e) {
        ul.innerHTML = '<li class="chat-history-hint">Could not load history.</li>';
    }
}
,

stripIntroAndAsk(text) {
    const lines = String(text || '').replace(/\r\n/g, '\n').split('\n');
    const cleaned = [...lines];

    while (cleaned.length && !cleaned[0].trim()) cleaned.shift();
    while (cleaned.length && !cleaned[cleaned.length - 1].trim()) cleaned.pop();

    const introPattern = /^(hi|hello|hey|great question|good question|nice question|sure|absolutely|of course|let me|here(?:'| i)s|thanks for|happy to)/i;
    const leadingQuestionPattern = /^(.*\?)$/;
    const trailingAskPattern = /(\?$)|(would you like|do you want|want me to|let me know|anything else|follow up)/i;

    if (cleaned.length && introPattern.test(cleaned[0].trim())) {
        cleaned.shift();
    }
    if (cleaned.length && leadingQuestionPattern.test(cleaned[0].trim()) && cleaned[0].trim().length < 120) {
        cleaned.shift();
    }
    while (cleaned.length && trailingAskPattern.test(cleaned[cleaned.length - 1].trim())) {
        cleaned.pop();
    }

    while (cleaned.length && !cleaned[0].trim()) cleaned.shift();
    while (cleaned.length && !cleaned[cleaned.length - 1].trim()) cleaned.pop();
    return cleaned.join('\n').trim();
}
,

generateDocTitleFromContent(content) {
    const firstLine = String(content || '')
        .split('\n')
        .map((line) => line.trim())
        .find((line) => line.length > 0) || '';

    if (!firstLine) return 'Study notes';

    let title = firstLine.replace(/^[-*#>\d.\s]+/, '').trim();
    if (!title) title = firstLine.trim();
    if (!title) return 'Study notes';

    const asciiWords = title
        .replace(/[^\w\s-]/g, ' ')
        .replace(/\s+/g, ' ')
        .trim()
        .split(' ')
        .filter(Boolean)
        .slice(0, 8);

    if (asciiWords.length > 0) {
        const candidate = asciiWords.join(' ');
        const titled = candidate.charAt(0).toUpperCase() + candidate.slice(1);
        return titled.length > 70 ? `${titled.slice(0, 67).trimEnd()}...` : titled;
    }

    const raw = title.replace(/\s+/g, ' ').trim();
    return raw.length > 70 ? `${raw.slice(0, 67).trimEnd()}...` : raw;
}
,

generateAIResponse(query) {
    const responses = {
        summarize: "Here's a concise summary of the topic:\n\n• Key concept 1: The fundamental principles\n• Key concept 2: The practical applications\n• Key concept 3: The future implications\n\nWould you like me to elaborate on any specific point?",
        explain: "Let me break this down step by step:\n\n1. First, we need to understand the basic framework\n2. Then we look at how the components interact\n3. Finally, we apply this to real-world scenarios\n\nThis approach will help you build a solid understanding.",
        default: "Great question! Based on my analysis across multiple sources:\n\nThe core concept revolves around three main pillars:\n\n📌 Theory — The foundational principles\n📌 Practice — Real-world applications\n📌 Innovation — Emerging developments\n\nI've pulled information from academic papers, community discussions, and web sources. Check the sources below for full references.",
    };

    const lowerQuery = query.toLowerCase();
    if (lowerQuery.includes('summarize')) return responses.summarize;
    if (lowerQuery.includes('explain')) return responses.explain;
    return responses.default;
}
,

showTyping(show) {
    const typingIndicator = document.getElementById('typingIndicator');
    const statusGlow = document.getElementById('statusGlow');
    const statusText = document.getElementById('statusText');

    if (show) {
        typingIndicator.style.display = 'flex';
        statusGlow.classList.add('active');
        statusText.textContent = 'Thinking...';
        this.isTyping = true;
    } else {
        typingIndicator.style.display = 'none';
        statusGlow.classList.remove('active');
        statusText.textContent = 'Online';
        this.isTyping = false;
    }
}

};

export function bindChatEvents(app) {
    const messageInput = document.getElementById('messageInput');
    const sendButton = document.getElementById('sendButton');

    messageInput?.addEventListener('input', () => {
        app.updateSendButton();
        app.adjustTextareaHeight();
    });

    messageInput?.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            app.sendMessage();
        }
    });

    sendButton?.addEventListener('click', () => app.sendMessage());

    const docToggle = document.getElementById('documentModeToggle');
    const researchToggle = document.getElementById('researchToggle');
    if (docToggle) {
        docToggle.checked = app.documentMode;
        docToggle.addEventListener('change', () => {
            app.documentMode = docToggle.checked;
            localStorage.setItem('academi_document_mode', app.documentMode ? '1' : '0');
        });
    }
    if (researchToggle) {
        researchToggle.checked = app.researchEnabled;
        researchToggle.addEventListener('change', () => {
            app.researchEnabled = researchToggle.checked;
            localStorage.setItem('academi_research', app.researchEnabled ? '1' : '0');
        });
    }

    const helpLearnToggle = document.getElementById('helpLearnToggle');
    if (helpLearnToggle) {
        helpLearnToggle.checked = app.helpYouLearnMode;
        helpLearnToggle.addEventListener('change', () => {
            app.helpYouLearnMode = helpLearnToggle.checked;
            localStorage.setItem('academi_help_learn', app.helpYouLearnMode ? '1' : '0');
            app.updateSendButton();
        });
    }

    const chatFileInput = document.getElementById('chatFileInput');
    if (chatFileInput) {
        chatFileInput.addEventListener('change', (e) => {
            const files = e.target.files;
            if (!files?.length) return;
            void app.uploadChatFiles(files).finally(() => {
                e.target.value = '';
            });
        });
    }

    const suggestionsToggle = document.getElementById('suggestionsToggle');
    if (suggestionsToggle) {
        suggestionsToggle.addEventListener('click', () => app.toggleSuggestionsPanel());
    }
    const suggestionsPanel = document.getElementById('suggestionsPanel');
    if (suggestionsPanel && suggestionsToggle) {
        const dismissIfOutside = (e) => {
            if (!suggestionsPanel.classList.contains('active')) return;
            const t = e.target;
            if (suggestionsPanel.contains(t) || suggestionsToggle.contains(t)) return;
            app.closeSuggestionsPanel();
        };
        document.addEventListener('pointerdown', dismissIfOutside, { passive: true });
    }

    document.querySelectorAll('.quick-action-btn').forEach((btn) => {
        btn.addEventListener('click', (e) => {
            const query = e.currentTarget.dataset.query;
            app.closeSuggestionsPanel();
            app.sendMessage(query);
        });
    });

    document.querySelectorAll('.tag-chip[data-topic]').forEach((tag) => {
        tag.addEventListener('click', (e) => {
            const topic = e.currentTarget.dataset.topic;
            app.closeSuggestionsPanel();
            app.sendMessage(`Tell me about ${topic}`);
        });
    });

    document.getElementById('chatHistoryBtn')?.addEventListener('click', () => app.openChatHistoryDrawer());
    document.getElementById('chatHistoryBackdrop')?.addEventListener('click', () => app.closeChatHistoryDrawer());
    document.getElementById('chatHistoryClose')?.addEventListener('click', () => app.closeChatHistoryDrawer());
    document.getElementById('newChatSessionBtn')?.addEventListener('click', () => app.createNewChatSession());

    document.getElementById('chatHistoryList')?.addEventListener('click', (e) => {
        const del = e.target.closest('.chat-history-del');
        if (del) {
            e.preventDefault();
            e.stopPropagation();
            const id = del.dataset.id;
            if (id) void app.deleteChatSession(id);
            return;
        }
        const item = e.target.closest('.chat-history-item');
        if (item) {
            const id = item.dataset.id;
            if (id) void app.loadChatSession(id);
        }
    });

    document.getElementById('messages')?.addEventListener('click', (e) => {
        const ragBtn = e.target.closest('.rag-save-btn');
        if (ragBtn) {
            const messageId = Number(ragBtn.dataset.messageId || 0);
            if (messageId) void app.saveRagAnalysisFromMessage(messageId, ragBtn);
            return;
        }
        const ragDismiss = e.target.closest('.rag-dismiss-btn');
        if (ragDismiss) {
            const messageId = Number(ragDismiss.dataset.messageId || 0);
            const message = app.messages.find((m) => m.id === messageId);
            if (message) {
                message.offerRagSave = false;
                message.pendingSave = null;
            }
            app._messageDomById.get(messageId)?.querySelector('.rag-save-prompt')?.remove();
            return;
        }
        const saveBtn = e.target.closest('.save-ai-btn');
        if (saveBtn) {
            const messageId = Number(saveBtn.dataset.messageId || 0);
            if (messageId) void app.saveAIMessageToDocs(messageId, saveBtn);
            return;
        }
        const copyBtn = e.target.closest('.copy-ai-btn');
        if (copyBtn) {
            const messageId = Number(copyBtn.dataset.messageId || 0);
            if (messageId) void app.copyAIMessageToClipboard(messageId, copyBtn);
        }
    });
}
