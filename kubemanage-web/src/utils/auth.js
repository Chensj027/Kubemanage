import { reactive } from 'vue'

export const ADMIN_AUTHORITY_ID = 111

const readJSON = (key, fallback) => {
  try { return JSON.parse(localStorage.getItem(key)) || fallback } catch (_) { return fallback }
}

export const authState = reactive({
  user: readJSON('user', null),
  menus: readJSON('menus', []),
  ruleNames: readJSON('ruleNames', []),
  loaded: false
})

export function setAuthInfo(data = {}) {
  authState.user = data.user || null
  authState.menus = data.menus || []
  authState.ruleNames = data.ruleNames || []
  authState.loaded = true
  localStorage.setItem('user', JSON.stringify(authState.user))
  localStorage.setItem('menus', JSON.stringify(authState.menus))
  localStorage.setItem('ruleNames', JSON.stringify(authState.ruleNames))
  if (authState.user) localStorage.setItem('username', authState.user.userName || authState.user.username || '')
}

export function clearAuth() {
  ['token', 'username', 'loginDate', 'user', 'menus', 'ruleNames'].forEach(k => localStorage.removeItem(k))
  authState.user = null
  authState.menus = []
  authState.ruleNames = []
  authState.loaded = false
}

export function getAuthorityId(user = authState.user) {
  return Number(user && (user.authorityId || user.AuthorityId))
}

export function isAdmin(user = authState.user) {
  return getAuthorityId(user) === ADMIN_AUTHORITY_ID
}

export function normalizeMenuPath(path) {
  if (typeof path !== 'string') return ''
  const cleanPath = path.trim().split(/[?#]/, 1)[0]
  if (!cleanPath) return ''
  const normalized = `/${cleanPath.replace(/^\/+|\/+$/g, '')}`
  return normalized === '/' ? normalized : normalized.replace(/\/+$/g, '')
}

export function collectMenuPaths(menus, out = new Set()) {
  const menuList = menus || []
  menuList.forEach(menu => {
    const path = normalizeMenuPath(menu.path)
    if (path) out.add(path)
    collectMenuPaths(menu.children, out)
  })
  return out
}

export function hasMenuPath(path, aliases = []) {
  if (isAdmin()) return true
  const allowedPaths = collectMenuPaths(authState.menus)
  return [path, ...aliases]
    .map(normalizeMenuPath)
    .filter(Boolean)
    .some(candidate => allowedPaths.has(candidate))
}

export function hasPermission(path, method = 'GET') {
  if (isAdmin()) return true
  const target = `${path},${String(method).toUpperCase()}`
  return authState.ruleNames.some(rule => String(rule).toUpperCase() === target.toUpperCase())
}
