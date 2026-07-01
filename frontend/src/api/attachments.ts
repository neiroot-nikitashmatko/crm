import { ApiError, requestBlob, requestJson, uploadMultipart } from '@/api/httpClient'
import { normalizeStoredActivity, normalizeStoredAttachment, type StoredActivity, type StoredAttachment } from '@/types/attachment'

interface AttachmentsListResponse {
  items: StoredAttachment[]
  activity?: unknown
}

interface DeleteAttachmentResponse {
  ok: boolean
  activity?: unknown
}

export interface UploadAttachmentsResult {
  items: StoredAttachment[]
  activity?: StoredActivity
}

export interface DeleteAttachmentResult {
  ok: boolean
  activity?: StoredActivity
}

export class AttachmentsApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'AttachmentsApiError'
  }
}

function wrapApiError(error: unknown): never {
  if (error instanceof ApiError) {
    throw new AttachmentsApiError(error.message, error.status)
  }
  throw error
}

function normalizeAttachments(items: unknown[], fallbackUploadedBy = ''): StoredAttachment[] {
  return items.map((item) => normalizeStoredAttachment(item, fallbackUploadedBy))
}

function normalizeUploadResult(payload: AttachmentsListResponse): UploadAttachmentsResult {
  return {
    items: normalizeAttachments(payload.items ?? []),
    activity: payload.activity ? normalizeStoredActivity(payload.activity) : undefined,
  }
}

function normalizeDeleteResult(payload: DeleteAttachmentResponse): DeleteAttachmentResult {
  return {
    ok: payload.ok,
    activity: payload.activity ? normalizeStoredActivity(payload.activity) : undefined,
  }
}

export async function uploadDealAttachments(dealId: string, files: File[]): Promise<UploadAttachmentsResult> {
  try {
    const payload = await uploadMultipart<AttachmentsListResponse>(
      `/api/v1/deals/${dealId}/attachments`,
      files,
    )
    return normalizeUploadResult(payload)
  } catch (error) {
    wrapApiError(error)
  }
}

export async function uploadTaskAttachments(taskId: string, files: File[]): Promise<UploadAttachmentsResult> {
  try {
    const payload = await uploadMultipart<AttachmentsListResponse>(
      `/api/v1/tasks/${taskId}/attachments`,
      files,
    )
    return normalizeUploadResult(payload)
  } catch (error) {
    wrapApiError(error)
  }
}

export async function deleteAttachment(attachmentId: string): Promise<DeleteAttachmentResult> {
  try {
    const payload = await requestJson<DeleteAttachmentResponse>(`/api/v1/attachments/${attachmentId}`, {
      method: 'DELETE',
    })
    return normalizeDeleteResult(payload)
  } catch (error) {
    wrapApiError(error)
  }
}

export async function downloadAttachmentFile(attachmentId: string, fileName: string): Promise<void> {
  try {
    const blob = await requestBlob(`/api/v1/attachments/${attachmentId}/content`, {
      method: 'GET',
    })
    const objectUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = fileName
    link.click()
    URL.revokeObjectURL(objectUrl)
  } catch (error) {
    wrapApiError(error)
  }
}
