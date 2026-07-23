import { requestJson } from '@/api/httpClient'
import {
  getMockNotificationSummary,
  markMockChatRead,
  mergeNotificationSummaryWithMocks,
} from '@/mocks/notificationBadges'
import { isAvitoChatsMockEnabled } from '@/mocks/avitoChats'

export interface NotificationSummary {
  newLeadsCount: number
  unreadChatsCount: number
}

export async function fetchNotificationSummary(): Promise<NotificationSummary> {
  if (import.meta.env.VITE_MOCK_AVITO_CHATS === 'true') {
    return getMockNotificationSummary()
  }

  try {
    const payload = await requestJson<NotificationSummary>('/api/v1/notifications/summary', {
      method: 'GET',
    })
    const real = {
      newLeadsCount: Number(payload.newLeadsCount) || 0,
      unreadChatsCount: Number(payload.unreadChatsCount) || 0,
    }
    return mergeNotificationSummaryWithMocks(real)
  } catch (error) {
    if (isAvitoChatsMockEnabled()) {
      return getMockNotificationSummary()
    }
    throw error
  }
}

export async function markAvitoChatRead(leadId: string): Promise<void> {
  const id = leadId.trim()
  if (!id) return

  if (id.startsWith('mock-lead-')) {
    markMockChatRead(id)
    return
  }

  await requestJson(`/api/v1/integrations/avito/chats/${encodeURIComponent(id)}/read`, {
    method: 'POST',
    body: '{}',
  })
}
