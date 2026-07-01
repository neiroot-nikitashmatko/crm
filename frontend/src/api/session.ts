export const AUTH_TOKEN_STORAGE_KEY = 'proclients.auth.token'
export const AUTH_USER_STORAGE_KEY = 'proclients.auth.user'

let unauthorizedHandler: (() => void) | null = null

export function setUnauthorizedHandler(handler: () => void): void {
  unauthorizedHandler = handler
}

export function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null
  return sessionStorage.getItem(AUTH_TOKEN_STORAGE_KEY)
}

export function setAuthToken(token: string): void {
  sessionStorage.setItem(AUTH_TOKEN_STORAGE_KEY, token)
}

export function clearAuthToken(): void {
  sessionStorage.removeItem(AUTH_TOKEN_STORAGE_KEY)
}

export function clearAuthSession(): void {
  clearAuthToken()
  sessionStorage.removeItem(AUTH_USER_STORAGE_KEY)
}

export function notifyUnauthorized(): void {
  clearAuthSession()
  unauthorizedHandler?.()
}
