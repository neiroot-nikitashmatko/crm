import { getFileExtension, isImageMimeType } from '@/utils/file'

export type AttachmentPreviewKind = 'image' | 'pdf' | 'text'

export function getAttachmentPreviewKind(attachment: {
  mimeType: string
  name: string
}): AttachmentPreviewKind | null {
  const mimeType = attachment.mimeType.trim().toLowerCase()

  if (isImageMimeType(mimeType)) {
    return 'image'
  }
  if (mimeType === 'application/pdf') {
    return 'pdf'
  }
  if (mimeType === 'text/plain') {
    return 'text'
  }

  const extension = getFileExtension(attachment.name)
  if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'].includes(extension)) {
    return 'image'
  }
  if (extension === 'pdf') {
    return 'pdf'
  }
  if (extension === 'txt') {
    return 'text'
  }

  return null
}

export function canPreviewAttachment(attachment: { mimeType: string; name: string }): boolean {
  return getAttachmentPreviewKind(attachment) !== null
}
