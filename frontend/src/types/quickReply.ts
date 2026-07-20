export interface QuickReply {
  id: string
  sectionId: string
  title: string
  body: string
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface QuickReplySection {
  id: string
  title: string
  sortOrder: number
  createdAt: string
  updatedAt: string
  replies: QuickReply[]
}
