const STORAGE_KEY = 'proclients.taskCommentDrafts'

function readAll(): Record<string, string> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return {}
    const parsed: unknown = JSON.parse(raw)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {}
    const result: Record<string, string> = {}
    for (const [taskId, text] of Object.entries(parsed as Record<string, unknown>)) {
      if (typeof text === 'string') result[taskId] = text
    }
    return result
  } catch {
    return {}
  }
}

function writeAll(drafts: Record<string, string>) {
  try {
    if (Object.keys(drafts).length === 0) {
      localStorage.removeItem(STORAGE_KEY)
      return
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(drafts))
  } catch {
    // ignore quota / private mode errors
  }
}

export function loadTaskCommentDraft(taskId: string): string {
  return readAll()[taskId] ?? ''
}

export function saveTaskCommentDraft(taskId: string, text: string) {
  const drafts = readAll()
  if (!text.trim()) {
    if (!(taskId in drafts)) return
    delete drafts[taskId]
  } else {
    drafts[taskId] = text
  }
  writeAll(drafts)
}

export function clearTaskCommentDraft(taskId: string) {
  const drafts = readAll()
  if (!(taskId in drafts)) return
  delete drafts[taskId]
  writeAll(drafts)
}
