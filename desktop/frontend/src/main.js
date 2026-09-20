import './style.css'
import {
  Session,
  Login,
  Logout,
  Inspect,
  Enqueue,
  Jobs,
  History,
  Cancel,
  Pause,
  Resume,
  Retry,
  RemoveJob,
  RemoveSeries,
  OpenPath,
  RevealPath,
  PreviewPath,
  OpenSeriesFolder,
  Settings,
  SaveSettings,
  PickFolder,
  LibrarySeries,
  LibrarySeasons,
  LibraryEpisodes,
  AppVersion,
  CheckForUpdates,
  OpenURL,
} from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { t, setLang, getLang, LANGS, statusLabel } from './i18n.js'
import { switchControl, switchChip, appleSelect, bindAppleSelects } from './ui.js'

/* Iconsax Linear — width/height set via CSS; viewBox keeps aspect */
const ICONS = {
  home: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M9.02 2.84l-5.39 4.2C2.73 7.74 2 9.23 2 10.36v7.41c0 2.32 1.89 4.22 4.21 4.22h11.58c2.32 0 4.21-1.9 4.21-4.21V10.5c0-1.21-.81-2.76-1.8-3.45l-6.18-4.33c-1.4-.98-3.65-.93-5 .12z"/><path d="M12 17.99v-3"/></svg>`,
  queue: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M9 17.5H5c-2 0-3-1-3-3v-7c0-2 1-3 3-3h14c2 0 3 1 3 3v7c0 2-1 3-3 3h-4"/><path d="M12 15v6M12 21l-2-2M12 21l2-2"/><path d="M7 9.5h10M7 13h6"/></svg>`,
  history: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 8v5l3 2"/><path d="M3.05 11a9 9 0 1 1 .5 4"/><path d="M3 16.5V11h5.5"/></svg>`,
  settings: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09a1.65 1.65 0 0 0 1.51-1 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9c.26.6.87 1 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>`,
  back: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M14.5 5.5L21 12l-6.5 6.5"/><path d="M21 12H3"/></svg>`,
  backLtr: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M9.5 5.5L3 12l6.5 6.5"/><path d="M3 12H21"/></svg>`,
  folder: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M22 11v6c0 4-1 5-5 5H7c-4 0-5-1-5-5V7c0-4 1-5 5-5h1.5c1.5 0 2.1.63 2.4 1.2l.7 1.5c.15.3.45.8.9.8H17c4 0 5 1 5 5z"/></svg>`,
  pause: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M10.65 19.11V4.89c0-1.35-.57-1.89-2.01-1.89H5.01C3.57 3 3 3.54 3 4.89v14.22C3 20.46 3.57 21 5.01 21h3.63c1.44 0 2.01-.54 2.01-1.89zM21 19.11V4.89C21 3.54 20.43 3 18.99 3h-3.63c-1.43 0-2.01.54-2.01 1.89v14.22c0 1.35.57 1.89 2.01 1.89h3.63C20.43 21 21 20.46 21 19.11z"/></svg>`,
  play: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 12V8.44c0-4.42 3.13-6.23 6.96-4.02l3.09 1.78 3.09 1.78c3.83 2.21 3.83 5.83 0 8.04l-3.09 1.78-3.09 1.78C7.13 21.79 4 19.98 4 15.56V12z"/></svg>`,
  refresh: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M22 12c0 5.52-4.48 10-10 10s-8.89-5.56-8.89-5.56m0 0h4.52m-4.52 0v5"/><path d="M2 12c0-5.52 4.44-10 10-10 7.11 0 10 5.56 10 5.56m0 0v-5m0 5h-4.44"/></svg>`,
  close: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 22c5.5 0 10-4.5 10-10S17.5 2 12 2 2 6.5 2 12s4.5 10 10 10z"/><path d="M9.17 14.83l5.66-5.66M14.83 14.83L9.17 9.17"/></svg>`,
  chevron: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M15 19.92L8.48 13.4c-.77-.77-.77-2.03 0-2.8L15 4.08"/></svg>`,
  trash: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M21 5.98c-3.33-.33-6.68-.5-10.02-.5-1.98 0-3.96.1-5.94.3L3 5.98"/><path d="M8.5 4.97l.22-1.31C8.88 2.71 9 2 10.44 2h3.12c1.44 0 1.57.75 1.72 1.67l.22 1.3"/><path d="M18.85 9.14l-.65 10.07C18.09 20.78 18 22 15.21 22H8.79C6 22 5.91 20.78 5.8 19.21L5.15 9.14"/><path d="M10.33 16.5h3.33M9.5 12.5h5"/></svg>`,
  search: `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M11.5 21a9.5 9.5 0 1 0 0-19 9.5 9.5 0 0 0 0 19z"/><path d="M22 22l-2-2"/></svg>`,
}

const state = {
  view: 'library',
  session: { loggedIn: false, username: '', hasToken: false },
  jobs: [],
  history: [],
  settings: {},
  library: [],
  seasons: [],
  episodes: [],
  drill: { level: 'root', seriesKey: '', seriesTitle: '', season: 0, seasonTitle: '' },
  query: '',
  libSearch: '',
  filter: 'all',
  inspect: null,
  inspectOpen: false,
  busy: false,
  error: '',
  selectedEps: [],
  version: '',
  update: null,
  toast: '',
  pendingRender: false,
}

let lastPointerAt = 0

function $(sel, root = document) {
  return root.querySelector(sel)
}

function errText(err) {
  if (!err) return ''
  if (typeof err === 'string') return err
  return err.message || String(err)
}

function needsLogin() {
  return !(state.session && (state.session.loggedIn || state.session.hasToken))
}

function counts() {
  const jobs = state.jobs || []
  return {
    queued: jobs.filter((j) => j.status === 'queued').length,
    running: jobs.filter((j) => j.status === 'running' || j.status === 'preparing').length,
    paused: jobs.filter((j) => j.status === 'paused').length,
    done: jobs.filter((j) => j.status === 'done').length,
    error: jobs.filter((j) => j.status === 'error' || j.status === 'canceled').length,
  }
}

function escapeHtml(value) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function escapeAttr(value) {
  return escapeHtml(value)
}

function pctSymbol() {
  const lang = getLang()
  return (lang === 'fa' || lang === 'ar') ? '٪' : '%'
}

function backIcon() {
  return document.documentElement.dir === 'rtl' ? ICONS.back : ICONS.backLtr
}

function coverHtml(cover, title, cls = 'thumb') {
  const initial = (title || '?').trim().charAt(0)
  if (cover) {
    return `<img class="${cls}" src="${escapeAttr(cover)}" alt="" loading="lazy" referrerpolicy="no-referrer" decoding="async">`
  }
  return `<div class="${cls} placeholder" aria-hidden="true">${escapeHtml(initial)}</div>`
}

function applyTheme(theme) {
  const pref = theme || 'system'
  let resolved = pref
  if (pref === 'system') {
    resolved = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  document.documentElement.setAttribute('data-theme', resolved)
  document.documentElement.style.colorScheme = resolved
}

function matchFilter(status) {
  const f = state.filter
  if (f === 'all') return true
  if (f === 'active') return status === 'queued' || status === 'preparing' || status === 'running' || status === 'paused'
  if (f === 'done') return status === 'done'
  if (f === 'error') return status === 'error' || status === 'canceled'
  return true
}

function textMatch(hay, needle) {
  if (!needle) return true
  return String(hay || '').toLowerCase().includes(String(needle).toLowerCase())
}

function filteredLibrary() {
  return (state.library || []).filter((item) => {
    if (!textMatch(item.title, state.libSearch)) return false
    if (state.filter === 'all') return true
    if (state.filter === 'active') return (item.activeCount || 0) > 0
    if (state.filter === 'done') return (item.doneCount || 0) > 0 && (item.activeCount || 0) === 0
    if (state.filter === 'error') {
      const jobs = (state.jobs || []).filter((j) => {
        const key = j.seriesKey || `movie:${j.contentId}`
        return key === item.key && (j.status === 'error' || j.status === 'canceled')
      })
      return jobs.length > 0
    }
    return true
  })
}

function filteredEpisodes() {
  return (state.episodes || []).filter((job) => {
    if (!matchFilter(job.status)) return false
    const label = `${job.title || ''} ${job.episode || ''}`
    return textMatch(label, state.libSearch)
  })
}

function filteredQueue() {
  return (state.jobs || []).filter((j) => {
    if (!(j.status === 'queued' || j.status === 'preparing' || j.status === 'running' || j.status === 'paused' || j.status === 'error')) return false
    if (!matchFilter(j.status) && state.filter !== 'all') {
      if (state.filter === 'active') return j.status === 'queued' || j.status === 'preparing' || j.status === 'running' || j.status === 'paused'
      return matchFilter(j.status)
    }
    return textMatch(j.title || j.seriesTitle, state.libSearch)
  })
}

async function refreshLibrary() {
  try {
    if (state.drill.level === 'root') {
      state.library = await LibrarySeries() || []
    } else if (state.drill.level === 'seasons') {
      state.seasons = await LibrarySeasons(state.drill.seriesKey) || []
    } else if (state.drill.level === 'episodes') {
      state.episodes = await LibraryEpisodes(state.drill.seriesKey, state.drill.season) || []
    }
  } catch (err) {
    state.error = errText(err)
  }
}

function showToast(msg) {
  state.toast = msg
  if (isUiLocked()) {
    paintToast(msg)
  } else {
    render()
  }
  clearTimeout(window.__toastTimer)
  window.__toastTimer = setTimeout(() => {
    if (state.toast === msg) {
      state.toast = ''
      document.querySelector('.toast')?.remove()
      if (!isUiLocked()) render()
    }
  }, 3200)
}

function paintToast(msg) {
  let el = document.querySelector('.toast')
  if (!el) {
    el = document.createElement('div')
    el.className = 'toast'
    document.body.appendChild(el)
  }
  el.textContent = msg
}

function isUiLocked() {
  if (state.inspectOpen) return true
  if (state.view === 'settings') return true
  if (document.querySelector('.apple-select.open')) return true
  const el = document.activeElement
  if (!el || el === document.body || el === document.documentElement) return false
  const tag = el.tagName
  if (tag === 'SELECT' || tag === 'TEXTAREA') return true
  if (tag === 'INPUT') {
    const type = (el.type || 'text').toLowerCase()
    if (type === 'checkbox' || type === 'radio' || type === 'button' || type === 'submit' || type === 'hidden') return false
    return true
  }
  return false
}

function requestRender() {
  if (isUiLocked()) {
    state.pendingRender = true
    return
  }
  // Don't wipe the DOM mid-click — Wails/WebKit drops the click otherwise
  if (Date.now() - lastPointerAt < 450) {
    state.pendingRender = true
    clearTimeout(window.__deferRenderTimer)
    window.__deferRenderTimer = setTimeout(flushPendingRender, 480)
    return
  }
  state.pendingRender = false
  render()
}

function flushPendingRender() {
  if (!state.pendingRender) return
  if (isUiLocked() || Date.now() - lastPointerAt < 450) {
    clearTimeout(window.__deferRenderTimer)
    window.__deferRenderTimer = setTimeout(flushPendingRender, 480)
    return
  }
  state.pendingRender = false
  render()
}

function render() {
  applyTheme(state.settings?.theme)
  const c = counts()
  const dirClass = document.documentElement.dir === 'rtl' ? 'dir-rtl' : 'dir-ltr'
  document.querySelector('#app').innerHTML = `
    <div class="shell">
      <aside class="sidebar">
        <div class="brand">
          <strong>${escapeHtml(t('appName'))}</strong>
          <span>v${escapeHtml(state.version || '…')} · ${escapeHtml(t('appTag'))}</span>
        </div>
        <nav class="nav">
          <button type="button" class="${state.view === 'library' ? 'active' : ''}" data-view="library">${ICONS.home} ${t('navLibrary')} <kbd>⌘1</kbd></button>
          <button type="button" class="${state.view === 'queue' ? 'active' : ''}" data-view="queue">${ICONS.queue} ${t('navQueue')} <kbd>⌘2</kbd></button>
          <button type="button" class="${state.view === 'history' ? 'active' : ''}" data-view="history">${ICONS.history} ${t('navHistory')} <kbd>⌘3</kbd></button>
          <button type="button" class="${state.view === 'settings' ? 'active' : ''}" data-view="settings">${ICONS.settings} ${t('navSettings')} <kbd>⌘4</kbd></button>
        </nav>
        <div class="user">
          <b>${escapeHtml(state.session.username || (state.session.hasToken ? t('tokenSaved') : t('notLoggedIn')))}</b>
          <span>${t('account')}</span>
          ${!needsLogin() ? `<button type="button" class="link" id="logout" data-action="logout">${t('logout')}</button>` : ''}
        </div>
      </aside>
      <section class="main">
        ${state.view === 'library' ? libraryView(c) : ''}
        ${state.view === 'queue' ? queueView(c) : ''}
        ${state.view === 'history' ? historyView() : ''}
        ${state.view === 'settings' ? settingsView() : ''}
      </section>
    </div>
    ${needsLogin() ? loginSheet() : ''}
    ${state.inspectOpen && state.inspect ? inspectSheet() : ''}
    ${state.toast ? `<div class="toast">${escapeHtml(state.toast)}</div>` : ''}
    ${state.update?.available ? updateBanner() : ''}
`
  bind()
}

function updateBanner() {
  const u = state.update
  return `
    <div class="update-banner">
      <span>${escapeHtml(u.message || '')}</span>
      <button class="btn primary" id="openUpdate">${t('downloadReady')}</button>
      <button class="btn" id="dismissUpdate">${t('later')}</button>
    </div>
  `
}

function filterChips() {
  const items = [
    ['all', t('filterAll')],
    ['active', t('filterActive')],
    ['done', t('filterDone')],
    ['error', t('filterError')],
  ]
  return `
    <div class="filters">
      ${items.map(([id, label]) => `
        <button class="chip ${state.filter === id ? 'active' : ''}" data-filter="${id}">${label}</button>
      `).join('')}
    </div>
  `
}

function libSearchBar() {
  return `
    <div class="lib-search">
      <span class="lib-search-icon">${ICONS.search}</span>
      <input id="libSearch" placeholder="${escapeAttr(t('libSearch'))}" value="${escapeAttr(state.libSearch)}" />
    </div>
  `
}

function breadcrumb() {
  const d = state.drill
  if (d.level === 'root') return ''
  return `
    <div class="crumb">
      <button class="btn plain crumb-back" id="drillBack">${backIcon()} ${t('back')}</button>
      <button class="crumb-link" data-crumb="root">${t('library')}</button>
      ${d.seriesTitle ? `<span class="crumb-sep">/</span><button class="crumb-link" data-crumb="seasons">${escapeHtml(d.seriesTitle)}</button>` : ''}
      ${d.level === 'episodes' && d.seasonTitle ? `<span class="crumb-sep">/</span><span class="crumb-current">${escapeHtml(d.seasonTitle)}</span>` : ''}
    </div>
  `
}

function libraryView(c) {
  const d = state.drill
  let body = ''
  let title = t('library')
  if (d.level === 'root') {
    const items = filteredLibrary()
    body = !items.length
      ? `<div class="empty"><h3>${t('emptyLibrary')}</h3><div>${t('emptyLibraryHint')}</div></div>`
      : `<div class="lib-grid">${items.map(seriesCard).join('')}</div>`
  } else if (d.level === 'seasons') {
    title = d.seriesTitle || t('seasons')
    const items = state.seasons || []
    body = !items.length
      ? `<div class="empty"><h3>${t('emptySeason')}</h3></div>`
      : `<div class="panel">${items.map(seasonRow).join('')}</div>`
  } else {
    title = d.seasonTitle || t('episodes')
    const items = filteredEpisodes()
    body = !items.length
      ? `<div class="empty"><h3>${t('emptyEpisode')}</h3></div>`
      : `<div class="panel">${items.map(jobRow).join('')}</div>`
  }
  return `
    <div class="toolbar">
      <input class="search" id="query" placeholder="${escapeAttr(t('searchPlaceholder'))}" value="${escapeAttr(state.query)}" />
      <button class="btn primary large" id="inspectBtn" ${state.busy ? 'disabled' : ''}>${t('inspect')}</button>
    </div>
    <div class="content">
      ${breadcrumb()}
      <div class="title-row">
        <h1 class="hero-title">${escapeHtml(title)}</h1>
        ${d.level !== 'root' ? `
          <div class="title-actions">
            <button class="btn" data-open-series-folder="${escapeAttr(d.seriesKey)}">${ICONS.folder} ${t('folder')}</button>
            <button class="btn danger" data-remove-series="${escapeAttr(d.seriesKey)}">${ICONS.trash} ${t('removeFromList')}</button>
          </div>
        ` : ''}
      </div>
      <div class="stats">
        <div class="stat"><span>${t('queued')}</span><b>${c.queued}</b></div>
        <div class="stat"><span>${t('active')}</span><b>${c.running}</b></div>
        <div class="stat"><span>${t('paused')}</span><b>${c.paused}</b></div>
        <div class="stat"><span>${t('doneError')}</span><b>${c.done} / ${c.error}</b></div>
      </div>
      ${libSearchBar()}
      ${filterChips()}
      <div class="err">${escapeHtml(state.error)}</div>
      ${body}
    </div>
  `
}

function seriesCard(item) {
  const isSeries = item.kind === 'series'
  const sub = isSeries
    ? `${t('seasonCount', { n: item.seasonCount || 0 })} · ${t('episodeCount', { n: item.episodeCount || 0 })} · ${t('doneCount', { n: item.doneCount || 0 })}`
    : `${item.doneCount ? t('finished') : (item.activeCount ? t('downloading') : t('movie'))}`
  const pct = Math.max(0, Math.min(100, Math.round(item.progress || 0)))
  return `
    <div class="lib-card-wrap">
      <button class="lib-card" data-open-series="${escapeAttr(item.key)}" data-series-title="${escapeAttr(item.title)}" data-series-kind="${escapeAttr(item.kind || '')}">
        ${coverHtml(item.cover, item.title, 'lib-cover')}
        <div class="lib-body">
          <div class="title">${escapeHtml(item.title)}</div>
          <div class="meta">${escapeHtml(sub)}</div>
          <div class="bar"><i style="width:${pct}%"></i></div>
        </div>
        <span class="lib-chevron">${ICONS.chevron}</span>
      </button>
      <div class="lib-card-actions">
        <button class="btn icon-btn" data-open-series-folder="${escapeAttr(item.key)}" title="${escapeAttr(t('folder'))}">${ICONS.folder}</button>
        <button class="btn danger icon-btn" data-remove-series="${escapeAttr(item.key)}" title="${escapeAttr(t('removeFromList'))}">${ICONS.trash}</button>
      </div>
    </div>
  `
}

function seasonRow(item) {
  const pct = Math.max(0, Math.min(100, Math.round(item.progress || 0)))
  const seasonNum = Number(item.season)
  const seasonAttr = Number.isFinite(seasonNum) ? String(seasonNum) : '0'
  return `
    <button type="button" class="row season-row"
      data-open-season="${escapeAttr(seasonAttr)}"
      data-season-title="${escapeAttr(item.title || '')}">
      ${coverHtml(item.cover, item.title)}
      <div>
        <div class="title">${escapeHtml(item.title)}</div>
        <div class="meta">${t('episodeCount', { n: item.episodeCount || 0 })} · ${t('doneCount', { n: item.doneCount || 0 })}</div>
        <div class="bar"><i style="width:${pct}%"></i></div>
      </div>
      <span class="lib-chevron">${ICONS.chevron}</span>
    </button>
  `
}

function queueView(c) {
  const jobs = filteredQueue()
  return `
    <div class="toolbar">
      <input class="search" id="query" placeholder="${escapeAttr(t('searchPlaceholder'))}" value="${escapeAttr(state.query)}" />
      <button class="btn primary large" id="inspectBtn" ${state.busy ? 'disabled' : ''}>${t('inspect')}</button>
    </div>
    <div class="content">
      <h1 class="hero-title">${t('activeQueue')}</h1>
      <div class="stats">
        <div class="stat"><span>${t('queued')}</span><b>${c.queued}</b></div>
        <div class="stat"><span>${t('active')}</span><b>${c.running}</b></div>
        <div class="stat"><span>${t('paused')}</span><b>${c.paused}</b></div>
        <div class="stat"><span>${t('doneError')}</span><b>${c.done} / ${c.error}</b></div>
      </div>
      ${libSearchBar()}
      ${filterChips()}
      <div class="err">${escapeHtml(state.error)}</div>
      <div class="panel">
        ${jobs.length === 0 ? `
          <div class="empty">
            <h3>${t('emptyQueue')}</h3>
            <div>${t('emptyQueueHint')}</div>
          </div>` : jobs.map(jobRow).join('')}
      </div>
    </div>
  `
}

function jobRow(job) {
  const pct = Math.max(0, Math.min(100, Math.round(job.progress || 0)))
  const reason = (job.status === 'done' || job.status === 'queued' || job.status === 'preparing' || job.status === 'running')
    ? ''
    : (job.error || '')
  const canPause = job.status === 'queued' || job.status === 'preparing' || job.status === 'running'
  const canResume = job.status === 'paused'
  const canRetry = job.status === 'error' || job.status === 'canceled' || job.status === 'done'
  const canCancel = job.status === 'queued' || job.status === 'preparing' || job.status === 'running' || job.status === 'paused'
  const title = job.title || (job.episode ? t('episodeN', { n: job.episode }) : '')
  const bits = [
    job.seasonTitle || (job.season ? t('seasonN', { n: job.season }) : ''),
    job.quality || '',
    job.format || '',
    job.message || '',
    `${pct}${pctSymbol()}`,
  ].filter(Boolean)
  return `
    <div class="row ${job.status}" data-job-id="${escapeAttr(job.id)}">
      ${coverHtml(job.cover, title)}
      <div>
        <div class="title">${escapeHtml(title)}</div>
        <div class="meta" data-job-meta>${escapeHtml(bits.join(' · '))}</div>
        ${reason ? `<div class="reason">${escapeHtml(reason)}</div>` : ''}
        <div class="bar"><i style="width:${pct}%"></i></div>
      </div>
      <div class="actions">
        <span class="badge ${job.status}" data-job-badge>${statusLabel(job.status)}</span>
        ${job.filePath ? `<button type="button" class="btn icon-btn" data-preview="${escapeAttr(job.filePath)}" title="Play">${ICONS.play}</button>` : ''}
        ${job.filePath ? `<button type="button" class="btn icon-btn" data-reveal="${escapeAttr(job.filePath)}" title="${escapeAttr(t('folder'))}">${ICONS.folder}</button>` : ''}
        ${canPause ? `<button type="button" class="btn icon-btn" data-pause="${job.id}">${ICONS.pause}</button>` : ''}
        ${canResume ? `<button type="button" class="btn primary icon-btn" data-resume="${job.id}">${ICONS.play}</button>` : ''}
        ${canRetry ? `<button type="button" class="btn icon-btn" data-retry="${job.id}">${ICONS.refresh}</button>` : ''}
        ${canCancel ? `<button type="button" class="btn danger icon-btn" data-cancel="${job.id}">${ICONS.close}</button>` : ''}
        <button type="button" class="btn danger icon-btn" data-remove="${job.id}" title="${escapeAttr(t('removeFromList'))}">${ICONS.trash}</button>
      </div>
    </div>
  `
}

function historyView() {
  const needle = state.libSearch
  const items = (state.history || []).filter((item) => textMatch(item.title, needle))
  return `
    <div class="toolbar"><div style="flex:1"></div></div>
    <div class="content">
      <h1 class="hero-title">${t('history')}</h1>
      ${libSearchBar()}
      <div class="panel">
        ${!items.length ? `<div class="empty"><h3>${t('emptyHistory')}</h3></div>` : items.map((item) => `
          <div class="row done">
            <div class="thumb placeholder">${escapeHtml((item.title || '?').charAt(0))}</div>
            <div>
              <div class="title">${escapeHtml(item.title)}</div>
              <div class="meta">${escapeHtml(item.quality)} · ${escapeHtml(item.format)} · ${escapeHtml(item.downloadedAt)} · ${Number(item.sizeMb || 0).toFixed(1)} ${t('mb')}</div>
            </div>
            <div class="actions">
              ${item.filePath ? `<button class="btn icon-btn" data-preview="${escapeAttr(item.filePath)}">${ICONS.play}</button>` : ''}
              ${item.filePath ? `<button class="btn icon-btn" data-reveal="${escapeAttr(item.filePath)}">${ICONS.folder}</button>` : ''}
            </div>
          </div>
        `).join('')}
      </div>
    </div>
  `
}

function settingsView() {
  const s = state.settings || {}
  const concurrent = s.concurrentDownloads || 2
  const lang = s.language || 'fa'
  return `
    <div class="toolbar"><div style="flex:1"></div></div>
    <div class="content">
      <h1 class="hero-title">${t('settings')}</h1>
      <div class="panel">
        <div class="settings-grid">
          <div class="field" style="grid-column:1/-1">
            <label>${t('downloadPath')}</label>
            <div style="display:flex;gap:8px">
              <input id="downloadPath" value="${escapeAttr(s.downloadPath || '')}" style="flex:1" />
              <button class="btn" id="pickFolder">${t('choose')}</button>
            </div>
          </div>
          <div class="field">
            <label>${t('language')}</label>
            ${appleSelect({ id: 'language', value: lang, options: LANGS.map((l) => ({ value: l.id, label: l.label })) })}
          </div>
          <div class="field">
            <label>${t('theme')}</label>
            ${appleSelect({ id: 'theme', value: s.theme || 'system', options: [
              { value: 'system', label: t('themeSystem') },
              { value: 'light', label: t('themeLight') },
              { value: 'dark', label: t('themeDark') },
            ]})}
          </div>
          <div class="field">
            <label>${t('defaultQuality')}</label>
            ${appleSelect({ id: 'defaultQuality', value: s.defaultQuality || '720p', options: ['1080p','720p','480p','360p'].map((q) => ({ value: q, label: q })) })}
          </div>
          <div class="field">
            <label>${t('defaultFormat')}</label>
            ${appleSelect({ id: 'defaultFormat', value: s.defaultFormat || 'mp4', options: [
              { value: 'mp4', label: 'MP4' },
              { value: 'mkv', label: 'MKV' },
            ]})}
          </div>
          <div class="field">
            <label>${t('concurrent')}</label>
            ${appleSelect({ id: 'concurrent', value: concurrent, options: [1,2,3,4,5].map((n) => ({ value: String(n), label: String(n) })) })}
          </div>
          <div class="settings-toggles">
            ${switchControl({ id: 'autoOpen', checked: !!s.autoOpenFolder, label: t('autoOpen') })}
            ${switchControl({ id: 'notifications', checked: !!s.notifications, label: t('notifications') })}
            ${switchControl({ id: 'checkUpdates', checked: !!s.checkUpdatesOnLaunch, label: t('checkUpdates') })}
          </div>
        </div>
        <div class="footer-actions" style="padding:0 16px 16px;justify-content:space-between;width:100%;flex-wrap:wrap">
          <div class="help" style="margin:0">${t('version', { v: state.version })} · ${t('shortcuts')}</div>
          <div style="display:flex;gap:8px">
            <button class="btn" id="checkUpdateBtn">${t('checkUpdateBtn')}</button>
            <button class="btn primary large" id="saveSettings">${t('saveSettings')}</button>
          </div>
        </div>
        ${state.update ? `<p class="help" style="padding:0 16px 16px">${escapeHtml(state.update.message || '')}</p>` : ''}
      </div>
    </div>
  `
}

function loginSheet() {
  return `
    <div class="overlay">
      <div class="sheet">
        <h2>${t('loginTitle')}</h2>
        <p class="help">${t('loginHelp')}</p>
        <div class="field">
          <label>${t('tokenLabel')}</label>
          <textarea id="token" placeholder="${escapeAttr(t('tokenPlaceholder'))}"></textarea>
        </div>
        <div class="err">${escapeHtml(state.error)}</div>
        <div class="footer-actions">
          <button class="btn primary large" id="loginBtn">${t('loginContinue')}</button>
        </div>
      </div>
    </div>
  `
}

function inspectSheet() {
  const info = state.inspect
  const isSeries = info.kind === 'series'
  const qualities = info.qualities || []
  const audio = info.audio || []
  const subs = info.subtitles || []
  const eps = info.episodes || []
  const defaultQ = (state.settings && state.settings.defaultQuality) || '720p'
  const groups = groupEpisodes(eps)
  const cats = (info.categories || []).join(' · ')
  const qOpts = qualities.length
    ? qualities.map((q) => ({ value: q.quality, label: q.resolution ? `${q.quality} · ${q.resolution}` : q.quality }))
    : [{ value: defaultQ, label: defaultQ }]
  const selectedQ = qualities.some((q) => q.quality === defaultQ) ? defaultQ : (qOpts[0]?.value || defaultQ)
  return `
    <div class="overlay">
      <div class="sheet">
        <div class="cover-row">
          ${info.cover ? `<img src="${escapeAttr(info.cover)}" alt="">` : `<div class="thumb placeholder" style="width:84px;height:84px">${escapeHtml((info.title || '?').charAt(0))}</div>`}
          <div>
            <h2>${escapeHtml(info.title)}</h2>
            <p class="help" style="margin:0">${isSeries ? `${t('series')} · ${eps.length}` : t('movie')}${cats ? ` · ${escapeHtml(cats)}` : ''}</p>
          </div>
        </div>
        ${info.description ? `<p class="help">${escapeHtml(info.description.slice(0, 220))}${info.description.length > 220 ? '…' : ''}</p>` : ''}
        ${isSeries && eps.length ? `
          <div class="field">
            <label>${t('episodes')}</label>
            <div class="eps-head">
              <span class="help" style="margin:0">${t('selectedOf', { n: state.selectedEps.length, total: eps.length })}</span>
              <div class="checks">
                <button type="button" class="btn" id="selectAllEps">${t('selectAll')}</button>
                <button type="button" class="btn" id="selectNoEps">${t('selectNone')}</button>
              </div>
            </div>
            <div class="eps">
              ${groups.map((group) => `
                ${group.title ? `<div class="season-label">${escapeHtml(group.title)}</div>` : ''}
                ${group.items.map((ep) => episodePick(ep)).join('')}
              `).join('')}
            </div>
          </div>
        ` : ''}
        <div class="field">
          <label>${t('quality')}</label>
          ${appleSelect({ id: 'quality', value: selectedQ, options: qOpts })}
        </div>
        ${audio.length ? `
          <div class="field">
            <label>${t('audio')}</label>
            <div class="eps">${audio.map((a) => switchChip({ checked: true, label: escapeHtml(a), attrs: `data-audio="${escapeAttr(a)}"` })).join('')}</div>
          </div>
        ` : ''}
        ${subs.length ? `
          <div class="field">
            <label>${t('subtitles')}</label>
            <div class="eps">${subs.map((s) => switchChip({ checked: true, label: escapeHtml(s), attrs: `data-sub="${escapeAttr(s)}"` })).join('')}</div>
          </div>
        ` : `<p class="help">${t('noSubs')}</p>`}
        <div class="field">
          <label>${t('outputFormat')}</label>
          ${appleSelect({ id: 'format', value: state.settings.defaultFormat || 'mp4', options: [
            { value: 'mp4', label: 'MP4' },
            { value: 'mkv', label: 'MKV' },
          ]})}
        </div>
        <div class="err">${escapeHtml(state.error)}</div>
        <div class="footer-actions">
          <button class="btn" id="closeInspect">${t('cancel')}</button>
          <button class="btn primary large" id="addQueue" ${state.busy ? 'disabled' : ''}>${t('addQueue')}</button>
        </div>
      </div>
    </div>
  `
}

function episodePick(ep) {
  const checked = state.selectedEps.includes(ep.id)
  const fullTitle = String(ep.title || '').trim()
  const numLabel = ep.number ? t('episodeN', { n: ep.number }) : ''
  // Prefer full Filimo title (جلسه N) over generic «قسمت N»
  const generic = !fullTitle || /^قسمت\s*\d+$/i.test(fullTitle) || /^episode\s*\d+$/i.test(fullTitle)
  const label = generic ? (numLabel || fullTitle || ep.id) : fullTitle
  const metaBits = []
  if (!generic && numLabel) metaBits.push(numLabel)
  if (ep.duration) metaBits.push(ep.duration)
  const meta = metaBits.join(' · ')
  return `
    <label class="ep-pick">
      ${coverHtml(ep.cover, label, 'ep-pick-cover')}
      <span class="ep-pick-body">
        <span class="ep-pick-title" title="${escapeAttr(label)}">${escapeHtml(label)}</span>
        ${meta ? `<span class="ep-pick-meta">${escapeHtml(meta)}</span>` : ''}
      </span>
      <span class="switch switch-sm">
        <input type="checkbox" ${checked ? 'checked' : ''} data-ep="${escapeAttr(ep.id)}" role="switch">
        <span class="switch-track" aria-hidden="true"><span class="switch-thumb"></span></span>
      </span>
    </label>
  `
}

function groupEpisodes(eps) {
  const map = new Map()
  for (const ep of eps) {
    const key = ep.seasonTitle || (ep.season ? t('seasonN', { n: ep.season }) : '')
    if (!map.has(key)) map.set(key, [])
    map.get(key).push(ep)
  }
  return [...map.entries()].map(([title, items]) => ({ title, items }))
}

async function openSeries(key, title, kind) {
  if (kind === 'movie' || String(key).startsWith('movie:')) {
    state.drill = { level: 'episodes', seriesKey: key, seriesTitle: title, season: 0, seasonTitle: title }
    state.episodes = await LibraryEpisodes(key, 0) || []
  } else {
    state.drill = { level: 'seasons', seriesKey: key, seriesTitle: title, season: 0, seasonTitle: '' }
    state.seasons = await LibrarySeasons(key) || []
    // یک فصل → مستقیم برو داخل قسمت‌ها
    if (state.seasons.length === 1) {
      const only = state.seasons[0]
      await openSeason(only.season, only.title || title)
      return
    }
  }
  state.view = 'library'
  render()
}

async function openSeason(season, title) {
  const seasonNum = Number.parseInt(String(season ?? '0'), 10)
  const safeSeason = Number.isFinite(seasonNum) ? seasonNum : 0
  state.drill.level = 'episodes'
  state.drill.season = safeSeason
  state.drill.seasonTitle = title || t('seasonN', { n: safeSeason })
  try {
    state.episodes = await LibraryEpisodes(state.drill.seriesKey, safeSeason) || []
  } catch (err) {
    state.error = errText(err)
    state.episodes = []
  }
  state.view = 'library'
  render()
}

async function drillBack() {
  if (state.drill.level === 'episodes' && !String(state.drill.seriesKey).startsWith('movie:')) {
    state.drill.level = 'seasons'
    state.drill.season = 0
    state.drill.seasonTitle = ''
    state.seasons = await LibrarySeasons(state.drill.seriesKey) || []
  } else {
    state.drill = { level: 'root', seriesKey: '', seriesTitle: '', season: 0, seasonTitle: '' }
    state.library = await LibrarySeries() || []
  }
  render()
}

function collectSettingsFromForm() {
  return {
    defaultQuality: $('#defaultQuality')?.value || '720p',
    defaultFormat: $('#defaultFormat')?.value || 'mp4',
    downloadPath: $('#downloadPath')?.value || '',
    autoOpenFolder: $('#autoOpen')?.checked || false,
    showInfoBeforeDL: true,
    concurrentDownloads: Number($('#concurrent')?.value || 2),
    theme: $('#theme')?.value || 'system',
    language: $('#language')?.value || 'fa',
    notifications: $('#notifications')?.checked || false,
    checkUpdatesOnLaunch: $('#checkUpdates')?.checked || false,
  }
}

function bind() {
  const query = $('#query')
  if (query) {
    query.addEventListener('input', () => { state.query = query.value })
    query.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') inspect()
    })
  }
  const libSearch = $('#libSearch')
  if (libSearch) {
    libSearch.addEventListener('input', () => {
      state.libSearch = libSearch.value
      clearTimeout(window.__libSearchTimer)
      window.__libSearchTimer = setTimeout(() => {
        const keep = state.libSearch
        const start = libSearch.selectionStart
        render()
        const el = $('#libSearch')
        if (el) {
          el.focus()
          el.value = keep
          state.libSearch = keep
          try { el.setSelectionRange(start, start) } catch {}
        }
      }, 120)
    })
  }
  document.querySelectorAll('[data-ep]').forEach((el) => {
    el.addEventListener('change', () => {
      const id = el.dataset.ep
      if (el.checked) state.selectedEps = [...new Set([...state.selectedEps, id])]
      else state.selectedEps = state.selectedEps.filter((x) => x !== id)
      const count = document.querySelector('.eps-head .help')
      const total = (state.inspect && state.inspect.episodes || []).length
      if (count) count.textContent = t('selectedOf', { n: state.selectedEps.length, total })
    })
  })
  $('#theme')?.addEventListener('change', () => {
    const theme = $('#theme').value || 'system'
    state.settings = { ...state.settings, theme }
    applyTheme(theme)
  })
  $('#language')?.addEventListener('change', () => {
    const language = $('#language').value || 'fa'
    state.settings = { ...state.settings, language }
    setLang(language)
    render()
  })
  bindAppleSelects()
}

async function handleAppClick(el) {
  if (el.closest('.apple-select-trigger, .apple-select-option, .apple-select-menu')) return false

  const viewBtn = el.closest('[data-view]')
  if (viewBtn) {
    const view = viewBtn.getAttribute('data-view')
    if (!view || view === state.view) {
      if (view === 'library') await refreshLibrary()
      render()
      return true
    }
    state.view = view
    state.error = ''
    if (view === 'library') await refreshLibrary()
    render()
    return true
  }

  const filterBtn = el.closest('[data-filter]')
  if (filterBtn) {
    state.filter = filterBtn.getAttribute('data-filter') || 'all'
    render()
    return true
  }

  if (el.closest('#inspectBtn')) { await inspect(); return true }
  if (el.closest('#loginBtn')) { await login(); return true }
  if (el.closest('#logout') || el.closest('[data-action="logout"]')) {
    state.session = await Logout()
    render()
    return true
  }
  if (el.closest('#closeInspect')) {
    state.inspectOpen = false
    state.error = ''
    render()
    return true
  }
  if (el.closest('#drillBack')) { await drillBack(); return true }
  if (el.closest('#addQueue')) { await enqueue(); return true }
  if (el.closest('#pickFolder')) {
    const folder = await PickFolder()
    if (folder) {
      const input = $('#downloadPath')
      if (input) input.value = folder
    }
    return true
  }
  if (el.closest('#saveSettings')) {
    state.settings = await SaveSettings(collectSettingsFromForm())
    setLang(state.settings.language || 'fa')
    applyTheme(state.settings.theme)
    state.view = 'library'
    showToast(t('settingsSaved'))
    return true
  }
  if (el.closest('#checkUpdateBtn')) {
    state.update = await CheckForUpdates()
    render()
    return true
  }
  if (el.closest('#openUpdate')) {
    if (state.update?.url) OpenURL(state.update.url)
    return true
  }
  if (el.closest('#dismissUpdate')) {
    if (state.update) state.update.available = false
    render()
    return true
  }
  if (el.closest('#selectAllEps')) {
    state.selectedEps = (state.inspect.episodes || []).map((ep) => ep.id)
    render()
    return true
  }
  if (el.closest('#selectNoEps')) {
    state.selectedEps = []
    render()
    return true
  }

  const crumb = el.closest('[data-crumb]')
  if (crumb) {
    if (crumb.getAttribute('data-crumb') === 'root') {
      state.drill = { level: 'root', seriesKey: '', seriesTitle: '', season: 0, seasonTitle: '' }
      state.library = await LibrarySeries() || []
      render()
    } else if (crumb.getAttribute('data-crumb') === 'seasons') {
      state.drill.level = 'seasons'
      state.drill.season = 0
      state.drill.seasonTitle = ''
      state.seasons = await LibrarySeasons(state.drill.seriesKey) || []
      render()
    }
    return true
  }

  const folderBtn = el.closest('[data-open-series-folder]')
  if (folderBtn) {
    try { await OpenSeriesFolder(folderBtn.getAttribute('data-open-series-folder')) } catch (err) {
      state.error = errText(err); render()
    }
    return true
  }

  const removeSeriesBtn = el.closest('[data-remove-series]')
  if (removeSeriesBtn) {
    if (!confirm(t('removeConfirm'))) return true
    try {
      await RemoveSeries(removeSeriesBtn.getAttribute('data-remove-series'))
      state.jobs = await Jobs()
      state.drill = { level: 'root', seriesKey: '', seriesTitle: '', season: 0, seasonTitle: '' }
      state.library = await LibrarySeries() || []
      showToast(t('removedToast'))
    } catch (err) {
      state.error = errText(err)
      render()
    }
    return true
  }

  const seasonEl = el.closest('[data-open-season]')
  if (seasonEl) {
    await openSeason(
      seasonEl.getAttribute('data-open-season'),
      seasonEl.getAttribute('data-season-title') || '',
    )
    return true
  }

  const seriesEl = el.closest('[data-open-series]')
  if (seriesEl) {
    await openSeries(
      seriesEl.getAttribute('data-open-series'),
      seriesEl.getAttribute('data-series-title') || '',
      seriesEl.getAttribute('data-series-kind') || '',
    )
    return true
  }

  const pauseBtn = el.closest('[data-pause]')
  if (pauseBtn) {
    try { await Pause(pauseBtn.getAttribute('data-pause')) } catch (err) { state.error = errText(err) }
    state.jobs = await Jobs()
    render()
    return true
  }
  const resumeBtn = el.closest('[data-resume]')
  if (resumeBtn) {
    try { await Resume(resumeBtn.getAttribute('data-resume')) } catch (err) { state.error = errText(err) }
    state.jobs = await Jobs()
    render()
    return true
  }
  const retryBtn = el.closest('[data-retry]')
  if (retryBtn) {
    try { await Retry(retryBtn.getAttribute('data-retry')) } catch (err) { state.error = errText(err) }
    state.jobs = await Jobs()
    render()
    return true
  }
  const cancelBtn = el.closest('[data-cancel]')
  if (cancelBtn) {
    await Cancel(cancelBtn.getAttribute('data-cancel'))
    state.jobs = await Jobs()
    render()
    return true
  }
  const removeBtn = el.closest('[data-remove]')
  if (removeBtn) {
    try {
      await RemoveJob(removeBtn.getAttribute('data-remove'))
      state.jobs = await Jobs()
      render()
    } catch (err) {
      state.error = errText(err)
      render()
    }
    return true
  }
  const openBtn = el.closest('[data-open]')
  if (openBtn) { OpenPath(openBtn.getAttribute('data-open')); return true }
  const revealBtn = el.closest('[data-reveal]')
  if (revealBtn) { RevealPath(revealBtn.getAttribute('data-reveal')); return true }
  const previewBtn = el.closest('[data-preview]')
  if (previewBtn) { PreviewPath(previewBtn.getAttribute('data-preview')); return true }

  return false
}

function softPatchJobs(jobs) {
  if (!Array.isArray(jobs)) return false
  let hit = 0
  for (const job of jobs) {
    const row = document.querySelector(`[data-job-id="${CSS.escape(String(job.id))}"]`)
    if (!row) continue
    hit += 1
    const pct = Math.max(0, Math.min(100, Math.round(job.progress || 0)))
    const bar = row.querySelector('.bar > i')
    if (bar) bar.style.width = `${pct}%`
    const meta = row.querySelector('[data-job-meta]')
    if (meta) {
      meta.textContent = `${job.quality || ''} · ${job.format || ''} · ${job.message || ''} · ${pct}${pctSymbol()}`
    }
    const badge = row.querySelector('[data-job-badge]')
    if (badge) {
      badge.className = `badge ${job.status}`
      badge.textContent = statusLabel(job.status)
    }
    row.className = `row ${job.status}`
  }
  const c = counts()
  document.querySelectorAll('.stat').forEach((stat) => {
    const label = stat.querySelector('span')?.textContent || ''
    const b = stat.querySelector('b')
    if (!b) return
    if (label === t('queued')) b.textContent = String(c.queued)
    else if (label === t('active')) b.textContent = String(c.running)
    else if (label === t('paused')) b.textContent = String(c.paused)
    else if (label === t('doneError')) b.textContent = `${c.done} / ${c.error}`
  })
  // If any job row is missing (new/removed), caller should full-render
  const rows = document.querySelectorAll('[data-job-id]').length
  if (state.view !== 'queue') return false
  if ((jobs || []).length === 0) return rows === 0
  return rows > 0 && rows === jobs.length
}

function bindAppClicks() {
  if (window.__filimoClicksBound) return
  window.__filimoClicksBound = true

  document.addEventListener('pointerdown', () => {
    lastPointerAt = Date.now()
  }, true)

  document.addEventListener('click', (e) => {
    const t = e.target
    if (!(t instanceof Element)) return
    // let apple-select handle itself
    if (t.closest('.apple-select-trigger, .apple-select-option')) return
    handleAppClick(t).catch((err) => {
      state.error = errText(err)
      render()
    })
  })
}

async function login() {
  state.error = ''
  try {
    state.session = await Login($('#token').value.trim())
    state.settings = await Settings()
    setLang(state.settings.language || 'fa')
    state.jobs = await Jobs()
    state.history = await History()
    state.library = await LibrarySeries() || []
  } catch (err) {
    state.error = errText(err)
  }
  render()
}

async function inspect() {
  state.error = ''
  state.busy = true
  render()
  try {
    const info = await Inspect(state.query)
    state.inspect = info
    state.inspectOpen = true
    const eps = (info && info.episodes) || []
    state.selectedEps = eps.length ? eps.map((ep) => ep.id) : (info && info.id ? [info.id] : [])
  } catch (err) {
    state.error = errText(err)
  }
  state.busy = false
  render()
}

async function enqueue() {
  const info = state.inspect
  if (!info) return
  const seriesKey = info.parentId || info.id
  const seriesTitle = info.seriesTitle || info.title
  let items = [{
    id: info.id,
    title: info.title,
    kind: info.kind || 'movie',
    seriesKey: info.kind === 'series' ? seriesKey : '',
    seriesTitle: info.kind === 'series' ? seriesTitle : '',
    season: info.season || 0,
    episode: info.episode || 0,
    seasonTitle: '',
    cover: info.cover || '',
  }]
  if (info.kind === 'series') {
    const eps = info.episodes || []
    const chosen = eps.filter((ep) => state.selectedEps.includes(ep.id))
    if (!chosen.length) {
      state.error = t('pickEpisode')
      render()
      return
    }
    items = chosen.map((ep) => ({
      id: ep.id,
      title: ep.title || `${seriesTitle} · ${t('episodeN', { n: ep.number || '' })}`.trim(),
      kind: 'series',
      seriesKey,
      seriesTitle,
      season: ep.season || 0,
      episode: ep.number || 0,
      seasonTitle: ep.seasonTitle || (ep.season ? t('seasonN', { n: ep.season }) : ''),
      cover: ep.cover || info.cover || '',
    }))
  }
  const audio = [...document.querySelectorAll('[data-audio]:checked')].map((el) => el.dataset.audio)
  const subtitles = [...document.querySelectorAll('[data-sub]:checked')].map((el) => el.dataset.sub)
  state.busy = true
  render()
  try {
    state.jobs = await Enqueue({
      items,
      quality: $('#quality')?.value || '720p',
      audio,
      subtitles,
      format: $('#format')?.value || 'mp4',
    })
    state.inspectOpen = false
    state.error = ''
    state.view = 'library'
    state.drill = { level: 'root', seriesKey: '', seriesTitle: '', season: 0, seasonTitle: '' }
    state.library = await LibrarySeries() || []
  } catch (err) {
    state.error = errText(err)
  }
  state.busy = false
  render()
}

function bindShortcuts() {
  window.addEventListener('keydown', (e) => {
    const meta = e.metaKey || e.ctrlKey
    if (meta && e.key === '1') { e.preventDefault(); state.view = 'library'; refreshLibrary().then(render) }
    if (meta && e.key === '2') { e.preventDefault(); state.view = 'queue'; render() }
    if (meta && e.key === '3') { e.preventDefault(); state.view = 'history'; render() }
    if (meta && e.key === '4') { e.preventDefault(); state.view = 'settings'; render() }
    if (meta && (e.key === 'l' || e.key === 'L')) {
      e.preventDefault()
      $('#query')?.focus()
    }
    if (meta && (e.key === 'f' || e.key === 'F')) {
      e.preventDefault()
      $('#libSearch')?.focus()
    }
    if (e.key === 'Escape') {
      if (state.inspectOpen) {
        state.inspectOpen = false
        render()
      } else if (state.view === 'library' && state.drill.level !== 'root') {
        drillBack()
      }
    }
  })
}

async function boot() {
  try {
    state.session = await Session()
    state.settings = await Settings()
    setLang(state.settings.language || 'fa')
    state.jobs = await Jobs()
    state.history = await History()
    state.library = await LibrarySeries() || []
    state.version = await AppVersion()
    applyTheme(state.settings.theme)
    if (state.settings.checkUpdatesOnLaunch) {
      CheckForUpdates().then((info) => {
        state.update = info
        if (info?.available) render()
      }).catch(() => {})
    }
  } catch (err) {
    state.error = errText(err)
  }
  EventsOn('queue:update', (jobs) => {
    state.jobs = jobs || []
    if (state.view === 'queue' && !isUiLocked() && softPatchJobs(state.jobs)) {
      return
    }
    clearTimeout(window.__queueUpdateTimer)
    window.__queueUpdateTimer = setTimeout(async () => {
      if (state.view === 'library') {
        try { await refreshLibrary() } catch {}
      }
      requestRender()
    }, state.view === 'queue' ? 250 : 1200)
  })
  EventsOn('history:update', (items) => {
    state.history = items || []
    if (state.view === 'history') requestRender()
  })
  EventsOn('job:done', (job) => {
    showToast(t('doneToast', { title: job?.title || '' }))
  })
  document.addEventListener('focusout', () => {
    setTimeout(flushPendingRender, 0)
  })
  document.addEventListener('change', (e) => {
    if (e.target && e.target.tagName === 'SELECT' && state.view !== 'settings') {
      setTimeout(flushPendingRender, 0)
    }
  })
  bindAppClicks()
  bindShortcuts()
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if ((state.settings?.theme || 'system') === 'system') applyTheme('system')
  })
  render()
}

boot()
