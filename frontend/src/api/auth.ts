import { getApiBaseUrl } from '@/api/httpClient'
import { clearAuthSession, setAuthToken, setAuthUserRaw } from '@/api/session'

interface LoginResponseUser {
  id: string
  phone: string
  role: string
  position?: string
  firstName?: string
  lastName?: string
  patronymic?: string
}

interface LoginResponse {
  user: LoginResponseUser
  token: string
}

interface ErrorResponse {
  error?: string
}

export class AuthApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'AuthApiError'
    this.status = status
  }
}

export interface LoginResult {
  user: LoginResponseUser
  token: string
}

export async function loginByPhone(phone: string, password: string): Promise<LoginResult> {
  const response = await fetch(`${getApiBaseUrl()}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ phone, password }),
  })
  if (!response.ok) {
    let message = 'Не удалось выполнить вход'
    try {
      const payload = (await response.json()) as ErrorResponse
      if (payload?.error) message = payload.error
    } catch {
      // ignore parse error
    }
    throw new AuthApiError(message, response.status)
  }
  const payload = (await response.json()) as LoginResponse
  if (!payload.token || !payload.user) {
    throw new AuthApiError('Некорректный ответ сервера', 500)
  }
  return {
    user: payload.user,
    token: payload.token,
  }
}

export function persistAuthSession(user: LoginResponseUser, token: string): void {
  setAuthToken(token)
  setAuthUserRaw(JSON.stringify(user))
}

export function clearPersistedAuthSession(): void {
  clearAuthSession()
}

export type { LoginResponseUser as AuthUser }
