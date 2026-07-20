export type LeadChatMessageDirection = 'incoming' | 'outgoing'

export type LeadChatMessageStatus = 'sending' | 'sent' | 'failed'

export interface LeadChatParticipant {
  nickname: string
  avatarUrl: string | null
}

export interface LeadChatMessage {
  id: string
  direction: LeadChatMessageDirection
  text: string
  kind?: 'text' | 'image'
  imageUrl?: string | null
  createdAt: number
  status?: LeadChatMessageStatus
}
