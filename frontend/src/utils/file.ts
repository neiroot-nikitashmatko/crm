export function formatFileSize(value: number): string {
  if (!Number.isFinite(value) || value <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let size = value
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex += 1
  }
  const precision = unitIndex === 0 ? 0 : 1
  return `${size.toFixed(precision)} ${units[unitIndex]}`
}

export function getFileExtension(name: string): string {
  const dotIndex = name.lastIndexOf('.')
  if (dotIndex === -1 || dotIndex === name.length - 1) return ''
  return name.slice(dotIndex + 1).toLowerCase()
}

export function isImageMimeType(mimeType: string): boolean {
  return /^image\//.test(mimeType)
}
