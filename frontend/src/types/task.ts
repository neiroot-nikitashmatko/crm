export type TaskKanbanColumnId = 'overdue' | 'today' | 'tomorrow' | 'later' | 'no-deadline'

export type TaskStatus = 'active' | 'completed'

export type TaskActivityType = 'system' | 'comment'

export interface TaskActivityEntry {
  id: string
  type: TaskActivityType
  author: string
  text: string
  createdAt: Date
}

export interface TaskAttachment {
  id: string
  name: string
  size: number
  mimeType: string
  uploadedBy: string
  uploadedAt: Date
}

export interface Task {
  id: string
  title: string
  text: string
  dueAt: number | null
  leadId?: string
  dealId?: string
  createdBy: string
  createdAt: Date
  status: TaskStatus
  activities: TaskActivityEntry[]
  attachments: TaskAttachment[]
}

export interface NewTaskPayload {
  title: string
  text: string
  dueAt: number | null
  leadId?: string
  dealId?: string
  createdBy?: string
}

export interface UpdateTaskPayload {
  title?: string
  text?: string
  dueAt?: number | null
}
