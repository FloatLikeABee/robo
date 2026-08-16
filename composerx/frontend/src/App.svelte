<script>
  import { onDestroy, onMount, tick } from 'svelte'
  import { fade } from 'svelte/transition'
  import TableFooterBar from './lib/TableFooterBar.svelte'
  import ButtonLeadingIcon from './lib/ButtonLeadingIcon.svelte'
  import PlatformAssistantDrawer from './components/PlatformAssistantDrawer.svelte'
  import ComposePublishPanel from './components/ComposePublishPanel.svelte'
  import { runAiProgress } from '@robo/platform-chat/aiProgress'
  import ComposePublishRecordsPanel from './components/ComposePublishRecordsPanel.svelte'
  import ContentMarkdownPanel from './components/ContentMarkdownPanel.svelte'
  import { downloadMarkdownFile, savedContentMarkdown } from './lib/contentMarkdown'
  import { resolveComposerProposedDraft } from './lib/composerDraft'

  const PAGES = {
    EMAIL_COMPOSER: 'email-composer',
    COMPOSE_CONTENT: 'compose-content',
    COMPOSE_PUBLISH: 'compose-publish', // legacy alias → compose-content
    COMPOSE_PUBLISH_RECORDS: 'compose-publish-records',
    EMAILS: 'emails',
    REFERENCE_DOCS: 'reference-docs' // legacy; redirected to compose-content
  }

  // Dev: default to same-origin + Vite proxy (see vite.config.js). Prod: set VITE_API_BASE.
  const API_BASE = (
    import.meta.env.VITE_API_BASE ??
    (import.meta.env.DEV ? '' : 'http://localhost:8043')
  ).replace(/\/$/, '')
  const AUTH_TOKEN_KEY = 'tranmail_auth_token'
  const SHARED_SESSION_COOKIE_KEY = 'userspanel_session_token'
  const AUTH_REMEMBER_KEY = 'tranmail_auth_remember'
  /** Align with UsersPanel JWT default (JWT_EXPIRY_HOURS, default 48). */
  const SESSION_COOKIE_MAX_AGE_SECONDS = 48 * 3600
  const THEME_KEY = 'tranmail-theme'
  const LAST_PAGE_KEY = 'email-composer-last-page'
  const TABLE_PAGE_SIZE_DEFAULT = 25
  const MAX_AI_COMPOSER_REF_DOCS = 3

  /** @param {number} total @param {number} limit */
  function lastPageStart(total, limit) {
    if (total <= 0 || limit <= 0) return 0
    return Math.floor((total - 1) / limit) * limit
  }

  function normalizePageId(page) {
    if (page === PAGES.COMPOSE_PUBLISH || page === PAGES.REFERENCE_DOCS || page === 'templates') {
      return PAGES.COMPOSE_CONTENT
    }
    // Contents module removed — published history is enough.
    if (page === PAGES.EMAILS) {
      return PAGES.COMPOSE_PUBLISH_RECORDS
    }
    return page
  }

  let currentPage = $state(PAGES.COMPOSE_CONTENT)
  /** @type {'light' | 'dark'} */
  let theme = $state('light')

  const navItems = [
    { id: PAGES.COMPOSE_CONTENT, label: 'Compose content', icon: 'compose' },
    { id: PAGES.COMPOSE_PUBLISH_RECORDS, label: 'Published contents', icon: 'emails', indent: true }
  ]
  const NAV_PAGE_IDS = new Set([
    ...navItems.map((item) => item.id),
    PAGES.EMAIL_COMPOSER,
    PAGES.COMPOSE_PUBLISH,
    PAGES.REFERENCE_DOCS,
    PAGES.EMAILS
  ])

  /** @param {string} key */
  function navIconPath(key) {
    switch (key) {
      case 'compose':
        return 'M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25Zm2.92 2.83H5v-.92l9.06-9.06.92.92L5.92 20.08ZM20.71 7.04a1 1 0 0 0 0-1.41l-2.34-2.34a1 1 0 0 0-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83Z'
      case 'emails':
        return 'M3 6h18v12H3V6Zm2.4 1.8 6.6 4.8 6.6-4.8H5.4Zm13.6 8.2V9.9l-7 5.1-7-5.1V16h14Z'
      case 'reference':
        return 'M6 4h7l4 4v12H6V4Zm2 2v12h7V9h-3V6H8Zm2 5h5v2h-5v-2Zm0 4h5v2h-5v-2Z'
      case 'jobs':
        return 'M4 4h16v3H4V4Zm2 5h12v11H6V9Zm2 2v7h8v-7H8Zm3 1h2v5h-2v-5Z'
      default:
        return 'M4 4h16v16H4z'
    }
  }

  /** @type {Array<{ id: number, name: string, updated_at: string }>} */
  let savedEmails = $state([])
  let savedEmailsTotal = $state(0)
  let savedEmailsOffset = $state(0)
  let savedEmailsLimit = $state(TABLE_PAGE_SIZE_DEFAULT)
  /** @type {Array<{ id: number, name: string, updated_at: string }>} */
  let savedEmailsComposer = $state([])
  /** @type {number | null} */
  let selectedSavedEmailId = $state(null)
  let templateName = $state('')
  let templateMarkdown = $state('')
  let isSavingEmail = $state(false)
  /** @type {'idle' | 'saving' | 'saved' | 'error'} */
  let autosaveStatus = $state('idle')
  /** @type {ReturnType<typeof setTimeout> | null} */
  let autosaveTimer = null
  let lastPersistedName = ''
  let lastPersistedMarkdown = ''
  let autosaveSkipNext = false
  /** @type {'preview' | 'source'} */
  let composerEditorView = $state('preview')
  /** @type {HTMLTextAreaElement | null} */
  let composerTextareaEl = $state(null)

  let emailDetailOpen = $state(false)
  /**
   * @typedef {{ name?: string, markdown_content?: string, html_content?: string }} EmailDetailRecord
   */
  /** @type {EmailDetailRecord | null} */
  let emailDetailRecord = $state(null)
  let emailDetailLoading = $state(false)
  /** @type {'preview' | 'source'} */
  let emailDetailView = $state('preview')

  let globalError = $state('')
  let globalInfo = $state('')
  /** @type {ReturnType<typeof setTimeout> | null} */
  let toastDismissTimer = null

  const TOAST_AUTO_DISMISS_MS = 4200

  function clearToastDismissTimer() {
    if (toastDismissTimer != null) {
      clearTimeout(toastDismissTimer)
      toastDismissTimer = null
    }
  }
  let authToken = $state('')
  let loginEmail = $state('')
  let loginPassword = $state('')
  let loginLoading = $state(false)
  let loginError = $state('')
  let loginRememberMe = $state(typeof localStorage !== 'undefined' ? localStorage.getItem(AUTH_REMEMBER_KEY) !== '0' : true)

  let confirmModalOpen = $state(false)
  let confirmModalTitle = $state('Confirm')
  let confirmModalMessage = $state('')
  let confirmModalDanger = $state(false)
  let confirmModalConfirmLabel = $state('OK')
  let confirmModalCancelLabel = $state('Cancel')
  /** @type {((ok: boolean) => void) | null} */
  let confirmDialogResolver = null

  /**
   * @param {{ message: string, title?: string, danger?: boolean, confirmLabel?: string, cancelLabel?: string }} opts
   */
  function openConfirmDialog(opts) {
    return new Promise((resolve) => {
      confirmModalTitle = opts.title ?? 'Confirm'
      confirmModalMessage = opts.message
      confirmModalDanger = opts.danger ?? false
      confirmModalConfirmLabel = opts.confirmLabel ?? 'OK'
      confirmModalCancelLabel = opts.cancelLabel ?? 'Cancel'
      confirmDialogResolver = resolve
      confirmModalOpen = true
    })
  }

  function finishConfirmDialog(ok) {
    if (!confirmModalOpen) return
    confirmModalOpen = false
    const r = confirmDialogResolver
    confirmDialogResolver = null
    if (r) r(ok)
  }

  // Reference library (RAG for AI composer)
  let referenceDocs = $state([])
  let isUploadingRefDoc = $state(false)
  let aiRefDocsModalOpen = $state(false)

  // AI composer assistant (right rail on compose page)
  let aiChatLoading = $state(false)
  let aiLoadingStatus = $state('')
  let aiResponseTimeMs = $state(null)
  let aiChatMsgSeq = 0
  /** @type {Array<{ id: number, role: string, content: string, proposedMarkdown?: string | null }>} */
  let aiChatMessages = $state([])
  let aiChatInput = $state('')
  /** @type {string[]} */
  let aiComposerRefIds = $state([])
  let aiComposerUseWebSearch = $state(false)
  /** @type {HTMLDivElement | null} */
  let aiChatLogEl = $state(null)
  /** @type {HTMLDivElement | null} */
  let aiChatEndAnchorEl = $state(null)

  // Platform assistant drawer (global, top-right button)
  let platformAssistantOpen = $state(false)

  $effect(() => {
    const docs = referenceDocs
    const ids = [...aiComposerRefIds]
    if (ids.length === 0) return
    if (!docs || docs.length === 0) return
    const valid = new Set(docs.map((d) => d.id))
    const capped = ids.filter((id) => valid.has(id)).slice(0, MAX_AI_COMPOSER_REF_DOCS)
    const unchanged = capped.length === ids.length && capped.every((v, i) => v === ids[i])
    if (!unchanged) aiComposerRefIds = capped
  })

  /** @param {string} info @param {string} [error] */
  function setMessage(info, error = '') {
    clearToastDismissTimer()
    globalInfo = info
    globalError = error
    if (!(info || '').trim() && !(error || '').trim()) return
    toastDismissTimer = setTimeout(() => {
      toastDismissTimer = null
      globalInfo = ''
      globalError = ''
    }, TOAST_AUTO_DISMISS_MS)
  }

  function getCookie(name) {
    const prefix = `${name}=`
    const row = document.cookie
      .split(';')
      .map((v) => v.trim())
      .find((v) => v.startsWith(prefix))
    if (!row) return ''
    return decodeURIComponent(row.slice(prefix.length))
  }

  function writeCookie(name, value, maxAgeSeconds) {
    document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`
  }

  function writeSessionCookie(name, value) {
    document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; SameSite=Lax`
  }

  function isJwtExpired(token) {
    const parts = token.split('.')
    if (parts.length !== 3) return false
    try {
      let b64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
      const pad = b64.length % 4
      if (pad) b64 += '='.repeat(4 - pad)
      const payload = JSON.parse(atob(b64))
      if (typeof payload.exp !== 'number') return false
      return Date.now() >= payload.exp * 1000 - 60_000
    } catch {
      return false
    }
  }

  function clearAllAuthStores() {
    try {
      localStorage.removeItem(AUTH_TOKEN_KEY)
    } catch {
      // ignore
    }
    try {
      sessionStorage.removeItem(AUTH_TOKEN_KEY)
    } catch {
      // ignore
    }
    writeCookie(SHARED_SESSION_COOKIE_KEY, '', 0)
  }

  function forceRelogin(message = 'Session expired. Please sign in again.') {
    authToken = ''
    clearAllAuthStores()
    loginError = message
  }

  function throwIfUnauthorized(path, method, res) {
    if (res.status !== 401) return
    const m = (method || 'GET').toUpperCase()
    if (m === 'POST' && path === '/auth/login') return
    forceRelogin()
    throw new Error('Session expired')
  }

  function readSessionToken() {
    let t = getCookie(SHARED_SESSION_COOKIE_KEY)
    if (t && isJwtExpired(t)) {
      clearAllAuthStores()
      return ''
    }
    if (t) return t
    const remember = typeof localStorage !== 'undefined' && localStorage.getItem(AUTH_REMEMBER_KEY) !== '0'
    t = remember
      ? localStorage.getItem(AUTH_TOKEN_KEY) || ''
      : sessionStorage.getItem(AUTH_TOKEN_KEY) || ''
    if (t && isJwtExpired(t)) {
      clearAllAuthStores()
      return ''
    }
    if (t && remember) writeCookie(SHARED_SESSION_COOKIE_KEY, t, SESSION_COOKIE_MAX_AGE_SECONDS)
    else if (t) writeSessionCookie(SHARED_SESSION_COOKIE_KEY, t)
    return t || ''
  }

  function storeSessionToken(token, rememberMe) {
    if (!token) {
      clearAllAuthStores()
      return
    }
    try {
      localStorage.setItem(AUTH_REMEMBER_KEY, rememberMe ? '1' : '0')
    } catch {
      // ignore
    }
    if (rememberMe) {
      localStorage.setItem(AUTH_TOKEN_KEY, token)
      sessionStorage.removeItem(AUTH_TOKEN_KEY)
      writeCookie(SHARED_SESSION_COOKIE_KEY, token, SESSION_COOKIE_MAX_AGE_SECONDS)
    } else {
      sessionStorage.setItem(AUTH_TOKEN_KEY, token)
      localStorage.removeItem(AUTH_TOKEN_KEY)
      writeSessionCookie(SHARED_SESSION_COOKIE_KEY, token)
    }
  }

  function getAuthHeaders(extra = {}) {
    if (!authToken) return extra
    return { Authorization: `Bearer ${authToken}`, ...extra }
  }

  async function apiGet(path) {
    const res = await fetch(`${API_BASE}${path}`, { headers: getAuthHeaders() })
    throwIfUnauthorized(path, 'GET', res)
    if (!res.ok) throw new Error(`GET ${path} failed: ${res.status}`)
    return res.json()
  }

  async function apiJson(path, method, body) {
    const res = await fetch(`${API_BASE}${path}`, {
      method,
      headers: getAuthHeaders({ 'Content-Type': 'application/json' }),
      body: JSON.stringify(body)
    })
    throwIfUnauthorized(path, method, res)
    if (!res.ok) {
      const raw = await res.text().catch(() => '')
      let msg = `${method} ${path} failed: ${res.status}`
      if (raw.trim()) {
        try {
          const data = JSON.parse(raw)
          if (data?.error) msg = String(data.error)
          else if (data?.message) msg = String(data.message)
        } catch {
          msg = raw.trim().slice(0, 280)
        }
      }
      if (res.status === 502 || res.status === 504) {
        msg += ' The AI request may have timed out — try again with web search off or a shorter prompt.'
      } else if (res.status === 500 && path.includes('/ai/composer-chat')) {
        msg += ' Check that ComposerX API (:8043), MongoDB, and ai.config.json are running.'
      }
      throw new Error(msg)
    }
    if (res.status === 204) return null
    return res.json()
  }

  async function login() {
    loginError = ''
    loginLoading = true
    try {
      const res = await fetch(`${API_BASE}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: loginEmail.trim(), password: loginPassword })
      })
      if (!res.ok) {
        let msg = `Login failed: ${res.status}`
        try {
          const errBody = await res.json()
          if (errBody?.error) msg = errBody.error
        } catch {
          // ignore
        }
        throw new Error(msg)
      }
      const data = await res.json()
      authToken = data?.token || ''
      if (!authToken) throw new Error('Missing token in login response')
      storeSessionToken(authToken, loginRememberMe)
      await Promise.all([loadSavedEmails(), loadReferenceDocs()])
    } catch (err) {
      const msg = err?.message || 'Login failed'
      if (err instanceof TypeError && /fetch|network|load failed/i.test(msg)) {
        const target = API_BASE || `${window.location.origin} (Vite proxy → :8043)`
        loginError = `Cannot reach TranMail API (${target}). Start the backend on port 8043 and UsersPanel on :5001.`
      } else {
        loginError = msg
      }
      authToken = ''
      storeSessionToken('')
    } finally {
      loginLoading = false
    }
  }

  function logout() {
    authToken = ''
    storeSessionToken('')
  }

  async function loadSavedEmailsTable() {
    try {
      let data = await apiGet(`/emails?limit=${savedEmailsLimit}&offset=${savedEmailsOffset}`)
      savedEmailsTotal = data.total ?? 0
      let items = data.items || []
      if (items.length === 0 && savedEmailsTotal > 0 && savedEmailsOffset > 0) {
        savedEmailsOffset = lastPageStart(savedEmailsTotal, savedEmailsLimit)
        data = await apiGet(`/emails?limit=${savedEmailsLimit}&offset=${savedEmailsOffset}`)
        items = data.items || []
        savedEmailsTotal = data.total ?? savedEmailsTotal
      }
      savedEmails = items
    } catch (err) {
      setMessage('', err.message)
    }
  }

  async function loadSavedEmailsComposer() {
    try {
      const data = await apiGet('/emails?limit=200&offset=0')
      savedEmailsComposer = data.items || []
      savedEmailsTotal = data.total ?? savedEmailsTotal
    } catch (err) {
      setMessage('', err.message)
    }
  }

  async function loadSavedEmails() {
    await Promise.all([loadSavedEmailsTable(), loadSavedEmailsComposer()])
  }

  async function selectSavedEmail(id) {
    try {
      const d = await apiGet(`/emails/${id}`)
      autosaveSkipNext = true
      selectedSavedEmailId = id
      templateName = d.name || ''
      templateMarkdown = d.markdown_content || d.html_content || ''
      lastPersistedName = templateName.trim()
      lastPersistedMarkdown = templateMarkdown.trim()
      autosaveStatus = lastPersistedMarkdown ? 'saved' : 'idle'
      applyComposerAISessionFromSavedEmail(d.composer_ai_session)
      await tick()
      autosaveSkipNext = false
    } catch (err) {
      autosaveSkipNext = false
      setMessage('', err.message)
    }
  }

  async function openComposerWithEmail(id) {
    navigateTo(PAGES.EMAIL_COMPOSER)
    await tick()
    await selectSavedEmail(id)
  }

  async function openEmailDetail(id) {
    emailDetailOpen = true
    emailDetailView = 'preview'
    emailDetailRecord = null
    emailDetailLoading = true
    try {
      emailDetailRecord = await apiGet(`/emails/${id}`)
    } catch (err) {
      setMessage('', err.message)
      emailDetailOpen = false
    } finally {
      emailDetailLoading = false
    }
  }

  function closeEmailDetail() {
    emailDetailOpen = false
    emailDetailRecord = null
    emailDetailView = 'preview'
  }

  function downloadEmailDetail() {
    if (!emailDetailRecord) return
    downloadMarkdownFile(emailDetailRecord.name, savedContentMarkdown(emailDetailRecord))
    setMessage('Download started.')
  }

  async function downloadSavedEmailRow(id, name) {
    try {
      const record = await apiGet(`/emails/${id}`)
      downloadMarkdownFile(record?.name || name, savedContentMarkdown(record))
      setMessage('Download started.')
    } catch (err) {
      setMessage('', err.message)
    }
  }

  async function deleteSavedEmailRow(id) {
    if (!id) return
    if (
      !(await openConfirmDialog({
        title: 'Delete saved content',
        message: 'Are you sure you want to delete this saved content? This cannot be undone.',
        danger: true,
        confirmLabel: 'Delete',
        cancelLabel: 'Cancel'
      }))
    ) {
      return
    }
    try {
      const res = await fetch(`${API_BASE}/emails/${id}`, { method: 'DELETE', headers: getAuthHeaders() })
      throwIfUnauthorized(`/emails/${id}`, 'DELETE', res)
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.error || `Delete failed (${res.status})`)
      }
      if (selectedSavedEmailId === id) {
        selectedSavedEmailId = null
        templateName = ''
        templateMarkdown = ''
        lastPersistedName = ''
        lastPersistedMarkdown = ''
        autosaveStatus = 'idle'
        resetComposerAiSession()
      }
      closeEmailDetail()
      await loadSavedEmails()
      setMessage('Content deleted.')
    } catch (err) {
      setMessage('', err.message)
    }
  }

  async function loadReferenceDocs() {
    try {
      const data = await apiGet('/ai/reference-docs')
      referenceDocs = data.items || []
    } catch (err) {
      setMessage('', err.message)
    }
  }

  function resetComposerAiSession() {
    aiChatMessages = []
    aiChatInput = ''
    aiChatMsgSeq = 0
    aiComposerRefIds = []
  }

  /** @param {unknown} raw */
  function applyComposerAISessionFromSavedEmail(raw) {
    if (raw == null || typeof raw !== 'object') {
      resetComposerAiSession()
      return
    }
    const sess = /** @type {Record<string, unknown>} */ (raw)
    const msgArr = sess.messages
    const msgs = Array.isArray(msgArr) ? msgArr : []
    let maxSeq =
      typeof sess.msg_seq === 'number' && Number.isFinite(sess.msg_seq) ? /** @type {number} */ (sess.msg_seq) : 0
    /** @type {Array<{ id: number, role: string, content: string, proposedMarkdown?: string | null }>} */
    const normalized = []
    for (let idx = 0; idx < msgs.length; idx++) {
      const m = msgs[idx]
      if (!m || typeof m !== 'object') continue
      const mo = /** @type {Record<string, unknown>} */ (m)
      let id =
        typeof mo.id === 'number' && Number.isFinite(mo.id)
          ? /** @type {number} */ (mo.id)
          : idx + 1
      if (id > maxSeq) maxSeq = id
      const role = mo.role === 'assistant' ? 'assistant' : 'user'
      const content = String(mo.content ?? '')
      /** @type {{ id: number, role: string, content: string, proposedMarkdown?: string | null }} */
      const row = { id, role, content }
      if (mo.proposedMarkdown != null || mo.proposed_markdown != null) {
        const ph = mo.proposedMarkdown ?? mo.proposed_markdown
        row.proposedMarkdown = ph != null ? String(ph) : null
      } else if (mo.proposedHtml != null || mo.proposed_html != null || mo.proposed_email_html != null) {
        const ph = mo.proposedHtml ?? mo.proposed_html ?? mo.proposed_email_html
        row.proposedMarkdown = ph != null ? String(ph) : null
      }
      normalized.push(row)
    }
    aiChatMessages = normalized
    aiChatMsgSeq = maxSeq
    const rid = sess.reference_document_ids
    const idList = Array.isArray(rid) ? rid.filter((x) => typeof x === 'string' && x) : []
    aiComposerRefIds = [...new Set(idList)].slice(0, MAX_AI_COMPOSER_REF_DOCS)
    aiChatInput = ''
  }

  async function scrollAiComposerChatToBottom() {
    await tick()
    await new Promise((resolve) => {
      requestAnimationFrame(() => {
        requestAnimationFrame(resolve)
      })
    })
    const logEl = aiChatLogEl
    const anchorEl = aiChatEndAnchorEl
    if (!logEl) return
    if (anchorEl && logEl.contains(anchorEl)) {
      anchorEl.scrollIntoView({ block: 'end', behavior: 'auto' })
      return
    }
    logEl.scrollTop = logEl.scrollHeight
  }

  function serializeComposerAISession() {
    return {
      messages: aiChatMessages.map((m) =>
        m.proposedMarkdown != null
          ? { id: m.id, role: m.role, content: m.content, proposedMarkdown: m.proposedMarkdown }
          : { id: m.id, role: m.role, content: m.content }
      ),
      reference_document_ids: aiComposerRefIds.slice(0, MAX_AI_COMPOSER_REF_DOCS),
      msg_seq: aiChatMsgSeq
    }
  }

  function toggleAiComposerRef(id) {
    if (!id) return
    if (aiComposerRefIds.includes(id)) {
      aiComposerRefIds = aiComposerRefIds.filter((x) => x !== id)
      return
    }
    if (aiComposerRefIds.length >= MAX_AI_COMPOSER_REF_DOCS) {
      setMessage('', `You can select at most ${MAX_AI_COMPOSER_REF_DOCS} reference documents.`)
      return
    }
    aiComposerRefIds = [...aiComposerRefIds, id]
  }

  function formatResponseTime(ms) {
    if (ms == null) return ''
    if (ms < 1000) return `${ms} ms`
    return `${(ms / 1000).toFixed(ms < 10000 ? 1 : 0)} s`
  }

  async function sendAiComposerMessage() {
    const text = aiChatInput.trim()
    if (!text || aiChatLoading) return
    aiChatInput = ''
    const userId = ++aiChatMsgSeq
    const nextMsgs = [...aiChatMessages, { id: userId, role: 'user', content: text }]
    aiChatMessages = nextMsgs
    aiChatLoading = true
    aiResponseTimeMs = null
    aiLoadingStatus = 'Reading your question…'
    const startedAt = performance.now()
    const stopProgress = runAiProgress(
      {
        app: 'composerx',
        userText: text,
        webSearch: aiComposerUseWebSearch,
        hasReferenceDocs: aiComposerRefIds.length > 0,
      },
      (status) => {
        aiLoadingStatus = status
      },
    )
    try {
      const markdownNow = templateMarkdown ?? ''
      const body = {
        messages: nextMsgs.map((m) => ({ role: m.role, content: m.content })),
        reference_document_ids: aiComposerRefIds.slice(0, MAX_AI_COMPOSER_REF_DOCS),
        current_markdown: markdownNow,
        use_web_search: aiComposerUseWebSearch,
      }
      const res = await apiJson('/ai/composer-chat', 'POST', body)
      aiResponseTimeMs = Math.round(performance.now() - startedAt)
      const assistantRaw = (res?.assistant_message ?? res?.assistantMessage ?? '').toString().trim()
      const proposed = resolveComposerProposedDraft(res)
      let displayAssistant = assistantRaw
      if (!displayAssistant && proposed) {
        displayAssistant =
          'Document draft updated in the editor — review Preview or Markdown on the left.'
      }
      if (!displayAssistant) {
        displayAssistant =
          '(Empty reply — check ai.config.json qwen_api_key / server logs, or try again.)'
      }
      if (proposed) {
        templateMarkdown = proposed
        composerEditorView = 'preview'
        const saved = await persistContent({ silent: true })
        if (!saved) {
          setMessage('', 'Draft generated but autosave failed — check the API and try editing the name.')
        }
      } else if (/invitation|email client|inline css|html/i.test(displayAssistant)) {
        setMessage(
          '',
          'The assistant replied without a document draft. Clear chat and try again, or ask explicitly for markdown in proposed_markdown.'
        )
      }
      aiChatMessages = [
        ...nextMsgs,
        { id: ++aiChatMsgSeq, role: 'assistant', content: displayAssistant, proposedMarkdown: proposed || null }
      ]
    } catch (err) {
      aiResponseTimeMs = Math.round(performance.now() - startedAt)
      const msg = err instanceof Error ? err.message : 'Composer chat failed.'
      setMessage('', msg)
      aiChatMessages = [
        ...nextMsgs,
        { id: ++aiChatMsgSeq, role: 'assistant', content: `Could not generate draft: ${msg}`, proposedMarkdown: null }
      ]
    } finally {
      stopProgress()
      aiChatLoading = false
      aiLoadingStatus = ''
    }
  }

  function ensureContentName() {
    const trimmed = templateName.trim()
    if (trimmed) return trimmed
    const match = templateMarkdown.match(/^#\s+(.+)$/m)
    if (match) {
      const derived = match[1].trim().slice(0, 120)
      templateName = derived
      return derived
    }
    return 'Untitled draft'
  }

  /** @param {{ silent?: boolean }} [opts] */
  async function persistContent(opts = {}) {
    const silent = opts.silent !== false
    const md = templateMarkdown.trim()
    if (!md) return false
    const name = ensureContentName()
    if (name === lastPersistedName && md === lastPersistedMarkdown && selectedSavedEmailId) {
      autosaveStatus = 'saved'
      return true
    }
    if (isSavingEmail) return false
    isSavingEmail = true
    autosaveStatus = 'saving'
    const isNew = selectedSavedEmailId == null
    try {
      const body = {
        name,
        markdown_content: templateMarkdown,
        created_by: 1,
        composer_ai_session: serializeComposerAISession()
      }
      if (selectedSavedEmailId) {
        await apiJson(`/emails/${selectedSavedEmailId}`, 'PUT', body)
      } else {
        const created = await apiJson('/emails', 'POST', body)
        selectedSavedEmailId = created.id
      }
      lastPersistedName = name
      lastPersistedMarkdown = md
      autosaveStatus = 'saved'
      if (isNew) {
        await loadSavedEmails()
      } else {
        await loadSavedEmailsComposer()
      }
      if (!silent) setMessage('Content saved.')
      return true
    } catch (err) {
      autosaveStatus = 'error'
      setMessage('', err.message || 'Autosave failed.')
      return false
    } finally {
      isSavingEmail = false
    }
  }

  function scheduleAutosave() {
    if (autosaveSkipNext) return
    if (autosaveTimer != null) clearTimeout(autosaveTimer)
    autosaveTimer = setTimeout(() => {
      autosaveTimer = null
      void persistContent({ silent: true })
    }, 1500)
  }

  /** @param {Event & { currentTarget: HTMLInputElement }} event */
  async function onUploadReferenceDoc(event) {
    const file = event.currentTarget.files?.[0]
    if (!file) return
    isUploadingRefDoc = true
    setMessage('Uploading reference…')
    try {
      const form = new FormData()
      form.append('file', file)
      form.append('name', file.name)
      if (file.type) form.append('mime_type', file.type)
      const res = await fetch(`${API_BASE}/ai/reference-docs/upload`, { method: 'POST', headers: getAuthHeaders(), body: form })
      throwIfUnauthorized('/ai/reference-docs/upload', 'POST', res)
      const data = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(data.error || `Upload failed (${res.status})`)
      await loadReferenceDocs()
      const newId = data?.id
      if (newId) {
        if (!aiComposerRefIds.includes(newId) && aiComposerRefIds.length < MAX_AI_COMPOSER_REF_DOCS) {
          aiComposerRefIds = [...aiComposerRefIds, newId]
        } else if (!aiComposerRefIds.includes(newId)) {
          aiComposerRefIds = [...aiComposerRefIds.slice(1), newId]
        }
      }
      setMessage(`Reference attached: ${data.name || file.name}.`)
    } catch (err) {
      setMessage('', err.message)
    } finally {
      isUploadingRefDoc = false
      event.currentTarget.value = ''
    }
  }

  /** @param {string} id */
  async function deleteReferenceDocRow(id) {
    if (!id) return
    if (
      !(await openConfirmDialog({
        title: 'Delete reference document',
        message: 'Remove this file from the knowledge library? This cannot be undone.',
        danger: true,
        confirmLabel: 'Delete',
        cancelLabel: 'Cancel'
      }))
    ) {
      return
    }
    try {
      const res = await fetch(`${API_BASE}/ai/reference-docs/${encodeURIComponent(id)}`, {
        method: 'DELETE',
        headers: getAuthHeaders()
      })
      throwIfUnauthorized(`/ai/reference-docs/${encodeURIComponent(id)}`, 'DELETE', res)
      if (!res.ok && res.status !== 204) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.error || `Delete failed (${res.status})`)
      }
      aiComposerRefIds = aiComposerRefIds.filter((x) => x !== id)
      await loadReferenceDocs()
      setMessage('Reference deleted.')
    } catch (err) {
      setMessage('', err.message)
    }
  }

  /** @param {string} text */
  function insertPlainAtComposerCursor(text) {
    const t = String(text ?? '')
    if (!t) return
    const el = composerTextareaEl
    if (el) {
      const start = el.selectionStart ?? templateMarkdown.length
      const end = el.selectionEnd ?? start
      templateMarkdown = templateMarkdown.slice(0, start) + t + templateMarkdown.slice(end)
      void tick().then(() => {
        el.focus()
        const pos = start + t.length
        el.setSelectionRange(pos, pos)
      })
      return
    }
    templateMarkdown = `${templateMarkdown || ''}${t}`
  }

  function navigateTo(page) {
    const next = normalizePageId(page)
    currentPage = next
    try {
      localStorage.setItem(LAST_PAGE_KEY, next)
    } catch {
      // ignore
    }
    if (next === PAGES.COMPOSE_CONTENT || next === PAGES.EMAIL_COMPOSER) {
      setMessage('')
      loadSavedEmailsComposer()
      loadReferenceDocs()
    }
  }

  function applyTheme(next) {
    theme = next
    document.documentElement.dataset.theme = next
    try {
      localStorage.setItem(THEME_KEY, next)
    } catch {
      // ignore
    }
  }

  function toggleTheme() {
    applyTheme(theme === 'light' ? 'dark' : 'light')
  }

  onMount(async () => {
    let initial = 'light'
    let initialPage = PAGES.COMPOSE_CONTENT
    try {
      const params = new URLSearchParams(window.location.search)
      const upToken = (params.get('userspanel_token') || '').trim()
      if (upToken) {
        storeSessionToken(upToken, true)
        params.delete('userspanel_token')
        const url = new URL(window.location.href)
        url.search = params.toString()
        window.history.replaceState({}, '', url.toString())
      }
      authToken = readSessionToken()
      const stored = localStorage.getItem(THEME_KEY)
      if (stored === 'light' || stored === 'dark') {
        initial = stored
      } else if (window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
        initial = 'dark'
      }
      const pageStored = localStorage.getItem(LAST_PAGE_KEY) || ''
      if (pageStored === 'merge-data' || pageStored === 'dashboard' || pageStored === 'attachments') {
        initialPage = PAGES.COMPOSE_CONTENT
      } else if (NAV_PAGE_IDS.has(pageStored) || pageStored === 'templates') {
        initialPage = normalizePageId(pageStored)
      }
    } catch {
      // ignore
    }
    applyTheme(initial)
    currentPage = initialPage
    if (!authToken) return

    await Promise.all([
      loadSavedEmails(),
      loadReferenceDocs()
    ])
  })

  onDestroy(() => {
    clearToastDismissTimer()
    if (autosaveTimer != null) clearTimeout(autosaveTimer)
  })

  $effect(() => {
    if (!aiRefDocsModalOpen) return
    /** @param {KeyboardEvent} e */
    const onKey = (e) => {
      if (e.key === 'Escape') aiRefDocsModalOpen = false
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  })

  $effect(() => {
    if (currentPage !== PAGES.EMAIL_COMPOSER) return
    if (autosaveSkipNext) return
    void templateMarkdown
    void templateName
    scheduleAutosave()
  })

  $effect(() => {
    if (currentPage !== PAGES.EMAIL_COMPOSER) return
    void $state.snapshot(aiChatMessages)
    void aiChatLoading
    void scrollAiComposerChatToBottom()
  })

</script>

<main class="shell" class:shell-login={!authToken}>
  {#if !authToken}
    <section class="tm-login-shell">
      <div class="tm-login-card">
        <div class="tm-login-brand">
          <span class="tm-login-icon" aria-hidden="true">✉️</span>
          <div>
            <h1 class="tm-login-title">Content Maker</h1>
            <p class="tm-login-sub">Sign in to continue</p>
          </div>
        </div>
        <form
          class="tm-login-form"
          onsubmit={(e) => {
            e.preventDefault()
            login()
          }}
        >
          <div class="tm-login-field">
            <label for="tm-login-email" class="tm-login-label">Email</label>
            <input
              id="tm-login-email"
              class="tm-login-input"
              type="email"
              autocomplete="email"
              placeholder="you@example.com"
              bind:value={loginEmail}
              required
            />
          </div>
          <div class="tm-login-field">
            <label for="tm-login-password" class="tm-login-label">Password</label>
            <input
              id="tm-login-password"
              class="tm-login-input"
              type="password"
              autocomplete="current-password"
              placeholder="••••••••"
              bind:value={loginPassword}
              required
            />
          </div>
          <label class="tm-login-remember">
            <input type="checkbox" bind:checked={loginRememberMe} />
            <span>Remember me on this device</span>
          </label>
          {#if loginError}
            <div class="tm-login-error" role="alert">{loginError}</div>
          {/if}
          <button class="tm-login-submit" type="submit" disabled={loginLoading}>
            {loginLoading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
      </div>
    </section>
  {:else}
  <header class="shell-topbar">
    <div class="shell-topbar-brand">
      <div class="brand-mark">
        <div class="brand-glyph" aria-hidden="true">
          <svg viewBox="0 0 64 64" class="brand-glyph-svg">
            <defs>
              <linearGradient id="brandGlyphBg" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stop-color="#60a5fa" />
                <stop offset="55%" stop-color="#3b82f6" />
                <stop offset="100%" stop-color="#1d4ed8" />
              </linearGradient>
            </defs>
            <rect x="3" y="3" width="58" height="58" rx="16" fill="url(#brandGlyphBg)" />
            <path
              d="M14 22h36v20H14V22Zm4 4 14 10 14-10H18Zm28 12V29l-14 10-14-10v9h28Z"
              fill="#f8fbff"
            />
            <path d="M40 44h10v4H40z" fill="#dbeafe" />
            <circle cx="50" cy="18" r="5" fill="#bfdbfe" />
          </svg>
        </div>
      </div>
        <div class="brand-copy">
        <div class="brand-title">Content Maker</div>
      </div>
    </div>

    <nav class="shell-tabs" aria-label="Sections">
      {#each navItems as item}
        <button
          type="button"
          class:shell-tab-active={currentPage === item.id}
          class="shell-tab"
          onclick={() => navigateTo(item.id)}
        >
          <span class="nav-item-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" class="nav-icon-svg">
              <path fill="currentColor" d={navIconPath(item.icon)} />
            </svg>
          </span>
          <span>{item.label}</span>
        </button>
      {/each}
    </nav>

    <div class="shell-topbar-actions">
      <button type="button" class="theme-toggle" onclick={toggleTheme} aria-label="Toggle color theme">
        <span class="theme-toggle-hint" aria-hidden="true">{theme === 'light' ? '☾' : '☀'}</span>
      </button>
      <button type="button" class="theme-toggle" onclick={logout} aria-label="Sign out">
        <span class="theme-toggle-label">Out</span>
      </button>
    </div>
  </header>

  <section class="shell-main">
    <header class="main-header">
      <div>
        <h1>
          {#if currentPage === PAGES.COMPOSE_CONTENT || currentPage === PAGES.COMPOSE_PUBLISH}
            Compose content
          {:else if currentPage === PAGES.EMAIL_COMPOSER}
            Compose content
          {:else if currentPage === PAGES.COMPOSE_PUBLISH_RECORDS}
            Published contents
          {/if}
        </h1>
      </div>

      <div class="header-actions">
        <button
          type="button"
          class="btn-secondary"
          aria-expanded={platformAssistantOpen}
          aria-controls="platform-assistant-drawer"
          onclick={() => (platformAssistantOpen = !platformAssistantOpen)}
        >
          <ButtonLeadingIcon name="ai" />
          AI Assistant
        </button>
        {#if currentPage === PAGES.COMPOSE_CONTENT || currentPage === PAGES.EMAIL_COMPOSER || currentPage === PAGES.COMPOSE_PUBLISH}
          <label class="btn-secondary btn-file">
            <input
              type="file"
              accept=".pdf,.txt,.md,.png,.jpg,.jpeg,.gif,.webp,text/plain,text/markdown,application/pdf,image/*"
              onchange={onUploadReferenceDoc}
            />
            <ButtonLeadingIcon name="upload" />
            {#if isUploadingRefDoc}
              Uploading…
            {:else}
              Upload reference
            {/if}
          </label>
        {/if}
      </div>
    </header>

    <div class="shell-main-fill">
      <div class="page-grid">
      {#if currentPage === PAGES.COMPOSE_CONTENT || currentPage === PAGES.COMPOSE_PUBLISH}
        <ComposePublishPanel
          apiBase={API_BASE}
          {getAuthHeaders}
          {theme}
          notify={(kind, msg) => {
            if ((kind || '').toLowerCase() === 'error') setMessage('', msg || '')
            else setMessage(msg || '')
          }}
        />
      {:else if currentPage === PAGES.COMPOSE_PUBLISH_RECORDS}
        <ComposePublishRecordsPanel
          apiBase={API_BASE}
          {getAuthHeaders}
          notify={(kind, msg) => {
            if ((kind || '').toLowerCase() === 'error') setMessage('', msg || '')
            else setMessage(msg || '')
          }}
        />
      {:else if currentPage === PAGES.EMAIL_COMPOSER}
        <section class="panel panel-wide composer">
          <header class="panel-header">
            <div class="field-group composer-name-group">
              <label for="templateName">Content name</label>
              <input
                id="templateName"
                class="composer-name-input"
                placeholder="Monthly account statement"
                bind:value={templateName}
              />
            </div>
            <span class="composer-autosave-status" aria-live="polite">
              {#if autosaveStatus === 'saving'}
                Saving…
              {:else if autosaveStatus === 'saved'}
                Saved
              {:else if autosaveStatus === 'error'}
                Save failed
              {/if}
            </span>
          </header>
          <div class="composer-main">
            <div class="composer-body">
              <div class="composer-editor-col">
              <div class="composer-editor-shell">
                <div class="editor">
                  <div class="editor-toolbar">
                    <span class="pill-label">Content Maker</span>
                    <div class="content-detail-tabs composer-editor-tabs" role="tablist" aria-label="Document view">
                      <button
                        type="button"
                        role="tab"
                        class="content-detail-tab"
                        class:content-detail-tab-active={composerEditorView === 'preview'}
                        aria-selected={composerEditorView === 'preview'}
                        onclick={() => (composerEditorView = 'preview')}
                      >
                        Preview
                      </button>
                      <button
                        type="button"
                        role="tab"
                        class="content-detail-tab"
                        class:content-detail-tab-active={composerEditorView === 'source'}
                        aria-selected={composerEditorView === 'source'}
                        onclick={() => (composerEditorView = 'source')}
                      >
                        Markdown
                      </button>
                    </div>
                  </div>
                  <div class="editor-body-host">
                    {#if composerEditorView === 'source'}
                      <textarea
                        bind:this={composerTextareaEl}
                        class="editor-input markdown-editor"
                        placeholder="# Title&#10;&#10;Write markdown here…"
                        bind:value={templateMarkdown}
                        spellcheck="false"
                      ></textarea>
                    {:else}
                      <div class="composer-preview-host">
                        <ContentMarkdownPanel markdown={templateMarkdown} mode="preview" />
                      </div>
                    {/if}
                  </div>
                </div>
              </div>
            </div>
              <aside class="composer-ai-rail" aria-label="AI assistant">
                <div class="composer-ai-rail-head">
                  <h2 class="composer-ai-rail-title">AI assistant</h2>
                </div>
                <div
                  class="ai-chat-log composer-ai-chat-log"
                  bind:this={aiChatLogEl}
                  role="log"
                  aria-live="polite"
                >
                  {#if aiChatMessages.length === 0}
                    <div class="composer-ai-chat-empty" aria-hidden="true"></div>
                  {:else}
                    {#each aiChatMessages as msg (msg.id)}
                      <div
                        class="ai-chat-bubble"
                        class:ai-chat-user={msg.role === 'user'}
                        class:ai-chat-asst={msg.role !== 'user'}
                      >
                        <div class="ai-chat-role">{msg.role === 'user' ? 'You' : 'Assistant'}</div>
                        <div class="ai-chat-text">{msg.content}</div>
                      </div>
                    {/each}
                  {/if}
                  {#if aiChatLoading}
                    <p class="sidebar-hint ai-loading-status">{aiLoadingStatus || 'Working…'}</p>
                  {/if}
                  <div
                    class="ai-chat-end-anchor"
                    bind:this={aiChatEndAnchorEl}
                    aria-hidden="true"
                  ></div>
                </div>
                <div class="ai-chat-compose composer-ai-compose">
                  <div class="composer-ai-ref-bar">
                    <button
                      type="button"
                      class="btn-secondary composer-ai-ref-docs-btn"
                      onclick={() => {
                        aiRefDocsModalOpen = true
                        loadReferenceDocs()
                      }}
                      aria-haspopup="dialog"
                      aria-expanded={aiRefDocsModalOpen}
                    >
                      <ButtonLeadingIcon name="refDocs" />
                      Reference files…
                    </button>
                    <label class="composer-ai-web-toggle">
                      <input type="checkbox" bind:checked={aiComposerUseWebSearch} />
                      Web search
                    </label>
                    <span class="panel-meta composer-ai-ref-count" aria-live="polite">
                      {aiComposerRefIds.length} / {MAX_AI_COMPOSER_REF_DOCS} selected
                    </span>
                  </div>
                  <textarea
                    class="ai-chat-input composer-ai-chat-input"
                    rows="4"
                    placeholder="e.g. Draft a 3-day workshop agenda as a markdown document…"
                    bind:value={aiChatInput}
                    onkeydown={(e) => {
                      if (e.key === 'Enter' && !e.shiftKey) {
                        e.preventDefault()
                        sendAiComposerMessage()
                      }
                    }}
                  ></textarea>
                  <div class="composer-ai-actions">
                    <button type="button" class="btn-ghost" onclick={() => (aiChatMessages = [])}>Clear chat</button>
                    <button
                      type="button"
                      class="btn-secondary"
                      onclick={sendAiComposerMessage}
                      disabled={aiChatLoading || !aiChatInput.trim()}
                    >
                      <ButtonLeadingIcon name="send" />
                      Send
                    </button>
                  </div>
                  {#if aiResponseTimeMs != null}
                    <div class="ai-response-time">Response {formatResponseTime(aiResponseTimeMs)}</div>
                  {/if}
                </div>
              </aside>
            </div>
          </div>
        </section>
      {/if}
      </div>
    </div>

    {#if globalError.trim() || globalInfo.trim()}
      <div class="toast-stack" aria-live="polite" aria-relevant="additions text" aria-atomic="true">
        {#if globalError.trim()}
          <div class="toast toast-error" transition:fade={{ duration: 160 }}>
            {globalError}
          </div>
        {/if}
        {#if globalInfo.trim()}
          <div class="toast toast-info" transition:fade={{ duration: 160 }}>
            {globalInfo}
          </div>
        {/if}
      </div>
    {/if}

    <PlatformAssistantDrawer bind:open={platformAssistantOpen} apiBase={API_BASE} {getAuthHeaders} />

    {#if emailDetailOpen}
      <div
        class="modal-backdrop"
        role="presentation"
        tabindex="-1"
        onclick={closeEmailDetail}
        onkeydown={(e) => e.key === 'Escape' && closeEmailDetail()}
      >
        <div
          class="modal-panel content-detail-modal"
          role="dialog"
          aria-modal="true"
          aria-labelledby="email-detail-title"
          tabindex="-1"
          onclick={(e) => e.stopPropagation()}
          onkeydown={(e) => e.stopPropagation()}
        >
          <header class="modal-header content-detail-header">
            <h2 id="email-detail-title" class="modal-title">
              {#if emailDetailRecord}
                {emailDetailRecord.name}
              {:else}
                Content preview
              {/if}
            </h2>
            <div class="content-detail-header-actions">
              {#if emailDetailRecord && !emailDetailLoading}
                <div class="content-detail-tabs" role="tablist" aria-label="Content view">
                  <button
                    type="button"
                    role="tab"
                    class="content-detail-tab"
                    class:content-detail-tab-active={emailDetailView === 'preview'}
                    aria-selected={emailDetailView === 'preview'}
                    onclick={() => (emailDetailView = 'preview')}
                  >
                    Preview
                  </button>
                  <button
                    type="button"
                    role="tab"
                    class="content-detail-tab"
                    class:content-detail-tab-active={emailDetailView === 'source'}
                    aria-selected={emailDetailView === 'source'}
                    onclick={() => (emailDetailView = 'source')}
                  >
                    Markdown
                  </button>
                </div>
                <button type="button" class="btn-secondary content-detail-download" onclick={downloadEmailDetail}>
                  <ButtonLeadingIcon name="download" />
                  Download
                </button>
              {/if}
              <button type="button" class="icon-btn modal-close" title="Close" aria-label="Close" onclick={closeEmailDetail}>
                <svg class="icon-svg" viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    fill="currentColor"
                    d="M19 6.41 17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"
                  />
                </svg>
              </button>
            </div>
          </header>
          <div class="modal-body content-detail-body">
            {#if emailDetailLoading}
              <p class="sidebar-hint">Loading…</p>
            {:else if emailDetailRecord}
              <ContentMarkdownPanel
                markdown={savedContentMarkdown(emailDetailRecord)}
                mode={emailDetailView}
              />
            {/if}
          </div>
        </div>
      </div>
    {/if}

    {#if aiRefDocsModalOpen}
      <div
        class="modal-backdrop"
        role="presentation"
        tabindex="-1"
        onclick={() => (aiRefDocsModalOpen = false)}
        onkeydown={(e) => e.key === 'Escape' && (aiRefDocsModalOpen = false)}
      >
        <div
          class="modal-panel ai-ref-docs-modal"
          role="dialog"
          aria-modal="true"
          aria-labelledby="ai-ref-docs-modal-title"
          tabindex="-1"
          onclick={(e) => e.stopPropagation()}
          onkeydown={(e) => {
            if (e.key === 'Escape') {
              e.preventDefault()
              aiRefDocsModalOpen = false
            }
            e.stopPropagation()
          }}
        >
          <header class="modal-header">
            <h2 id="ai-ref-docs-modal-title" class="modal-title">Reference files</h2>
            <button
              type="button"
              class="icon-btn modal-close"
              title="Close"
              aria-label="Close"
              onclick={() => (aiRefDocsModalOpen = false)}
            >
              <svg class="icon-svg" viewBox="0 0 24 24" aria-hidden="true">
                <path
                  fill="currentColor"
                  d="M19 6.41 17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12 19 6.41Z"
                />
              </svg>
            </button>
          </header>
          <div class="modal-body ai-ref-docs-modal-body">
            <p class="sidebar-hint composer-ai-ref-modal-hint">
              Up to {MAX_AI_COMPOSER_REF_DOCS} documents inform the assistant for this draft. Upload references from the
              Compose content toolbar.
            </p>
            <label class="btn-secondary btn-file" style="margin-bottom: 0.75rem;">
              <input
                type="file"
                accept=".pdf,.txt,.md,.png,.jpg,.jpeg,.gif,.webp,text/plain,text/markdown,application/pdf,image/*"
                onchange={onUploadReferenceDoc}
              />
              <ButtonLeadingIcon name="upload" />
              {#if isUploadingRefDoc}
                Uploading…
              {:else}
                Upload reference
              {/if}
            </label>
            {#if referenceDocs.length === 0}
              <p class="sidebar-hint">No reference documents uploaded yet.</p>
            {:else}
              <div class="ai-ref-pick ai-ref-picker-scroll">
                <ul class="ai-ref-list">
                  {#each referenceDocs as doc (doc.id)}
                    <li>
                      <label class="ai-ref-row">
                        <input
                          type="checkbox"
                          checked={aiComposerRefIds.includes(doc.id)}
                          onchange={() => toggleAiComposerRef(doc.id)}
                        />
                        <span class="ai-ref-name">{doc.name}</span>
                        <code class="template-tag ai-ref-kind">{doc.kind}</code>
                      </label>
                    </li>
                  {/each}
                </ul>
              </div>
            {/if}
          </div>
        </div>
      </div>
    {/if}

    {#if confirmModalOpen}
      <div
        class="modal-backdrop confirm-dialog-backdrop"
        role="presentation"
        tabindex="-1"
        onclick={() => finishConfirmDialog(false)}
        onkeydown={(e) => e.key === 'Escape' && finishConfirmDialog(false)}
      >
        <div
          class="modal-panel confirm-dialog-panel"
          role="alertdialog"
          aria-modal="true"
          aria-labelledby="confirm-dialog-title"
          aria-describedby="confirm-dialog-desc"
          tabindex="-1"
          onclick={(e) => e.stopPropagation()}
          onkeydown={(e) => {
            if (e.key === 'Escape') {
              e.preventDefault()
              finishConfirmDialog(false)
            }
            e.stopPropagation()
          }}
        >
          <header class="modal-header">
            <h2 id="confirm-dialog-title" class="modal-title">{confirmModalTitle}</h2>
          </header>
          <div class="modal-body">
            <p id="confirm-dialog-desc" class="confirm-dialog-message">{confirmModalMessage}</p>
            <div class="modal-footer-row confirm-dialog-actions">
              <button type="button" class="btn-ghost" onclick={() => finishConfirmDialog(false)}>
                {confirmModalCancelLabel}
              </button>
              <button
                type="button"
                class="btn-secondary"
                class:btn-confirm-danger={confirmModalDanger}
                onclick={() => finishConfirmDialog(true)}
              >
                {#if confirmModalDanger}
                  <ButtonLeadingIcon name="danger" />
                {:else}
                  <ButtonLeadingIcon name="confirmOk" />
                {/if}
                {confirmModalConfirmLabel}
              </button>
            </div>
          </div>
        </div>
      </div>
    {/if}
  </section>
  {/if}
</main>

<style>
  .shell {
    display: flex;
    flex-direction: column;
    height: 100%;
    max-height: 100%;
    min-height: 0;
    overflow: hidden;
    background: radial-gradient(circle at top left, var(--shell-gradient-a), transparent 55%),
      radial-gradient(circle at bottom right, var(--shell-gradient-b), transparent 55%);
  }

  .shell-topbar {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-shrink: 0;
    padding: 0.65rem 0.85rem 0.5rem;
    border-bottom: 1px solid var(--color-border-subtle);
    background: linear-gradient(160deg, var(--sidebar-bg-start), var(--sidebar-bg-end));
    color: var(--sidebar-text);
  }

  .shell-topbar-brand {
    display: flex;
    align-items: center;
    gap: 0.65rem;
    flex-shrink: 0;
  }

  .shell-tabs {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    flex: 1;
    min-width: 0;
    overflow-x: auto;
    scrollbar-width: none;
  }

  .shell-tabs::-webkit-scrollbar {
    display: none;
  }

  .shell-tab {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    flex-shrink: 0;
    padding: 0.45rem 0.7rem;
    border: 1px solid transparent;
    border-radius: 0.65rem;
    background: transparent;
    color: var(--sidebar-text-muted);
    font-size: 0.8rem;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.15s ease, border-color 0.15s ease, color 0.15s ease;
  }

  .shell-tab:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--sidebar-text);
  }

  .shell-tab-active {
    background: linear-gradient(135deg, var(--sidebar-nav-active-start), var(--sidebar-nav-active-end));
    border-color: var(--sidebar-border);
    color: var(--sidebar-text);
  }

  .shell-topbar-actions {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    flex-shrink: 0;
  }

  .shell-main {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .shell-login {
    display: block;
    height: 100vh;
    max-height: none;
    min-height: 100vh;
    overflow: auto;
  }

  /* Login: soft dark (TranMail icy-blue accent — never harsh black) */
  .tm-login-shell {
    width: 100%;
    min-height: 100dvh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
    color-scheme: dark;
    background:
      radial-gradient(circle at 18% 14%, rgba(56, 189, 248, 0.14), transparent 45%),
      radial-gradient(circle at 85% 88%, rgba(37, 99, 235, 0.12), transparent 48%),
      linear-gradient(165deg, #1a2636 0%, #1f354a 52%, #1a2839 100%);
  }

  :global([data-theme='dark']) .tm-login-shell {
    background:
      radial-gradient(circle at 18% 14%, rgba(96, 165, 250, 0.12), transparent 45%),
      radial-gradient(circle at 85% 78%, rgba(30, 58, 138, 0.22), transparent 50%),
      linear-gradient(165deg, #151f2e 0%, #1c2d42 55%, #172536 100%);
  }

  .tm-login-card {
    width: 100%;
    max-width: 24rem;
    border-radius: 1rem;
    border: 1px solid rgba(106, 174, 204, 0.38);
    background: rgba(28, 42, 58, 0.94);
    padding: 2rem;
    box-shadow: 0 16px 42px rgba(8, 14, 24, 0.42);
    backdrop-filter: blur(10px);
  }

  :global([data-theme='dark']) .tm-login-card {
    border-color: rgba(96, 165, 250, 0.22);
    background: rgba(22, 34, 48, 0.94);
    box-shadow: 0 18px 44px rgba(0, 0, 0, 0.48);
  }

  .tm-login-brand {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
  }

  .tm-login-icon {
    font-size: 1.5rem;
    line-height: 1;
  }

  .tm-login-title {
    margin: 0;
    font-size: 1.125rem;
    font-weight: 600;
    color: #eaf4fb;
  }

  :global([data-theme='dark']) .tm-login-title {
    color: #f1f7fc;
  }

  .tm-login-sub {
    margin: 0;
    font-size: 0.75rem;
    color: #93adc4;
  }

  :global([data-theme='dark']) .tm-login-sub {
    color: #8ea9bf;
  }

  .tm-login-form {
    display: grid;
    gap: 1rem;
  }

  .tm-login-field {
    display: grid;
    gap: 0.375rem;
  }

  .tm-login-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: #b8cfe0;
  }

  :global([data-theme='dark']) .tm-login-label {
    color: #c5dae9;
  }

  .tm-login-input {
    width: 100%;
    border-radius: 0.5rem;
    border: 1px solid rgba(106, 163, 191, 0.45);
    background: #1e3044;
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
    color: #f0f7fc;
    outline: none;
    transition:
      border-color 0.15s ease,
      box-shadow 0.15s ease;
    box-sizing: border-box;
  }

  .tm-login-input::placeholder {
    color: #769cb5;
  }

  .tm-login-input:focus {
    border-color: #38bdf8;
    box-shadow: 0 0 0 3px rgba(56, 189, 248, 0.28);
  }

  :global([data-theme='dark']) .tm-login-input {
    border-color: rgba(96, 165, 250, 0.28);
    background: #182738;
    color: #f1f5f9;
  }

  :global([data-theme='dark']) .tm-login-input:focus {
    border-color: #60a5fa;
    box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.22);
  }

  .tm-login-remember {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: #b8cfe0;
    cursor: pointer;
    user-select: none;
  }

  .tm-login-remember input {
    margin: 0;
    accent-color: #0284c7;
  }

  :global([data-theme='dark']) .tm-login-remember {
    color: #c5dae9;
  }

  .tm-login-error {
    border-radius: 0.5rem;
    border: 1px solid rgba(248, 113, 113, 0.45);
    background: rgba(88, 28, 28, 0.35);
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
    color: #fecaca;
  }

  :global([data-theme='dark']) .tm-login-error {
    border-color: rgba(248, 113, 113, 0.42);
    background: rgba(80, 24, 24, 0.38);
    color: #fecaca;
  }

  .tm-login-submit {
    margin-top: 0.25rem;
    width: 100%;
    border: none;
    border-radius: 0.5rem;
    padding: 0.625rem 1rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #fff;
    background: #0284c7;
    cursor: pointer;
    box-shadow: 0 1px 2px rgba(15, 23, 42, 0.08);
    transition: background 0.15s ease;
  }

  .tm-login-submit:hover:not(:disabled) {
    background: #0369a1;
  }

  .tm-login-submit:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .shell-sidebar {
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow-y: auto;
    padding: 1.2rem 1rem;
    border-right: 1px solid var(--color-border-subtle);
    backdrop-filter: blur(16px);
    background: linear-gradient(160deg, var(--sidebar-bg-start), var(--sidebar-bg-end));
    color: var(--sidebar-text);
  }

  .sidebar-header {
    display: flex;
    align-items: center;
    gap: 0.65rem;
    margin-bottom: 1.05rem;
  }

  .brand-mark {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .brand-glyph {
    width: 2rem;
    height: 2rem;
    border-radius: 0.7rem;
    overflow: hidden;
    box-shadow: 0 10px 24px rgba(15, 23, 42, 0.35);
  }

  .brand-glyph-svg {
    width: 100%;
    height: 100%;
    display: block;
  }

  .brand-copy {
    display: flex;
    flex-direction: column;
  }

  .brand-title {
    font-size: 0.94rem;
    font-weight: 600;
    letter-spacing: 0.01em;
    text-transform: none;
  }

  .brand-subtitle {
    font-size: 0.7rem;
    color: var(--sidebar-text-muted);
  }

  .nav {
    display: flex;
    flex-direction: column;
    gap: 0.16rem;
    margin-bottom: auto;
  }

  .nav-item {
    display: flex;
    width: 100%;
    justify-content: flex-start;
    align-items: center;
    gap: 0.3rem;
    padding: 0.36rem 0.56rem;
    border-radius: 999px;
    background: transparent;
    color: var(--sidebar-text);
    border: 1px solid transparent;
    font-size: 0.76rem;
    box-shadow: none;
  }

  .nav-item span {
    flex: 0 1 auto;
    text-align: left;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    line-height: 1.1;
  }

  .nav-item-icon {
    width: 0.78rem;
    height: 0.78rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    opacity: 0.92;
    flex: 0 0 0.78rem;
    margin-right: 0.06rem;
  }

  .nav-icon-svg {
    width: 0.78rem;
    height: 0.78rem;
    display: block;
  }

  .nav-item-nested {
    margin-left: 0.55rem;
    padding-left: 0.75rem;
    border-left: 2px solid var(--sidebar-border);
    border-radius: 0 999px 999px 0;
    font-size: 0.72rem;
    opacity: 0.96;
  }

  .nav-item:hover {
    border-color: var(--sidebar-border);
    background: rgba(255, 255, 255, 0.06);
    box-shadow: none;
  }

  .nav-item-active {
    background: linear-gradient(135deg, var(--sidebar-nav-active-start), var(--sidebar-nav-active-end));
    border-color: var(--sidebar-border);
    color: #ffffff;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.45);
  }

  .sidebar-footer {
    padding-top: 1.5rem;
    border-top: 1px solid var(--sidebar-border);
    font-size: 0.75rem;
    color: var(--sidebar-text-muted);
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .theme-toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 0.45rem 0.75rem;
    border-radius: 0.65rem;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid var(--sidebar-border);
    color: var(--sidebar-text);
    font-size: 0.78rem;
    font-weight: 500;
    box-shadow: none;
  }

  .theme-toggle:hover {
    background: rgba(255, 255, 255, 0.12);
    box-shadow: none;
  }

  .theme-toggle-hint {
    font-size: 1rem;
    opacity: 0.9;
  }

  .badge.beta {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.2rem 0.55rem;
    border-radius: 999px;
    background: var(--badge-bg);
    border: 1px solid var(--badge-border);
    color: var(--badge-text);
    font-size: 0.7rem;
    width: fit-content;
  }

  .shell-main {
    padding: 1.25rem 1.75rem 1.25rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
  }

  .shell-main-fill {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
    overflow: hidden;
  }

  .toast-stack {
    position: fixed;
    top: 1rem;
    right: 1rem;
    z-index: 400;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.5rem;
    max-width: min(22rem, calc(100vw - 2rem));
    pointer-events: none;
  }

  .toast {
    pointer-events: auto;
    margin: 0;
    padding: 0.55rem 0.95rem;
    border-radius: 0.65rem;
    font-size: 0.8125rem;
    line-height: 1.4;
    box-shadow:
      0 4px 14px rgba(15, 23, 42, 0.12),
      0 0 0 1px var(--color-border-subtle);
    background: var(--color-surface);
    color: var(--color-text);
  }

  .toast-error {
    background: var(--color-danger-soft);
    border: 1px solid var(--color-danger-border);
    color: var(--color-danger);
    box-shadow:
      0 4px 14px rgba(15, 23, 42, 0.1),
      0 0 0 1px var(--color-danger-border);
  }

  .toast-info {
    background: var(--color-primary-soft);
    border: 1px solid var(--flash-info-border);
    color: var(--color-primary);
    box-shadow:
      0 4px 14px rgba(15, 23, 42, 0.1),
      0 0 0 1px var(--flash-info-border);
  }

  .main-header {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .main-header h1 {
    margin: 0;
    font-size: 1.4rem;
    letter-spacing: 0.01em;
  }

  .page-subtitle {
    margin: 0.35rem 0 0;
    font-size: 0.85rem;
    color: var(--color-text-subtle);
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  /* Button variants aligned to theme */
  .primary-action {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.42rem;
    background: linear-gradient(135deg, var(--color-primary-muted), var(--color-primary));
    border-color: rgba(30, 64, 175, 0.55);
    color: #ffffff;
  }

  .primary-action:hover {
    background: linear-gradient(135deg, var(--color-primary), var(--color-primary-hover));
  }

  :global(.btn-leading-icon) {
    flex-shrink: 0;
  }

  .btn-secondary,
  .btn-file {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.42rem;
    border-radius: 999px;
    padding: 0.55rem 1.1rem;
    font-size: 0.9rem;
    font-weight: 500;
    border-width: 1px;
    border-style: solid;
    background: linear-gradient(135deg, var(--color-secondary), var(--color-secondary-hover));
    color: #ffffff;
    border-color: rgba(37, 99, 235, 0.45);
    box-shadow: 0 8px 22px rgba(30, 64, 175, 0.22);
  }

  .btn-secondary:hover,
  .btn-file:hover {
    filter: brightness(1.06);
  }

  .btn-ghost {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 999px;
    padding: 0.5rem 1rem;
    background: transparent;
    color: var(--color-text-subtle);
    border-color: transparent;
    box-shadow: none;
  }

  .btn-ghost:hover {
    background: var(--color-primary-soft);
  }

  .page-grid {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    grid-auto-rows: minmax(0, 1fr);
    gap: 1.1rem;
    align-items: stretch;
  }

  .panel {
    background-color: var(--color-surface);
    border-radius: 1.2rem;
    border: 1px solid var(--color-border-subtle);
    box-shadow: var(--shadow-soft);
    padding: 1rem 1.1rem 1rem;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .panel-wide {
    grid-column: 1 / -1;
  }

  .panel-header {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
  }

  .panel-header h2 {
    margin: 0;
    font-size: 0.95rem;
  }

  .panel-meta {
    display: block;
    margin-top: 0.15rem;
    font-size: 0.75rem;
    color: var(--color-text-subtle);
  }

  .panel-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    font-size: 0.85rem;
  }

  .table-scroll {
    flex: 1;
    min-height: 0;
    overflow: auto;
    border-radius: 0.65rem;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg);
  }

  .grid {
    width: 100%;
    font-size: 0.8rem;
  }

  .grid th,
  .grid td {
    padding: 0.55rem 0.55rem;
    text-align: left;
  }

  .grid thead th {
    position: sticky;
    top: 0;
    z-index: 2;
    font-size: 0.75rem;
    font-weight: 600;
    border-bottom: 1px solid var(--color-border-subtle);
    color: var(--color-text-subtle);
    background-color: var(--color-surface);
    box-shadow: 0 1px 0 var(--color-border-subtle);
  }

  .grid tbody tr:nth-child(even) {
    background-color: var(--color-bg-elevated);
  }

  .grid-empty td {
    text-align: center;
    color: var(--color-text-subtle);
  }

  .grid-actions {
    text-align: right;
  }

  .grid-actions-col {
    width: 6.75rem;
  }

  .icon-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.1rem;
  }

  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2.1rem;
    height: 2.1rem;
    padding: 0;
    border-radius: 0.5rem;
    background: transparent !important;
    color: var(--color-text-subtle);
    border: 1px solid transparent;
    box-shadow: none !important;
  }

  .icon-btn:hover {
    background: var(--color-primary-soft) !important;
    color: var(--color-primary);
    box-shadow: none !important;
  }

  .icon-btn-danger:hover {
    background: var(--color-danger-soft) !important;
    color: var(--color-danger);
  }

  .icon-svg {
    width: 1.15rem;
    height: 1.15rem;
    display: block;
  }

  .email-row-name {
    font-weight: 500;
  }

  .email-row-date {
    font-size: 0.8rem;
    color: var(--color-text-subtle);
    white-space: nowrap;
  }

  .modal-backdrop {
    position: fixed;
    inset: 0;
    z-index: 100;
    background: rgba(15, 23, 42, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
  }

  .confirm-dialog-backdrop {
    z-index: 200;
    background: rgba(15, 23, 42, 0.58);
  }

  .confirm-dialog-panel {
    max-width: 26rem;
    width: calc(100vw - 2rem);
  }

  .confirm-dialog-message {
    margin: 0;
    font-size: 0.9rem;
    line-height: 1.5;
    color: var(--color-text);
    white-space: pre-wrap;
  }

  .btn-secondary.btn-confirm-danger:not(:disabled) {
    color: var(--color-danger);
    border-color: var(--color-danger-border);
  }

  .btn-secondary.btn-confirm-danger:not(:disabled):hover {
    background: var(--color-danger-soft);
    color: var(--color-danger);
  }

  .modal-panel {
    background: var(--color-surface);
    border-radius: 1rem;
    border: 1px solid var(--color-border-subtle);
    max-width: 44rem;
    width: 100%;
    max-height: min(90vh, 680px);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    box-shadow: var(--shadow-soft);
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.85rem 1rem;
    border-bottom: 1px solid var(--color-border-subtle);
  }

  .modal-title {
    margin: 0;
    font-size: 1.05rem;
    font-weight: 600;
  }

  .modal-close {
    flex-shrink: 0;
  }

  .modal-body {
    padding: 1rem;
    overflow: auto;
  }

  .platform-assistant-drawer-host {
    position: fixed;
    inset: 0;
    z-index: 100;
  }

  .platform-assistant-drawer-scrim {
    position: absolute;
    inset: 0;
    margin: 0;
    padding: 0;
    border: none;
    background: rgba(2, 8, 23, 0.52);
    cursor: pointer;
  }

  .platform-assistant-drawer {
    position: absolute;
    right: 0;
    top: 0;
    bottom: 0;
    width: min(440px, 100vw);
    background: var(--color-surface);
    border-left: 1px solid var(--color-border-subtle);
    border-radius: 0;
    max-width: none;
    max-height: none;
    box-shadow: 0 12px 48px rgba(0, 0, 0, 0.18);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .platform-assistant-drawer-header {
    flex-shrink: 0;
  }

  .platform-assistant-drawer-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
    padding: 1rem;
    overflow: hidden;
  }

  .platform-assistant-chat-log {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }

  .platform-assistant-compose {
    margin-top: 0;
    flex-shrink: 0;
  }

  .platform-assistant-input {
    min-width: 0;
  }

  .detail-meta {
    margin: 0 0 0.65rem;
    font-size: 0.85rem;
    color: var(--color-text-subtle);
  }

  .mono {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }

  .detail-html-label {
    font-size: 0.75rem;
    font-weight: 600;
    margin: 0 0 0.35rem;
    color: var(--color-text-subtle);
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .content-name-link {
    appearance: none;
    border: 0;
    background: transparent;
    padding: 0;
    margin: 0;
    font: inherit;
    color: var(--color-primary);
    text-align: left;
    cursor: pointer;
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  .content-name-link:hover {
    color: var(--color-primary-hover, var(--color-primary));
  }

  .content-detail-modal {
    width: min(960px, calc(100vw - 2rem));
    max-width: min(960px, calc(100vw - 2rem));
    max-height: min(92vh, 820px);
  }

  .content-detail-header {
    flex-wrap: wrap;
    gap: 0.65rem;
  }

  .content-detail-header-actions {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    flex-wrap: wrap;
    margin-left: auto;
  }

  .content-detail-tabs {
    display: inline-flex;
    align-items: center;
    gap: 0.2rem;
    padding: 0.15rem;
    border-radius: 999px;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg);
  }

  .content-detail-tab {
    appearance: none;
    border: 0;
    background: transparent;
    color: var(--color-text-subtle);
    font: inherit;
    font-size: 0.72rem;
    font-weight: 600;
    padding: 0.28rem 0.65rem;
    border-radius: 999px;
    cursor: pointer;
  }

  .content-detail-tab-active {
    background: var(--color-primary-soft);
    color: var(--color-primary);
  }

  .content-detail-download {
    font-size: 0.72rem;
    padding: 0.32rem 0.6rem;
  }

  .content-detail-body {
    min-height: 0;
    overflow: auto;
  }

  .detail-html {
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.75rem;
    padding: 0.75rem;
    background: var(--color-bg);
    max-height: 320px;
    overflow: auto;
    font-size: 0.85rem;
    line-height: 1.45;
  }

  .library-template-modal {
    max-width: min(72rem, 96vw);
    width: 100%;
    min-height: min(65vh, 640px);
    max-height: min(92vh, 920px);
  }

  .library-template-modal .modal-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
  }

  .library-form-html {
    margin-top: 0.25rem;
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .library-html-textarea {
    width: 100%;
    min-height: 12rem;
    resize: vertical;
    border-radius: 0.75rem;
    border: 1px solid var(--color-border-subtle);
    padding: 0.55rem 0.65rem;
    font-size: 0.8rem;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    background: var(--color-bg);
    color: var(--color-text);
  }

  .composer-ai-rail {
    order: 2;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-height: 0;
    min-width: 0;
    height: 100%;
    max-height: 100%;
    padding: 0.55rem 0.65rem;
    border-radius: 0.85rem;
    border: 1px dashed var(--color-border-subtle);
    background: var(--color-bg-elevated);
    overflow: hidden;
  }

  .composer-ai-rail-head {
    flex: 0 0 auto;
  }

  .composer-ai-rail-title {
    margin: 0;
    font-size: 0.82rem;
    font-weight: 650;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--color-text-subtle);
  }

  .composer-ai-ref-bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 0.4rem 0.75rem;
    flex-shrink: 0;
  }

  .composer-ai-ref-docs-btn {
    font-size: 0.74rem;
    padding: 0.28rem 0.55rem;
  }

  .composer-ai-ref-count {
    margin: 0;
    flex: 1 1 auto;
    text-align: right;
    font-size: 0.72rem;
    white-space: nowrap;
  }

  .composer-ai-web-toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.72rem;
    color: var(--color-text-subtle);
    user-select: none;
  }

  .composer-ai-web-toggle input {
    margin: 0;
  }

  .ai-ref-docs-modal {
    max-width: min(26rem, calc(100vw - 2rem));
  }

  .ai-ref-docs-modal-body {
    max-height: min(60vh, 24rem);
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
  }

  .composer-ai-ref-modal-hint {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.35rem 0.65rem;
  }

  .ai-ref-picker-scroll {
    max-height: min(48vh, 18rem);
    overflow-y: auto;
  }

  .composer-ai-chat-log {
    flex: 1 1 0;
    min-height: 0;
    max-height: none;
  }

  .composer-ai-chat-empty {
    flex: 1 1 0;
    min-height: 0;
  }

  .composer-ai-compose {
    flex: 0 0 auto;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .composer-ai-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 0.15rem;
  }

  .ai-response-time {
    text-align: right;
    color: var(--color-accent, #8b5cf6);
    font-size: 0.7rem;
    font-variant-numeric: tabular-nums;
    opacity: 0.9;
  }

  .ai-ref-pick {
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.75rem;
    padding: 0.5rem 0.65rem;
    max-height: 9.5rem;
    overflow-y: auto;
    background: var(--color-bg);
  }

  .ai-ref-list {
    list-style: none;
    margin: 0.35rem 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .ai-ref-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .ai-ref-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .ai-ref-kind {
    flex-shrink: 0;
    font-size: 0.65rem;
  }

  .ai-chat-end-anchor {
    flex-shrink: 0;
    height: 1px;
    overflow: hidden;
    pointer-events: none;
    margin: 0;
    align-self: stretch;
  }

  .ai-chat-log {
    flex: 1 1 0;
    min-height: 0;
    overflow-y: auto;
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.75rem;
    padding: 0.65rem;
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
    background: var(--color-bg);
  }

  .ai-chat-bubble {
    border-radius: 0.65rem;
    padding: 0.5rem 0.65rem;
    font-size: 0.85rem;
    line-height: 1.45;
  }

  .ai-chat-user {
    align-self: flex-end;
    background: rgba(59, 130, 246, 0.14);
    max-width: 92%;
  }

  .ai-chat-asst {
    align-self: stretch;
    background: rgba(15, 23, 42, 0.06);
  }

  :global([data-theme='dark']) .ai-chat-asst {
    background: rgba(255, 255, 255, 0.07);
  }

  .ai-chat-role {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    opacity: 0.75;
    margin-bottom: 0.25rem;
  }

  .ai-chat-text {
    white-space: pre-wrap;
    word-break: break-word;
  }

  .ai-chat-apply-row {
    margin-top: 0.5rem;
  }

  .ai-chat-compose {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    flex-shrink: 0;
  }

  .ai-chat-input {
    width: 100%;
    resize: vertical;
    min-height: 4.5rem;
    border-radius: 0.75rem;
    border: 1px solid var(--color-border-subtle);
    padding: 0.55rem 0.65rem;
    font-size: 0.85rem;
    font-family: inherit;
    background: var(--color-bg);
    color: var(--color-text);
    box-sizing: border-box;
  }

  .composer-ai-chat-input {
    flex-shrink: 0;
    min-height: 5rem;
    max-height: 12rem;
  }

  .modal-footer-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 1rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--color-border-subtle);
  }

  .modal-footer-row.confirm-dialog-actions {
    margin-top: 1.15rem;
    padding-top: 0;
    border-top: none;
  }

  .modal-delete-btn {
    color: var(--color-danger);
  }

  .modal-delete-btn:hover {
    background: var(--color-danger-soft) !important;
  }

  .field-hint {
    font-weight: 400;
    color: var(--color-text-subtle);
    font-size: 0.72rem;
  }

  .modal-footer-actions {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .btn-secondary.btn-delete-selected:not(:disabled) {
    color: var(--color-danger);
    border-color: var(--color-danger-border);
  }

  .btn-secondary.btn-delete-selected:not(:disabled):hover {
    background: var(--color-danger-soft);
  }

  .panel-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .btn-file input[type='file'] {
    display: none;
  }

  .composer {
    padding: 0.9rem 1rem 0.9rem;
    flex: 1;
    min-height: 0;
  }

  .composer-main {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
  }

  .composer-body {
    display: grid;
    grid-template-columns: minmax(0, 2fr) minmax(260px, 1fr);
    gap: 0.9rem;
    min-height: 0;
    align-items: stretch;
  }

  .composer-editor-col {
    order: 1;
    display: flex;
    flex-direction: column;
    gap: 0;
    min-height: 0;
    min-width: 0;
  }

  .composer-editor-shell {
    flex: 1;
    min-height: 0;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .editor-body-host {
    position: relative;
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .editor-toolbar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.45rem 0.75rem;
    padding: 0.5rem 0.7rem;
    border-bottom: 1px solid var(--color-border-subtle);
    background: linear-gradient(90deg, var(--color-primary-soft), transparent);
  }

  .composer-assets-hint {
    margin: 0;
    font-size: 0.72rem;
  }

  .composer-assets-block + .composer-assets-block {
    margin-top: 0.45rem;
  }

  .composer-assets-meta {
    display: flex;
    align-items: baseline;
    gap: 0.45rem;
    margin-bottom: 0.25rem;
  }

  .composer-dataset-name {
    font-size: 0.74rem;
    font-weight: 500;
    color: var(--color-text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .composer-token-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
  }

  .composer-field-token {
    font-size: 0.68rem;
    padding: 0.2rem 0.45rem;
  }

  .composer-main > .composer-body:first-of-type {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .field-group {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }

  .field-group label {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-text-subtle);
  }

  .field-group input {
    border-radius: 0.8rem;
    border: 1px solid var(--color-border-subtle);
    padding: 0.55rem 0.75rem;
    font-size: 0.85rem;
    background-color: var(--color-bg);
    color: var(--color-text);
  }

  .composer-name-group {
    min-width: min(42rem, 68vw);
    max-width: min(48rem, 72vw);
    flex: 1;
  }

  .composer-name-input {
    width: 100%;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .editor {
    display: flex;
    flex-direction: column;
    border-radius: 0.9rem;
    border: 1px solid var(--color-border-subtle);
    background-color: var(--color-bg);
    overflow: hidden;
    min-height: 0;
    flex: 1;
  }

  .composer-toolbar-pickers {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.65rem 1rem;
    margin-left: auto;
    justify-content: flex-end;
  }

  .starter-picker {
    display: flex;
    align-items: center;
    gap: 0.45rem;
  }

  .starter-picker-label {
    font-size: 0.72rem;
    font-weight: 500;
    color: var(--color-text-subtle);
    white-space: nowrap;
  }

  .starter-select {
    font-size: 0.78rem;
    padding: 0.35rem 0.55rem;
    border-radius: 0.55rem;
    border: 1px solid var(--color-border-subtle);
    background-color: var(--color-bg);
    color: var(--color-text);
    min-width: 10.5rem;
    max-width: 100%;
    font-family: inherit;
    cursor: pointer;
  }

  .starter-select:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 1px;
  }

  .composer-template-select {
    min-width: 14rem;
    max-width: min(28rem, 92vw);
  }

  .template-tag-cell {
    vertical-align: middle;
    white-space: nowrap;
  }

  .template-tag {
    font-size: 0.72rem;
    padding: 0.12rem 0.35rem;
    border-radius: 0.35rem;
    background: var(--color-primary-soft);
    color: var(--color-text);
    border: 1px solid var(--color-border-subtle);
  }

  .template-builtin-badge {
    display: inline-block;
    margin-left: 0.35rem;
    font-size: 0.65rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-subtle);
  }

  .template-desc-cell {
    font-size: 0.78rem;
    color: var(--color-text-subtle);
    max-width: 22rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: middle;
  }

  .muted-cell {
    color: var(--color-text-subtle);
    font-size: 0.8rem;
  }

  .library-meta-row {
    margin-bottom: 0.65rem;
  }

  .library-description-text {
    margin: 0.25rem 0 0;
    font-size: 0.85rem;
    line-height: 1.45;
    color: var(--color-text);
  }

  .library-desc-textarea {
    width: 100%;
    font-family: inherit;
    font-size: 0.85rem;
    padding: 0.4rem 0.5rem;
    border-radius: 0.45rem;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg);
    color: var(--color-text);
    resize: vertical;
  }

  .pill-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-text-subtle);
    text-transform: uppercase;
    letter-spacing: 0.09em;
  }

  .token-hint {
    margin: 0 0 0.35rem;
  }

  .editor-input {
    border: none;
    min-height: 0;
    flex: 1;
    width: 100%;
    overflow: auto;
    background-color: var(--color-bg);
    color: var(--color-text);
    resize: none;
    padding: 0.75rem 1rem;
    box-sizing: border-box;
  }

  .markdown-editor {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
      monospace;
    font-size: 0.875rem;
    line-height: 1.55;
    tab-size: 2;
  }

  .detail-markdown {
    margin: 0;
    padding: 0.75rem;
    border-radius: 0.45rem;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg-elevated);
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
      monospace;
    font-size: 0.85rem;
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 50vh;
    overflow: auto;
  }

  .editor-input:focus-visible {
    outline: none;
  }

  .template-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .template-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .template-list li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.35rem;
    padding: 0.25rem 0.4rem;
    border-radius: 0.5rem;
  }

  .template-list li button {
    background: transparent;
    border: none;
    padding: 0;
    margin: 0;
    box-shadow: none;
    color: var(--color-text);
    font-size: 0.78rem;
  }

  .template-list li button:hover {
    text-decoration: underline;
  }

  .template-active {
    background-color: var(--color-primary-soft);
  }

  .template-delete {
    cursor: pointer;
    font-size: 0.9rem;
    color: var(--color-text-subtle);
  }

  .sidebar-block-title {
    font-size: 0.8rem;
    font-weight: 600;
    margin-bottom: 0.3rem;
  }

  .sidebar-hint {
    font-size: 0.75rem;
    color: var(--color-text-subtle);
    margin: 0;
  }

  .token-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
  }

  .token-list-clickable {
    gap: 0.32rem;
  }

  .token-btn {
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg);
    color: var(--color-text);
    border-radius: 999px;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
      monospace;
    font-size: 0.7rem;
    padding: 0.2rem 0.45rem;
    cursor: pointer;
    line-height: 1.2;
  }

  .token-btn:hover {
    border-color: var(--flash-info-border);
    background: var(--color-primary-soft);
  }

  .merge-source-groups {
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
    margin-top: 0.2rem;
  }

  .merge-source-group {
    border: 1px dashed var(--color-border-subtle);
    border-radius: 0.65rem;
    padding: 0.4rem;
    background: var(--color-bg-elevated);
  }

  .merge-source-title {
    font-size: 0.72rem;
    font-weight: 600;
    margin-bottom: 0.28rem;
    color: var(--color-text-subtle);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .panel-tabs {
    display: flex;
    gap: 0.35rem;
    padding: 0 1rem 0.6rem;
  }

  .panel-tab {
    border-radius: 999px;
    border: 1px solid var(--color-border-subtle);
    background: var(--color-bg);
    color: var(--color-text-subtle);
    font-size: 0.75rem;
    font-weight: 600;
    padding: 0.33rem 0.7rem;
    cursor: pointer;
  }

  .panel-tab-active {
    background: var(--color-primary-soft);
    color: var(--color-primary);
    border-color: var(--flash-info-border);
  }

  .merge-field-config {
    display: flex;
    flex-direction: column;
    gap: 0.7rem;
  }

  .merge-field-source-card {
    border: 1px solid var(--color-border-subtle);
    border-radius: 0.75rem;
    padding: 0.6rem 0.65rem;
    background: var(--color-bg);
  }

  .merge-field-source-header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.75rem;
    margin-bottom: 0.45rem;
  }

  .merge-field-source-name {
    font-size: 0.83rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .merge-field-add-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 0.4rem;
    margin-bottom: 0.45rem;
  }

  .merge-field-add-row input {
    border-radius: 0.6rem;
    border: 1px solid var(--color-border-subtle);
    padding: 0.38rem 0.55rem;
    font-size: 0.78rem;
    background-color: var(--color-bg);
    color: var(--color-text);
    min-width: 0;
  }

  .merge-field-token-list {
    margin-top: 0.2rem;
  }

  .token-chip {
    display: inline-flex;
    align-items: center;
    border: 1px solid var(--color-border-subtle);
    border-radius: 999px;
    overflow: hidden;
    background: var(--color-bg);
  }

  .token-chip .token-btn {
    border: none;
    border-radius: 0;
    background: transparent;
  }

  .token-remove-btn {
    border: none;
    border-left: 1px solid var(--color-border-subtle);
    background: transparent;
    color: var(--color-text-subtle);
    font-size: 0.8rem;
    line-height: 1;
    padding: 0.16rem 0.35rem;
    cursor: pointer;
  }

  .token-remove-btn:hover {
    color: var(--color-danger);
    background: var(--color-danger-soft);
  }

  .merge-key-field {
    margin-top: 0.65rem;
    padding-top: 0.55rem;
    border-top: 1px dashed var(--flash-info-border);
  }

  .merge-key-label {
    display: block;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--color-text-subtle);
    margin-bottom: 0.2rem;
  }

  .merge-key-desc {
    margin: 0 0 0.35rem;
    font-size: 0.7rem;
    line-height: 1.35;
  }

  .merge-key-desc code {
    font-size: 0.68rem;
  }

  .merge-key-input {
    width: 100%;
    border-radius: 0.65rem;
    border: 1px solid var(--color-border-subtle);
    padding: 0.4rem 0.55rem;
    font-size: 0.78rem;
    background-color: var(--color-bg);
    color: var(--color-text);
    font-family: inherit;
    box-sizing: border-box;
  }

  .composer-autosave-status {
    margin-left: auto;
    font-size: 0.72rem;
    font-weight: 500;
    color: var(--color-text-subtle);
    min-width: 4.5rem;
    text-align: right;
  }

  .composer-editor-tabs {
    margin-left: auto;
  }

  .composer-preview-host {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 0.75rem 1rem;
  }

  .composer-preview-host :global(.content-markdown-preview) {
    max-height: none;
    border: none;
    padding: 0;
    background: transparent;
  }

  .composer-preview-host :global(.content-markdown-empty) {
    font-size: 0.85rem;
  }

  @media (max-width: 960px) {
    .shell-topbar {
      flex-wrap: wrap;
      padding-bottom: 0.45rem;
    }

    .shell-tabs {
      order: 3;
      flex-basis: 100%;
      padding-top: 0.15rem;
    }

    .shell-tab {
      font-size: 0.74rem;
      padding: 0.4rem 0.55rem;
    }

    .shell-main,
    .main {
      min-height: 0;
      overflow: auto;
    }
  }

  @media (max-width: 720px) {
    .main-header {
      flex-direction: column;
      align-items: flex-start;
    }

    .page-grid {
      grid-template-columns: minmax(0, 1fr);
    }

    .composer-body {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>
