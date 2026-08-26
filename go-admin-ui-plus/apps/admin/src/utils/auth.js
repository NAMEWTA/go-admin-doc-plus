import Cookies from 'js-cookie'

const TokenKey = 'Admin-Token'

const isDesktopHost = () =>
  typeof globalThis.go?.desktop?.Bridge?.Bootstrap === 'function'

export function getToken() {
  if (isDesktopHost()) return localStorage.getItem(TokenKey) || undefined
  return Cookies.get(TokenKey)
}

export function setToken(token) {
  if (isDesktopHost()) {
    localStorage.setItem(TokenKey, token)
    return token
  }
  return Cookies.set(TokenKey, token)
}

export function removeToken() {
  if (isDesktopHost()) {
    localStorage.removeItem(TokenKey)
    return
  }
  return Cookies.remove(TokenKey)
}
