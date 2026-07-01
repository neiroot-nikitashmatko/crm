import type { Task } from '@/types/task'
import { ApiError, requestJson } from '@/api/httpClient'

interface TasksListResponse {
  items: Task[]
}

interface TaskItemResponse {
  item: Task
}

export class TasksApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'TasksApiError'
  }
}

async function tasksRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new TasksApiError(error.message, error.status)
    }
    throw error
  }
}

export async function fetchTasks(): Promise<Task[]> {
  const payload = await tasksRequestJson<TasksListResponse>('/api/v1/tasks', { method: 'GET' })
  return payload.items
}

export async function createTask(payload: {
  title: string
  text: string
  dueAt: number | null
  leadId?: string
  dealId?: string
  createdBy: string
}): Promise<Task> {
  const response = await tasksRequestJson<TaskItemResponse>('/api/v1/tasks', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return response.item
}

export async function patchTask(
  taskId: string,
  payload: { title?: string; text?: string; dueAt?: number | null },
): Promise<Task> {
  const response = await tasksRequestJson<TaskItemResponse>(`/api/v1/tasks/${taskId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
  return response.item
}

export async function completeTask(taskId: string): Promise<Task> {
  const response = await tasksRequestJson<TaskItemResponse>(`/api/v1/tasks/${taskId}/complete`, {
    method: 'POST',
  })
  return response.item
}

export async function completeTasksByLead(leadId: string): Promise<void> {
  await tasksRequestJson<{ ok: boolean }>('/api/v1/tasks/complete-by-lead', {
    method: 'POST',
    body: JSON.stringify({ leadId }),
  })
}
