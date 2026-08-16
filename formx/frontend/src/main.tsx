import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { setAuthToken } from './lib/api'

/** Morph Utils / Morph AI handoff: ?userspanel_token=… → shared session cookie. */
function consumePlatformTokenFromUrl() {
  try {
    const params = new URLSearchParams(window.location.search)
    const token = (params.get('userspanel_token') || '').trim()
    if (!token) return
    setAuthToken(token, { rememberMe: true })
    params.delete('userspanel_token')
    const url = new URL(window.location.href)
    url.search = params.toString()
    window.history.replaceState({}, '', url.toString())
  } catch {
    /* ignore */
  }
}

consumePlatformTokenFromUrl()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
