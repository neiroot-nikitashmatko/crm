import { isAvitoChatsMockEnabled } from '@/mocks/avitoChats'

export interface MockNotificationSummary {
  newLeadsCount: number
  unreadChatsCount: number
}

/** Непрочитанные моковые чаты: leadId → количество входящих. */
const unreadMockByLead = new Map<string, number>([['mock-lead-anna', 2]])
const mockNewLeadsCount = 1

export function getMockUnreadCountForLead(leadId: string): number {
  return unreadMockByLead.get(leadId.trim()) ?? 0
}

export function getMockNotificationSummary(): MockNotificationSummary {
  let unreadChats = 0
  for (const count of unreadMockByLead.values()) {
    if (count > 0) unreadChats += 1
  }
  return {
    newLeadsCount: mockNewLeadsCount,
    unreadChatsCount: unreadChats,
  }
}

export function markMockChatRead(leadId: string): void {
  const id = leadId.trim()
  if (!id) return
  unreadMockByLead.delete(id)
}

export function mergeNotificationSummaryWithMocks(
  real: MockNotificationSummary,
): MockNotificationSummary {
  if (!isAvitoChatsMockEnabled()) return real
  const mock = getMockNotificationSummary()
  return {
    newLeadsCount: Math.max(real.newLeadsCount, mock.newLeadsCount),
    unreadChatsCount: Math.max(real.unreadChatsCount, mock.unreadChatsCount),
  }
}
