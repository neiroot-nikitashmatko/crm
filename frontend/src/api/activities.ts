import { ApiError, requestJson } from '@/api/httpClient'
import { normalizeStoredActivity, type StoredActivity } from '@/types/attachment'

export class ActivitiesApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'ActivitiesApiError'
  }
}

function wrapApiError(error: unknown): never {
  if (error instanceof ApiError) {
    throw new ActivitiesApiError(error.message, error.status)
  }
  throw error
}

export async function createDealComment(dealId: string, text: string): Promise<StoredActivity> {
  try {
    const payload = await requestJson<{ item: unknown }>(`/api/v1/deals/${dealId}/activities`, {
      method: 'POST',
      body: JSON.stringify({ text }),
    })
    return normalizeStoredActivity(payload.item)
  } catch (error) {
    wrapApiError(error)
  }
}

export async function createLeadComment(leadId: string, text: string): Promise<StoredActivity> {
  try {
    const payload = await requestJson<{ item: unknown }>(`/api/v1/leads/${leadId}/activities`, {
      method: 'POST',
      body: JSON.stringify({ text }),
    })
    return normalizeStoredActivity(payload.item)
  } catch (error) {
    wrapApiError(error)
  }
}

export async function createTaskComment(taskId: string, text: string): Promise<StoredActivity> {
  try {
    const payload = await requestJson<{ item: unknown }>(`/api/v1/tasks/${taskId}/activities`, {
      method: 'POST',
      body: JSON.stringify({ text }),
    })
    return normalizeStoredActivity(payload.item)
  } catch (error) {
    wrapApiError(error)
  }
}
