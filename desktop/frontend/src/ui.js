import { t } from './i18n.js'

const chevronDown = `<svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M6 9l6 6 6-6"/></svg>`
const checkMark = `<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M5 12.5l4.5 4.5L19 7"/></svg>`

/** Apple-style switch */
export function switchControl({ id, checked = false, label = '', attrs = '' }) {
  return `
    <label class="switch-row">
      <span class="switch-label">${label}</span>
      <span class="switch">
        <input type="checkbox" id="${id}" ${checked ? 'checked' : ''} ${attrs} role="switch">
        <span class="switch-track" aria-hidden="true"><span class="switch-thumb"></span></span>
      </span>
    </label>
  `
}

/** Compact switch for lists (episodes / tracks) */
export function switchChip({ checked = false, label = '', attrs = '' }) {
  return `
    <label class="switch-chip">
      <span class="switch-chip-label">${label}</span>
      <span class="switch switch-sm">
        <input type="checkbox" ${checked ? 'checked' : ''} ${attrs} role="switch">
        <span class="switch-track" aria-hidden="true"><span class="switch-thumb"></span></span>
      </span>
    </label>
  `
}

/** Apple-style select (custom menu) */
export function appleSelect({ id, value, options, wide = false }) {
  const current = options.find((o) => String(o.value) === String(value)) || options[0]
  const label = current ? current.label : ''
  return `
    <div class="apple-select ${wide ? 'wide' : ''}" data-select-id="${id}">
      <input type="hidden" id="${id}" value="${escapeAttr(current?.value ?? '')}">
      <button type="button" class="apple-select-trigger" aria-haspopup="listbox" aria-expanded="false">
        <span class="apple-select-value">${escapeHtml(label)}</span>
        <span class="apple-select-chevron">${chevronDown}</span>
      </button>
      <div class="apple-select-menu" role="listbox" hidden>
        ${options.map((o) => `
          <button type="button" class="apple-select-option ${String(o.value) === String(current?.value) ? 'selected' : ''}"
            role="option" data-value="${escapeAttr(o.value)}" aria-selected="${String(o.value) === String(current?.value)}">
            <span class="apple-select-check">${String(o.value) === String(current?.value) ? checkMark : ''}</span>
            <span>${escapeHtml(o.label)}</span>
          </button>
        `).join('')}
      </div>
    </div>
  `
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

export function bindAppleSelects(root = document) {
  root.querySelectorAll('.apple-select').forEach((wrap) => {
    const trigger = wrap.querySelector('.apple-select-trigger')
    const menu = wrap.querySelector('.apple-select-menu')
    const hidden = wrap.querySelector('input[type="hidden"]')
    if (!trigger || !menu || !hidden) return

    const close = () => {
      menu.hidden = true
      trigger.setAttribute('aria-expanded', 'false')
      wrap.classList.remove('open')
    }
    const open = () => {
      root.querySelectorAll('.apple-select.open').forEach((other) => {
        if (other !== wrap) {
          other.classList.remove('open')
          const m = other.querySelector('.apple-select-menu')
          const t = other.querySelector('.apple-select-trigger')
          if (m) m.hidden = true
          if (t) t.setAttribute('aria-expanded', 'false')
        }
      })
      menu.hidden = false
      trigger.setAttribute('aria-expanded', 'true')
      wrap.classList.add('open')
    }

    trigger.addEventListener('click', (e) => {
      e.preventDefault()
      e.stopPropagation()
      if (wrap.classList.contains('open')) close()
      else open()
    })

    menu.querySelectorAll('.apple-select-option').forEach((opt) => {
      opt.addEventListener('click', (e) => {
        e.preventDefault()
        e.stopPropagation()
        const val = opt.dataset.value
        hidden.value = val
        const label = opt.querySelector('span:last-child')?.textContent || val
        wrap.querySelector('.apple-select-value').textContent = label
        menu.querySelectorAll('.apple-select-option').forEach((o) => {
          const on = o === opt
          o.classList.toggle('selected', on)
          o.setAttribute('aria-selected', on ? 'true' : 'false')
          const check = o.querySelector('.apple-select-check')
          if (check) check.innerHTML = on ? checkMark : ''
        })
        close()
        hidden.dispatchEvent(new Event('change', { bubbles: true }))
      })
    })
  })

  if (!root.__appleSelectDocBound) {
    root.__appleSelectDocBound = true
    document.addEventListener('click', () => {
      document.querySelectorAll('.apple-select.open').forEach((wrap) => {
        wrap.classList.remove('open')
        const m = wrap.querySelector('.apple-select-menu')
        const t = wrap.querySelector('.apple-select-trigger')
        if (m) m.hidden = true
        if (t) t.setAttribute('aria-expanded', 'false')
      })
    })
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        document.querySelectorAll('.apple-select.open').forEach((wrap) => {
          wrap.classList.remove('open')
          const m = wrap.querySelector('.apple-select-menu')
          const t = wrap.querySelector('.apple-select-trigger')
          if (m) m.hidden = true
          if (t) t.setAttribute('aria-expanded', 'false')
        })
      }
    })
  }
}

export { t }
