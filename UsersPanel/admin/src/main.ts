import { mount } from 'svelte'
import './app.css'
import { applyTheme, readTheme } from './lib/theme'
import App from './App.svelte'

applyTheme(readTheme())

if (typeof window !== 'undefined' && (!window.location.hash || window.location.hash === '#')) {
  window.location.hash = '#/login'
}

const app = mount(App, {
  target: document.getElementById('app')!,
})

export default app
