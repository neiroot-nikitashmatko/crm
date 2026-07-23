import { ApiError, getApiBaseUrl, requestJson } from '@/api/httpClient'
import { getAuthToken, notifyUnauthorized } from '@/api/session'
import {
  appendMockAvitoMessage,
  getMockAvitoChats,
  getMockAvitoLeadChat,
  isAvitoChatsMockEnabled,
  isMockAvitoLeadId,
} from '@/mocks/avitoChats'
import { getMockUnreadCountForLead } from '@/mocks/notificationBadges'
import type { AvitoChatListItem } from '@/types/avitoChat'
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
  createdAt?: string
  updatedAt?: string
  unreadCount?: number
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

interface AvitoChatsListResponse {
  items: AvitoChatDto[]
}

function toTimestamp(value?: string): number {
  if (!value) return Date.now()
  const parsed = Date.parse(value)
  return Number.isFinite(parsed) ? parsed : Date.now()
}

function normalizeChatListItem(raw: AvitoChatDto): AvitoChatListItem {
  return {
    id: raw.id,
    chatId: raw.chatId,
    leadId: raw.leadId,
    peerNickname: (raw.peerNickname || '').trim() || 'Пользователь Авито',
    peerAvatarUrl: (raw.peerAvatarUrl || '').trim(),
    itemTitle: (raw.itemTitle || '').trim(),
    updatedAt: toTimestamp(raw.updatedAt || raw.createdAt),
    unreadCount: Math.max(0, Number(raw.unreadCount) || 0),
  }
}

function withMockUnread(items: AvitoChatListItem[]): AvitoChatListItem[] {
  if (!isAvitoChatsMockEnabled()) return items
  return items.map((item) => {
    if (!isMockAvitoLeadId(item.leadId)) return item
    return {
      ...item,
      unreadCount: getMockUnreadCountForLead(item.leadId),
    }
  })
}

function withMockChats(items: AvitoChatListItem[]): AvitoChatListItem[] {
  if (!isAvitoChatsMockEnabled()) return items
  const mocks = withMockUnread(getMockAvitoChats())
  if (import.meta.env.VITE_MOCK_AVITO_CHATS === 'true') {
    return mocks
  }
  return [...mocks, ...items].sort((a, b) => b.updatedAt - a.updatedAt)
}

export async function fetchAvitoChats(): Promise<AvitoChatListItem[]> {
  if (import.meta.env.VITE_MOCK_AVITO_CHATS === 'true') {
    return withMockUnread(getMockAvitoChats())
  }

  try {
    const payload = await requestJson<AvitoChatsListResponse>('/api/v1/integrations/avito/chats', {
      method: 'GET',
    })
    return withMockChats((payload.items ?? []).map(normalizeChatListItem))
  } catch (error) {
    if (isAvitoChatsMockEnabled()) {
      return withMockUnread(getMockAvitoChats())
    }
    if (error instanceof ApiError) {
      throw new AvitoChatApiError(error.message, error.status)
    }
    throw error
  }
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

const leadChatCache = new Map<string, AvitoLeadChatBundle>()
const leadChatInflight = new Map<string, Promise<AvitoLeadChatBundle>>()

function cloneChatBundle(bundle: AvitoLeadChatBundle): AvitoLeadChatBundle {
  return {
    linked: bundle.linked,
    participant: { ...bundle.participant },
    messages: bundle.messages.map((item) => ({ ...item })),
  }
}

export function getCachedAvitoLeadChat(leadId: string): AvitoLeadChatBundle | null {
  const cached = leadChatCache.get(leadId.trim())
  return cached ? cloneChatBundle(cached) : null
}

export function setCachedAvitoLeadChat(leadId: string, bundle: AvitoLeadChatBundle): void {
  const id = leadId.trim()
  if (!id) return
  leadChatCache.set(id, cloneChatBundle(bundle))
}

export function patchCachedAvitoLeadChatMessage(leadId: string, message: LeadChatMessage): void {
  const id = leadId.trim()
  const cached = leadChatCache.get(id)
  if (!cached) return
  const index = cached.messages.findIndex((item) => item.id === message.id)
  const nextMessages =
    index === -1
      ? [...cached.messages, message].sort((a, b) => a.createdAt - b.createdAt)
      : cached.messages.map((item, i) => (i === index ? message : item))
  leadChatCache.set(id, {
    ...cached,
    linked: true,
    messages: nextMessages,
  })
}

async function loadAvitoLeadChat(leadId: string): Promise<AvitoLeadChatBundle> {
  if (isAvitoChatsMockEnabled() && isMockAvitoLeadId(leadId)) {
    const mock = getMockAvitoLeadChat(leadId)
    if (mock) {
      setCachedAvitoLeadChat(leadId, mock)
      return cloneChatBundle(mock)
    }
  }

  const payload = await avitoRequestJson<AvitoLeadChatBundleResponse>(
    `/api/v1/integrations/avito/chats/${encodeURIComponent(leadId)}`,
    { method: 'GET' },
  )

  const nickname = payload.chat?.peerNickname?.trim() || 'Пользователь Авито'
  const avatarUrl = payload.chat?.peerAvatarUrl?.trim() || null

  const bundle: AvitoLeadChatBundle = {
    linked: Boolean(payload.linked),
    participant: { nickname, avatarUrl },
    messages: (payload.messages ?? []).map(mapMessage),
  }
  setCachedAvitoLeadChat(leadId, bundle)
  return cloneChatBundle(bundle)
}

export async function fetchAvitoLeadChat(leadId: string): Promise<AvitoLeadChatBundle> {
  const id = leadId.trim()
  if (!id) {
    return {
      linked: false,
      participant: { nickname: 'Пользователь Авито', avatarUrl: null },
      messages: [],
    }
  }

  const inflight = leadChatInflight.get(id)
  if (inflight) return cloneChatBundle(await inflight)

  const request = loadAvitoLeadChat(id).finally(() => {
    leadChatInflight.delete(id)
  })
  leadChatInflight.set(id, request)
  return request
}

/** Warm cache in background (e.g. on chat list hover). */
export function prefetchAvitoLeadChat(leadId: string): void {
  const id = leadId.trim()
  if (!id || leadChatCache.has(id) || leadChatInflight.has(id)) return
  void fetchAvitoLeadChat(id).catch(() => {
    // ignore prefetch errors
  })
}

export async function sendAvitoLeadMessage(
  leadId: string,
  text: string,
  files: File[] = [],
): Promise<LeadChatMessage[]> {
  if (isAvitoChatsMockEnabled() && isMockAvitoLeadId(leadId)) {
    if (files.length > 0) {
      throw new AvitoChatApiError('В моках отправка файлов пока не поддержана', 400)
    }
    return appendMockAvitoMessage(leadId, text)
  }

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

export type AvitoMessageSSEHandler = (payload: {
  leadId: string
  message: LeadChatMessage
  createdLead: boolean
}) => void

export type AvitoChatReadSSEHandler = (payload: { leadId: string }) => void

export function subscribeAvitoMessages(
  onMessage: AvitoMessageSSEHandler,
  options?: { signal?: AbortSignal; onChatRead?: AvitoChatReadSSEHandler },
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
                const parsed = JSON.parse(currentData) as {
                  leadId?: string
                  message?: AvitoMessageDto
                  createdLead?: boolean
                }
                if (parsed.leadId && parsed.message) {
                  onMessage({
                    leadId: parsed.leadId,
                    message: mapMessage(parsed.message),
                    createdLead: Boolean(parsed.createdLead),
                  })
                }
              } catch {
                // ignore malformed payload
              }
            } else if (currentEvent === 'avito-chat-read' && currentData.trim() !== '') {
              try {
                const parsed = JSON.parse(currentData) as { leadId?: string }
                const leadId = parsed.leadId?.trim()
                if (leadId) {
                  options?.onChatRead?.({ leadId })
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
