export const AUTH_TOKEN_STORAGE_KEY = 'proclients.auth.token'
export const AUTH_USER_STORAGE_KEY = 'proclients.auth.user'

let unauthorizedHandler: (() => void) | null = null

export function setUnauthorizedHandler(handler: () => void): void {
  unauthorizedHandler = handler
}

function migrateKeyFromSession(key: string): void {
  if (typeof window === 'undefined') return
  if (localStorage.getItem(key) !== null) {
    sessionStorage.removeItem(key)
    return
  }
  const fromSession = sessionStorage.getItem(key)
  if (fromSession !== null) {
    localStorage.setItem(key, fromSession)
    sessionStorage.removeItem(key)
  }
}

/** Переносит сессию из sessionStorage в localStorage (общий для вкладок). */
export function migrateAuthStorage(): void {
  migrateKeyFromSession(AUTH_TOKEN_STORAGE_KEY)
  migrateKeyFromSession(AUTH_USER_STORAGE_KEY)
}

function readStorageItem(key: string): string | null {
  if (typeof window === 'undefined') return null
  migrateKeyFromSession(key)
  return localStorage.getItem(key)
}

function writeStorageItem(key: string, value: string): void {
  localStorage.setItem(key, value)
  sessionStorage.removeItem(key)
}

function removeStorageItem(key: string): void {
  localStorage.removeItem(key)
  sessionStorage.removeItem(key)
}

export function getAuthToken(): string | null {
  return readStorageItem(AUTH_TOKEN_STORAGE_KEY)
}

export function setAuthToken(token: string): void {
  writeStorageItem(AUTH_TOKEN_STORAGE_KEY, token)
}

export function clearAuthToken(): void {
  removeStorageItem(AUTH_TOKEN_STORAGE_KEY)
}

export function getAuthUserRaw(): string | null {
  return readStorageItem(AUTH_USER_STORAGE_KEY)
}

export function setAuthUserRaw(value: string): void {
  writeStorageItem(AUTH_USER_STORAGE_KEY, value)
}

export function clearAuthSession(): void {
  clearAuthToken()
  removeStorageItem(AUTH_USER_STORAGE_KEY)
}

export function notifyUnauthorized(): void {
  clearAuthSession()
  unauthorizedHandler?.()
}
