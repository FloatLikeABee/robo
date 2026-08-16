import { resolveAcademiApiBaseUrl, authBearerHeaders, sessionHeaders as makeSessionHeaders } from './api.js';
import { escapeHtml as htmlEscape, markdownToHtml as mdToHtml } from './markdown.js';
import { chatMethods, bindChatEvents } from './chat.js';
import { docsMethods, bindDocsEvents } from './docs.js';
import { communityMethods, bindCommunityEvents } from './community.js';

export class AcademiApp {
    constructor() {
        this.currentScreen = 'chat';
        this.messages = [
            {
                id: 1,
                text: "Hello! I'm Docs, your AI study assistant. Ask me anything or pick a topic below!",
                isUser: false,
                hasSource: false,
            },
        ];
        this.isTyping = false;
        this.messageId = 1;
        this.apiBaseUrl = resolveAcademiApiBaseUrl();
        this.authToken = localStorage.getItem('authToken') || null;
        this.documentMode = localStorage.getItem('academi_document_mode') === '1';
        this.researchEnabled = localStorage.getItem('academi_research') !== '0';
        this.helpYouLearnMode = localStorage.getItem('academi_help_learn') === '1';
        this.pendingDocIds = [];
        this.docsList = [];
        this._statusResetTimer = null;
        this._learnAnalysisText = '';
        this._learnAnalysisTitle = 'Learning analysis';
        this._learnJobSeq = 0;
        this._learnFetchPending = false;
        this._learnWantNotifyForJobId = null;
        this._learnLastAnalysisOk = true;
        this.currentSessionId = null;
        this._persistTimer = null;
        this.currentUser = null;
        try {
            this.currentUser = JSON.parse(localStorage.getItem('authUser') || 'null');
        } catch {
            this.currentUser = null;
        }
        this._communityDetailPostId = null;
        this._communityPendingDocId = null;
        this._communityAnalysisText = '';
        this._messageDomById = new Map();
        this._markdownCache = new Map();
        this._communityPosts = [];
        this._docsLoadedAt = 0;
        this._communityLoadedAt = 0;
        this._screenCacheMs = 30000;
        this._docsLoadPromise = null;
        this._communityLoadPromise = null;

        this.init();
    }

    init() {
        this.bindEvents();
        this.switchScreen('chat');
        void this.ensureSession()
            .then(() => {
                this.applyProfileUser();
                return this.bootstrapChatSession();
            })
            .finally(() => {
                this.updateSendButton();
            });
    }

    /** MorphAI apps menu passes ?userspanel_token=… — exchange for Academi JWT (same UsersPanel auth). */
    async ensureSession() {
        try {
            const params = new URLSearchParams(window.location.search);
            const upToken = params.get('userspanel_token');
            if (upToken) {
                const res = await fetch(`${this.apiBaseUrl}/auth/dev-login`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
                    body: JSON.stringify({ token: upToken }),
                });
                if (res.ok) {
                    const data = await res.json();
                    if (data.token) {
                        this.authToken = data.token;
                        localStorage.setItem('authToken', data.token);
                    }
                    if (data.user) {
                        this.currentUser = data.user;
                        localStorage.setItem('authUser', JSON.stringify(data.user));
                    }
                    params.delete('userspanel_token');
                    const url = new URL(window.location.href);
                    url.search = params.toString();
                    window.history.replaceState({}, '', url.toString());
                    this.applyProfileUser();
                    return;
                }
                console.warn('UsersPanel SSO failed:', await res.text());
            }
            if (this.authToken) {
                this.applyProfileUser();
                return;
            }
            await this.ensureMockSession();
        } catch (e) {
            console.warn('Session bootstrap:', e);
            if (!this.authToken) await this.ensureMockSession();
        }
    }

    async ensureMockSession() {
        try {
            const res = await fetch(`${this.apiBaseUrl}/auth/mock`, { method: 'POST' });
            if (!res.ok) return;
            const data = await res.json();
            if (data.token) {
                this.authToken = data.token;
                localStorage.setItem('authToken', data.token);
            }
            if (data.user) {
                this.currentUser = data.user;
                localStorage.setItem('authUser', JSON.stringify(data.user));
            }
            this.applyProfileUser();
        } catch (e) {
            console.warn('Demo session:', e);
        }
    }

    applyProfileUser() {
        const el = document.getElementById('profileUserName');
        if (!el) return;
        const name = this.currentUser?.name || this.currentUser?.email || 'Guest';
        el.textContent = name;
    }

    authBearerHeaders(extra = {}) {
        return authBearerHeaders(this.authToken, extra);
    }

    sessionHeaders() {
        return makeSessionHeaders(this.authToken);
    }

    escapeHtml(text) {
        return htmlEscape(text);
    }

    markdownToHtml(text) {
        return mdToHtml(text);
    }

    bindEvents() {
        document.querySelectorAll('.tab-btn').forEach((btn) => {
            btn.addEventListener('click', (e) => {
                const screen = e.currentTarget.dataset.screen;
                this.switchScreen(screen);
            });
        });

        document.getElementById('themeToggle')?.addEventListener('click', () => {
            this.toggleTheme();
        });

        bindChatEvents(this);
        bindDocsEvents(this);
        bindCommunityEvents(this);
    }

    switchScreen(screenName) {
        document.querySelectorAll('.tab-btn').forEach((btn) => {
            btn.classList.remove('active');
        });
        const activeTab = document.querySelector(`[data-screen="${screenName}"]`);
        if (activeTab) {
            activeTab.classList.add('active');
        }

        document.querySelectorAll('.screen').forEach((screen) => {
            screen.classList.remove('active');
            screen.style.removeProperty('display');
        });

        const activeScreen = document.getElementById(`${screenName}Screen`);
        if (activeScreen) {
            activeScreen.classList.add('active');
        }

        this.currentScreen = screenName;
        this.updateUI();
        if (screenName === 'docs') {
            void this.loadDocs();
        }
        if (screenName === 'community') {
            void this.loadCommunity();
        }
    }

    invalidateDocsCache() {
        this._docsLoadedAt = 0;
    }

    invalidateCommunityCache() {
        this._communityLoadedAt = 0;
    }

    flashStatus(msg) {
        const statusText = document.getElementById('statusText');
        if (!statusText) return;
        if (this._statusResetTimer) clearTimeout(this._statusResetTimer);
        statusText.textContent = msg;
        this._statusResetTimer = setTimeout(() => {
            this.updateUI();
            this._statusResetTimer = null;
        }, 4500);
    }

    toggleTheme() {
        const root = document.documentElement;
        const currentTheme = root.getAttribute('data-theme') || 'dark';
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';

        root.setAttribute('data-theme', newTheme);
        document.getElementById('themeToggle').textContent = newTheme === 'dark' ? '🌙 Dark' : '☀️ Light';
    }

    updateUI() {
        const statusText = document.getElementById('statusText');
        if (!statusText) return;
        if (this.currentScreen === 'chat') {
            statusText.textContent = this.isTyping ? 'Thinking...' : 'Online';
        } else {
            statusText.textContent = 'Ready';
        }
    }
}

Object.assign(
    AcademiApp.prototype,
    chatMethods,
    docsMethods,
    communityMethods,
);
