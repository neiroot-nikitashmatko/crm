import type { AvitoChatListItem } from '@/types/avitoChat'
import type { LeadChatMessage, LeadChatParticipant } from '@/types/leadChat'

export interface MockAvitoChatBundle {
  chat: AvitoChatListItem
  participant: LeadChatParticipant
  messages: LeadChatMessage[]
}

const now = Date.now()
const hour = 60 * 60 * 1000
const minute = 60 * 1000

const mockBundles: Record<string, MockAvitoChatBundle> = {
  'mock-lead-anna': {
    chat: {
      id: 'mock-chat-1',
      chatId: 'avito-mock-chat-1',
      leadId: 'mock-lead-anna',
      peerNickname: 'Анна_Соколова',
      peerAvatarUrl: 'https://i.pravatar.cc/128?u=anna-sokolova',
      itemTitle: 'Диван угловой, серый',
      updatedAt: now - 12 * minute,
      unreadCount: 0,
    },
    participant: {
      nickname: 'Анна_Соколова',
      avatarUrl: 'https://i.pravatar.cc/128?u=anna-sokolova',
    },
    messages: [
      {
        id: 'mock-msg-1-1',
        direction: 'incoming',
        text: 'Здравствуйте! Диван ещё в наличии?',
        kind: 'text',
        createdAt: now - 2 * hour,
      },
      {
        id: 'mock-msg-1-2',
        direction: 'outgoing',
        text: 'Добрый день! Да, в наличии. Можем изготовить под ваш размер.',
        kind: 'text',
        createdAt: now - 2 * hour + 8 * minute,
        status: 'sent',
      },
      {
        id: 'mock-msg-1-3',
        direction: 'incoming',
        text: 'Отлично. А сколько по срокам и доставке до Мытищ?',
        kind: 'text',
        createdAt: now - 90 * minute,
      },
      {
        id: 'mock-msg-1-4',
        direction: 'outgoing',
        text: 'Срок 7–10 дней. Доставка по МО — от 2500 ₽, точную сумму скажем после замера.',
        kind: 'text',
        createdAt: now - 80 * minute,
        status: 'sent',
      },
      {
        id: 'mock-msg-1-5',
        direction: 'incoming',
        text: 'Хорошо, давайте на замер. Когда можете?',
        kind: 'text',
        createdAt: now - 12 * minute,
      },
    ],
  },
  'mock-lead-igor': {
    chat: {
      id: 'mock-chat-2',
      chatId: 'avito-mock-chat-2',
      leadId: 'mock-lead-igor',
      peerNickname: 'ИгорьП',
      peerAvatarUrl: 'https://i.pravatar.cc/128?u=igor-p',
      itemTitle: 'Кресло реклайнер',
      updatedAt: now - 45 * minute,
      unreadCount: 0,
    },
    participant: {
      nickname: 'ИгорьП',
      avatarUrl: 'https://i.pravatar.cc/128?u=igor-p',
    },
    messages: [
      {
        id: 'mock-msg-2-1',
        direction: 'incoming',
        text: 'Вечер добрый. Кресло можно в ткани другого цвета?',
        kind: 'text',
        createdAt: now - 5 * hour,
      },
      {
        id: 'mock-msg-2-2',
        direction: 'outgoing',
        text: 'Да, есть палитра из 12 оттенков. Пришлю образцы.',
        kind: 'text',
        createdAt: now - 4 * hour,
        status: 'sent',
      },
      {
        id: 'mock-msg-2-3',
        direction: 'incoming',
        text: 'Спасибо, жду. Интересует тёмно-зелёный.',
        kind: 'text',
        createdAt: now - 45 * minute,
      },
    ],
  },
  'mock-lead-marina': {
    chat: {
      id: 'mock-chat-3',
      chatId: 'avito-mock-chat-3',
      leadId: 'mock-lead-marina',
      peerNickname: 'Марина',
      peerAvatarUrl: '',
      itemTitle: 'Кухонный уголок',
      updatedAt: now - 3 * hour,
      unreadCount: 0,
    },
    participant: {
      nickname: 'Марина',
      avatarUrl: null,
    },
    messages: [
      {
        id: 'mock-msg-3-1',
        direction: 'incoming',
        text: 'Подскажите, делаете ли вы кухонные уголки на заказ?',
        kind: 'text',
        createdAt: now - 6 * hour,
      },
      {
        id: 'mock-msg-3-2',
        direction: 'outgoing',
        text: 'Да, работаем по индивидуальным размерам. Напишите длину сторон и желаемую ткань.',
        kind: 'text',
        createdAt: now - 5 * hour,
        status: 'sent',
      },
      {
        id: 'mock-msg-3-3',
        direction: 'incoming',
        text: 'Пока прикидываю. Вернусь завтра с размерами.',
        kind: 'text',
        createdAt: now - 3 * hour,
      },
    ],
  },
}

function cloneMessages(messages: LeadChatMessage[]): LeadChatMessage[] {
  return messages.map((item) => ({ ...item }))
}

export function isMockAvitoLeadId(leadId: string): boolean {
  return leadId.trim().startsWith('mock-lead-')
}

export function getMockAvitoChats(): AvitoChatListItem[] {
  return Object.values(mockBundles)
    .map((bundle) => ({ ...bundle.chat }))
    .sort((a, b) => b.updatedAt - a.updatedAt)
}

export function getMockAvitoLeadChat(leadId: string): {
  linked: boolean
  participant: LeadChatParticipant
  messages: LeadChatMessage[]
} | null {
  const bundle = mockBundles[leadId.trim()]
  if (!bundle) return null
  return {
    linked: true,
    participant: { ...bundle.participant },
    messages: cloneMessages(bundle.messages),
  }
}

export function appendMockAvitoMessage(
  leadId: string,
  text: string,
): LeadChatMessage[] {
  const key = leadId.trim()
  const bundle = mockBundles[key]
  if (!bundle) {
    throw new Error('Моковый чат не найден')
  }

  const trimmed = text.trim()
  if (!trimmed) {
    throw new Error('Пустое сообщение')
  }

  const message: LeadChatMessage = {
    id: `mock-msg-out-${Date.now()}`,
    direction: 'outgoing',
    text: trimmed,
    kind: 'text',
    createdAt: Date.now(),
    status: 'sent',
  }
  bundle.messages = [...bundle.messages, message]
  bundle.chat = { ...bundle.chat, updatedAt: message.createdAt }
  return [message]
}

/** Моки включены в dev, пока явно не выключены через VITE_MOCK_AVITO_CHATS=false */
export function isAvitoChatsMockEnabled(): boolean {
  if (import.meta.env.VITE_MOCK_AVITO_CHATS === 'false') return false
  if (import.meta.env.VITE_MOCK_AVITO_CHATS === 'true') return true
  return import.meta.env.DEV
}
