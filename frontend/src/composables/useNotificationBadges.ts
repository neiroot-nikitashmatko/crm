import { computed, ref } from 'vue'
import { getApiBaseUrl } from '@/api/httpClient'
import {
  fetchNotificationSummary,
  markAvitoChatRead as markAvitoChatReadRequest,
} from '@/api/notifications'
import { subscribeAvitoMessages } from '@/api/avitoChat'
import { getAuthToken } from '@/api/session'
import { useAuth } from '@/composables/useAuth'
import {
  bindNewLeadSoundVisibility,
  notifyNewChatMessageArrival,
  unlockNewLeadSound,
} from '@/utils/newLeadSound'

const newLeadsCount = ref(0)
const unreadChatsCount = ref(0)
const loading = ref(false)

let started = false
let refreshTimer: ReturnType<typeof setInterval> | null = null
let leadsAbort: AbortController | null = null
let messagesAbort: AbortController | null = null
let refreshInFlight: Promise<void> | null = null
let refreshAgain = false
let unbindSoundVisibility: (() => void) | null = null
let unlockClickBound = false

function bindUnlockOnFirstClick(): void {
  if (unlockClickBound || typeof document === 'undefined') return
  unlockClickBound = true
  const unlock = () => {
    void unlockNewLeadSound()
  }
  document.addEventListener('pointerdown', unlock, { once: true, capture: true })
  document.addEventListener('keydown', unlock, { once: true, capture: true })
}

function formatBadgeCount(count: number): string {
  if (count <= 0) return ''
  if (count > 99) return '99+'
  return String(count)
}

async function refreshNotificationSummary(): Promise<void> {
  const token = getAuthToken()
  if (!token) {
    newLeadsCount.value = 0
    unreadChatsCount.value = 0
    return
  }

  if (refreshInFlight) {
    refreshAgain = true
    return refreshInFlight
  }

  refreshInFlight = (async () => {
    do {
      refreshAgain = false
      loading.value = true
      try {
        const summary = await fetchNotificationSummary()
        newLeadsCount.value = summary.newLeadsCount
        unreadChatsCount.value = summary.unreadChatsCount
      } catch {
        // Keep previous counters on transient errors.
      } finally {
        loading.value = false
      }
    } while (refreshAgain)
    refreshInFlight = null
  })()

  return refreshInFlight
}

function scheduleRefresh(): void {
  void refreshNotificationSummary()
}

async function markAvitoChatRead(leadId: string): Promise<void> {
  try {
    await markAvitoChatReadRequest(leadId)
  } catch {
    // ignore
  }
  await refreshNotificationSummary()
}

function startLeadEventsStream(): void {
  if (leadsAbort) return
  const token = getAuthToken()
  if (!token) return

  const controller = new AbortController()
  leadsAbort = controller

  void (async () => {
    try {
      const response = await fetch(`${getApiBaseUrl()}/api/v1/events/leads`, {
        method: 'GET',
        headers: { Authorization: `Bearer ${token}` },
        signal: controller.signal,
      })
      if (!response.ok || !response.body) {
        leadsAbort = null
        return
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      let currentEvent = ''
      let currentData = ''

      while (true) {
        const { value, done } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })

        while (true) {
          const idx = buffer.indexOf('\n')
          if (idx === -1) break
          const line = buffer.slice(0, idx).replace(/\r$/, '')
          buffer = buffer.slice(idx + 1)

          if (line === '') {
            if (currentEvent === 'lead-created' && currentData.trim() !== '') {
              scheduleRefresh()
            }
            currentEvent = ''
            currentData = ''
            continue
          }

          if (line.startsWith('event:')) {
            currentEvent = line.slice('event:'.length).trim()
            continue
          }
          if (line.startsWith('data:')) {
            const chunk = line.slice('data:'.length).trim()
            currentData = currentData ? `${currentData}\n${chunk}` : chunk
          }
        }
      }
    } catch {
      // aborted / network
    } finally {
      if (leadsAbort === controller) {
        leadsAbort = null
      }
    }
  })()
}

function startMessageEventsStream(): void {
  if (messagesAbort) return
  const controller = new AbortController()
  messagesAbort = controller
  subscribeAvitoMessages(
    ({ message, createdLead }) => {
      scheduleRefresh()
      if (createdLead) return
      if (message.direction !== 'incoming') return
      void notifyNewChatMessageArrival({
        preview: message.kind === 'image' ? '[Изображение]' : message.text,
      })
    },
    { signal: controller.signal },
  )
}

function handleVisibilityChange(): void {
  if (document.visibilityState === 'visible') {
    scheduleRefresh()
  }
}

function startNotificationBadges(): void {
  if (started) return
  started = true
  scheduleRefresh()
  startLeadEventsStream()
  startMessageEventsStream()
  refreshTimer = setInterval(scheduleRefresh, 60_000)
  document.addEventListener('visibilitychange', handleVisibilityChange)
  unbindSoundVisibility = bindNewLeadSoundVisibility()
  bindUnlockOnFirstClick()
}

function stopNotificationBadges(): void {
  if (!started) return
  started = false
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
  if (leadsAbort) {
    leadsAbort.abort()
    leadsAbort = null
  }
  if (messagesAbort) {
    messagesAbort.abort()
    messagesAbort = null
  }
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  unbindSoundVisibility?.()
  unbindSoundVisibility = null
  newLeadsCount.value = 0
  unreadChatsCount.value = 0
}

export function useNotificationBadges() {
  const { canAccessSection } = useAuth()

  const showLeadsBadge = computed(() => canAccessSection('leads'))
  const showChatsBadge = computed(() => canAccessSection('chats'))

  const newLeadsBadge = computed(() =>
    showLeadsBadge.value ? formatBadgeCount(newLeadsCount.value) : '',
  )
  const unreadChatsBadge = computed(() =>
    showChatsBadge.value ? formatBadgeCount(unreadChatsCount.value) : '',
  )

  return {
    newLeadsCount,
    unreadChatsCount,
    newLeadsBadge,
    unreadChatsBadge,
    showLeadsBadge,
    showChatsBadge,
    loading,
    refreshNotificationSummary,
    markAvitoChatRead,
    startNotificationBadges,
    stopNotificationBadges,
  }
}
