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

export function getApiBaseUrl(): string {
  const rawBaseUrl = import.meta.env.VITE_API_BASE_URL
  if (typeof rawBaseUrl !== 'string' || rawBaseUrl.trim() === '') {
    throw new Error('Не задан VITE_API_BASE_URL')
  }
  return rawBaseUrl.replace(/\/+$/, '')
}

function parseApiErrorMessage(status: number, rawError?: string): string {
  if (status === 401) {
    return 'Сессия устарела. Войдите в систему заново.'
  }
  if (rawError?.includes('foreign key')) {
    return 'Сессия устарела. Войдите в систему заново.'
  }
  if (rawError?.includes('phone')) {
    return 'Некорректный формат телефона. Укажите номер полностью, например +79001234567.'
  }
  if (rawError) {
    return rawError
  }
  if (status === 0) {
    return 'Сервер недоступен. Проверьте, что backend запущен.'
  }
  return `Ошибка API (${status})`
}

export async function requestJson<T>(
  path: string,
  init?: RequestInit,
  options?: { auth?: boolean },
): Promise<T> {
  const useAuth = options?.auth !== false
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
    let message = parseApiErrorMessage(response.status)
    try {
      const payload = (await response.json()) as ErrorResponse
      message = parseApiErrorMessage(response.status, payload.error)
    } catch {
      // ignore parse error
    }

    if (response.status === 401 && useAuth) {
      notifyUnauthorized()
    }

    throw new ApiError(message, response.status)
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
    let message = parseApiErrorMessage(response.status)
    try {
      const payload = (await response.json()) as ErrorResponse
      message = parseApiErrorMessage(response.status, payload.error)
    } catch {
      // ignore parse error
    }

    if (response.status === 401) {
      notifyUnauthorized()
    }

    throw new ApiError(message, response.status)
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
    let message = parseApiErrorMessage(response.status)
    try {
      const payload = (await response.json()) as ErrorResponse
      message = parseApiErrorMessage(response.status, payload.error)
    } catch {
      // ignore parse error
    }

    if (response.status === 401) {
      notifyUnauthorized()
    }

    throw new ApiError(message, response.status)
  }

  return (await response.json()) as T
}
