import { getAuthToken, notifyUnauthorized } from '@/api/session'

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

interface ErrorResponse {
  error?: string
}

export interface RequestJsonOptions {
  auth?: boolean
  logoutOn401?: boolean
}

export function getApiBaseUrl(): string {
  const rawBaseUrl = import.meta.env.VITE_API_BASE_URL
  if (typeof rawBaseUrl !== 'string' || rawBaseUrl.trim() === '') {
    throw new Error('Не задан VITE_API_BASE_URL')
  }
  return rawBaseUrl.replace(/\/+$/, '')
}

function parseApiErrorMessage(status: number, rawError?: string): string {
  if (isAuthSessionError(rawError)) {
    return 'Сессия устарела. Войдите в систему заново.'
  }
  if (rawError?.includes('foreign key')) {
    return 'Сессия устарела. Войдите в систему заново.'
  }
  if (rawError === 'некорректный телефон') {
    return 'Некорректный формат телефона. Укажите номер полностью, например +79001234567.'
  }
  if (rawError) {
    return rawError
  }
  if (status === 401) {
    return 'Не удалось выполнить запрос. Попробуйте ещё раз.'
  }
  if (status === 0) {
    return 'Сервер недоступен. Проверьте, что backend запущен.'
  }
  return `Ошибка API (${status})`
}

function isAuthSessionError(rawError?: string): boolean {
  const message = rawError?.trim().toLowerCase() ?? ''
  return (
    message.includes('authorization required') ||
    message.includes('invalid or expired token') ||
    message.includes('unauthorized')
  )
}

function shouldLogoutOn401(status: number, rawError: string | undefined, logoutOn401: boolean): boolean {
  return logoutOn401 && status === 401 && isAuthSessionError(rawError)
}

async function readApiError(response: Response): Promise<string | undefined> {
  try {
    const payload = (await response.json()) as ErrorResponse
    return payload.error
  } catch {
    return undefined
  }
}

function throwApiError(status: number, rawError: string | undefined, logoutOn401: boolean): never {
  const message = parseApiErrorMessage(status, rawError)
  if (shouldLogoutOn401(status, rawError, logoutOn401)) {
    notifyUnauthorized()
  }
  throw new ApiError(message, status)
}

export async function requestJson<T>(
  path: string,
  init?: RequestInit,
  options?: RequestJsonOptions,
): Promise<T> {
  const useAuth = options?.auth !== false
  const logoutOn401 = options?.logoutOn401 !== false
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init?.headers as Record<string, string> | undefined),
  }

  if (useAuth) {
    const token = getAuthToken()
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }
  }

  let response: Response
  try {
    response = await fetch(`${getApiBaseUrl()}${path}`, {
      ...init,
      headers,
    })
  } catch {
    throw new ApiError('Сервер недоступен. Проверьте, что backend запущен.', 0)
  }

  if (!response.ok) {
    throwApiError(response.status, await readApiError(response), useAuth && logoutOn401)
  }

  return (await response.json()) as T
}

export async function requestBlob(path: string, init?: RequestInit): Promise<Blob> {
  const headers: Record<string, string> = {
    ...(init?.headers as Record<string, string> | undefined),
  }

  const token = getAuthToken()
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  let response: Response
  try {
    response = await fetch(`${getApiBaseUrl()}${path}`, {
      ...init,
      headers,
    })
  } catch {
    throw new ApiError('Сервер недоступен. Проверьте, что backend запущен.', 0)
  }

  if (!response.ok) {
    throwApiError(response.status, await readApiError(response), true)
  }

  return response.blob()
}

export async function uploadMultipart<T>(path: string, files: File[]): Promise<T> {
  const formData = new FormData()
  for (const file of files) {
    formData.append('file', file)
  }

  const headers: Record<string, string> = {}
  const token = getAuthToken()
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  let response: Response
  try {
    response = await fetch(`${getApiBaseUrl()}${path}`, {
      method: 'POST',
      headers,
      body: formData,
    })
  } catch {
    throw new ApiError('Сервер недоступен. Проверьте, что backend запущен.', 0)
  }

  if (!response.ok) {
    throwApiError(response.status, await readApiError(response), true)
  }

  return (await response.json()) as T
}
