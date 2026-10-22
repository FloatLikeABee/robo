import { mount } from 'svelte'
import './app.css'
import App from './App.svelte'
import { installLocaleListener } from './lib/locale.js'

installLocaleListener()

const app = mount(App, { target: document.getElementById('app')! })
export default app
