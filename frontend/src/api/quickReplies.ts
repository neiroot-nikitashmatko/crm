import { ApiError, requestJson } from '@/api/httpClient'
import type { QuickReply, QuickReplySection } from '@/types/quickReply'

export class QuickRepliesApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'QuickRepliesApiError'
  }
}

interface SectionsResponse {
  items: QuickReplySection[]
}

interface SectionItemResponse {
  item: QuickReplySection
}

interface ReplyItemResponse {
  item: QuickReply
}

async function quickRepliesRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new QuickRepliesApiError(error.message, error.status)
    }
    throw error
  }
}

function normalizeSection(raw: QuickReplySection): QuickReplySection {
  return {
    ...raw,
    replies: Array.isArray(raw.replies) ? raw.replies : [],
  }
}

export async function fetchQuickReplySections(): Promise<QuickReplySection[]> {
  const payload = await quickRepliesRequestJson<SectionsResponse>('/api/v1/quick-reply-sections', {
    method: 'GET',
  })
  return (payload.items ?? []).map(normalizeSection)
}

export async function createQuickReplySection(title: string): Promise<QuickReplySection> {
  const payload = await quickRepliesRequestJson<SectionItemResponse>('/api/v1/quick-reply-sections', {
    method: 'POST',
    body: JSON.stringify({ title }),
  })
  return normalizeSection(payload.item)
}

export async function updateQuickReplySection(id: string, title: string): Promise<QuickReplySection> {
  const payload = await quickRepliesRequestJson<SectionItemResponse>(
    `/api/v1/quick-reply-sections/${encodeURIComponent(id)}`,
    {
      method: 'PATCH',
      body: JSON.stringify({ title }),
    },
  )
  return normalizeSection(payload.item)
}

export async function deleteQuickReplySection(id: string): Promise<void> {
  await quickRepliesRequestJson<{ ok: boolean }>(`/api/v1/quick-reply-sections/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}

export async function createQuickReply(
  sectionId: string,
  title: string,
  body: string,
): Promise<QuickReply> {
  const payload = await quickRepliesRequestJson<ReplyItemResponse>(
    `/api/v1/quick-reply-sections/${encodeURIComponent(sectionId)}/replies`,
    {
      method: 'POST',
      body: JSON.stringify({ title, body }),
    },
  )
  return payload.item
}

export async function updateQuickReply(id: string, title: string, body: string): Promise<QuickReply> {
  const payload = await quickRepliesRequestJson<ReplyItemResponse>(
    `/api/v1/quick-replies/${encodeURIComponent(id)}`,
    {
      method: 'PATCH',
      body: JSON.stringify({ title, body }),
    },
  )
  return payload.item
}

export async function deleteQuickReply(id: string): Promise<void> {
  await quickRepliesRequestJson<{ ok: boolean }>(`/api/v1/quick-replies/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}
