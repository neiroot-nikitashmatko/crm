export interface StoredAttachment {
  id: string
  name: string
  size: number
  mimeType: string
  uploadedBy: string
  uploadedAt: number
}

export interface StoredActivity {
  id: string
  type: 'system' | 'comment'
  author: string
  text: string
  createdAt: number
}

export function normalizeStoredAttachment(raw: unknown, fallbackUploadedBy = ''): StoredAttachment {
  const item = raw as Record<string, unknown>
  return {
    id: String(item?.id ?? ''),
    name: String(item?.name ?? 'Файл'),
    size: Number(item?.size ?? 0),
    mimeType: String(item?.mimeType ?? item?.mime_type ?? 'application/octet-stream'),
    uploadedBy: String(item?.uploadedBy ?? item?.uploaded_by ?? fallbackUploadedBy),
    uploadedAt: Number(item?.uploadedAt ?? item?.uploaded_at ?? Date.now()),
  }
}

export function normalizeStoredActivity(raw: unknown, fallbackAuthor = ''): StoredActivity {
  const item = raw as Record<string, unknown>
  const type = item?.type === 'comment' ? 'comment' : 'system'
  return {
    id: String(item?.id ?? ''),
    type,
    author: String(item?.author ?? fallbackAuthor),
    text: String(item?.text ?? ''),
    createdAt: Number(item?.createdAt ?? item?.created_at ?? Date.now()),
  }
}
