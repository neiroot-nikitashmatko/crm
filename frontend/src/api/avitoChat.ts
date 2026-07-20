import { ApiError, getApiBaseUrl, requestJson } from '@/api/httpClient'
import { getAuthToken, notifyUnauthorized } from '@/api/session'
import type { LeadChatMessage, LeadChatParticipant } from '@/types/leadChat'

export class AvitoChatApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'AvitoChatApiError'
  }
}

interface AvitoChatDto {
  id: string
  chatId: string
  leadId: string
  peerUserId?: number
  peerNickname: string
  peerAvatarUrl: string
  itemId?: number
  itemTitle: string
}

interface AvitoMessageDto {
  id: string
  chatId: string
  messageId: string
  direction: 'incoming' | 'outgoing' | string
  messageType: string
  text: string
  authorId?: number
  sentAt: string
  createdAt: string
}

interface AvitoLeadChatBundleResponse {
  linked: boolean
  chat?: AvitoChatDto
  messages: AvitoMessageDto[]
}

interface AvitoMessagesResponse {
  items: AvitoMessageDto[]
}

export interface AvitoLeadChatBundle {
  linked: boolean
  participant: LeadChatParticipant
  messages: LeadChatMessage[]
}

function isImageUrl(value: string): boolean {
  const trimmed = value.trim()
  if (!/^https?:\/\//i.test(trimmed)) return false
  return /\.(jpe?g|png|gif|webp|bmp|heic|heif)(\?|$)/i.test(trimmed) || trimmed.includes('avito')
}

function mapMessage(dto: AvitoMessageDto): LeadChatMessage {
  const direction = dto.direction === 'outgoing' ? 'outgoing' : 'incoming'
  const createdAt = Date.parse(dto.sentAt || dto.createdAt)
  const text = dto.text ?? ''
  const messageType = (dto.messageType || 'text').toLowerCase()
  const kind =
    messageType === 'image' || isImageUrl(text) || text === '[Изображение]' || text === '[image]'
      ? 'image'
      : 'text'

  return {
    id: dto.id || dto.messageId,
    direction,
    text: kind === 'image' && (text === '[Изображение]' || text === '[image]') ? '' : text,
    kind,
    imageUrl: kind === 'image' && isImageUrl(text) ? text.trim() : null,
    createdAt: Number.isFinite(createdAt) ? createdAt : Date.now(),
    status: direction === 'outgoing' ? 'sent' : undefined,
  }
}

async function avitoRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new AvitoChatApiError(error.message, error.status)
    }
    throw error
  }
}

export async function fetchAvitoLeadChat(leadId: string): Promise<AvitoLeadChatBundle> {
  const payload = await avitoRequestJson<AvitoLeadChatBundleResponse>(
    `/api/v1/integrations/avito/chats/${encodeURIComponent(leadId)}`,
    { method: 'GET' },
  )

  const nickname = payload.chat?.peerNickname?.trim() || 'Пользователь Авито'
  const avatarUrl = payload.chat?.peerAvatarUrl?.trim() || null

  return {
    linked: Boolean(payload.linked),
    participant: { nickname, avatarUrl },
    messages: (payload.messages ?? []).map(mapMessage),
  }
}

export async function sendAvitoLeadMessage(
  leadId: string,
  text: string,
  files: File[] = [],
): Promise<LeadChatMessage[]> {
  const path = `/api/v1/integrations/avito/chats/${encodeURIComponent(leadId)}/messages`

  if (files.length === 0) {
    const payload = await avitoRequestJson<AvitoMessagesResponse>(path, {
      method: 'POST',
      body: JSON.stringify({ text }),
    })
    return (payload.items ?? []).map(mapMessage)
  }

  const formData = new FormData()
  if (text.trim()) {
    formData.append('text', text.trim())
  }
  for (const file of files) {
    formData.append('file', file, file.name)
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
    throw new AvitoChatApiError('Сервер недоступен. Проверьте, что backend запущен.', 0)
  }

  if (!response.ok) {
    let message = `Ошибка API (${response.status})`
    try {
      const payload = (await response.json()) as { error?: string }
      if (payload.error) message = payload.error
    } catch {
      // ignore
    }
    if (response.status === 401) {
      notifyUnauthorized()
    }
    throw new AvitoChatApiError(message, response.status)
  }

  const payload = (await response.json()) as AvitoMessagesResponse
  return (payload.items ?? []).map(mapMessage)
}

export type AvitoMessageSSEHandler = (payload: { leadId: string; message: LeadChatMessage }) => void

export function subscribeAvitoMessages(
  onMessage: AvitoMessageSSEHandler,
  options?: { signal?: AbortSignal },
): void {
  const token = getAuthToken()
  if (!token) return

  void (async () => {
    try {
      const response = await fetch(`${getApiBaseUrl()}/api/v1/events/avito-messages`, {
        method: 'GET',
        headers: { Authorization: `Bearer ${token}` },
        signal: options?.signal,
      })
      if (!response.ok || !response.body) return

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      let currentEvent = ''
      let currentData = ''

      while (true) {
        const { value, done } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })

        while (true) {
          const idx = buffer.indexOf('\n')
          if (idx === -1) break
          const line = buffer.slice(0, idx).replace(/\r$/, '')
          buffer = buffer.slice(idx + 1)

          if (line === '') {
            if (currentEvent === 'avito-message' && currentData.trim() !== '') {
              try {
                const parsed = JSON.parse(currentData) as { leadId?: string; message?: AvitoMessageDto }
                if (parsed.leadId && parsed.message) {
                  onMessage({
                    leadId: parsed.leadId,
                    message: mapMessage(parsed.message),
                  })
                }
              } catch {
                // ignore malformed payload
              }
            }
            currentEvent = ''
            currentData = ''
            continue
          }

          if (line.startsWith('event:')) {
            currentEvent = line.slice('event:'.length).trim()
            continue
          }
          if (line.startsWith('data:')) {
            const chunk = line.slice('data:'.length).trim()
            currentData = currentData ? `${currentData}\n${chunk}` : chunk
          }
        }
      }
    } catch {
      // aborted / network
    }
  })()
}
