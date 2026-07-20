/** New-lead alerts: in-tab sound + OS notification when tab/window is in background. */

let audioContext: AudioContext | null = null
let unlocked = false
let notificationPermissionAsked = false

function getAudioContext(): AudioContext | null {
  const Ctx =
    window.AudioContext ||
    (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
  if (!Ctx) return null
  if (!audioContext) {
    audioContext = new Ctx()
  }
  return audioContext
}

/** Call after a user gesture so Chrome allows playback / notifications. */
export async function unlockNewLeadSound(): Promise<void> {
  const ctx = getAudioContext()
  if (!ctx) return

  try {
    if (ctx.state === 'suspended') {
      await ctx.resume()
    }
    const buffer = ctx.createBuffer(1, 1, ctx.sampleRate)
    const source = ctx.createBufferSource()
    source.buffer = buffer
    source.connect(ctx.destination)
    source.start(0)
  } catch {
    return
  }

  unlocked = ctx.state === 'running'
  void ensureNotificationPermission()
}

async function ensureNotificationPermission(): Promise<void> {
  if (typeof Notification === 'undefined') return
  if (notificationPermissionAsked) return
  if (Notification.permission !== 'default') {
    notificationPermissionAsked = true
    return
  }
  notificationPermissionAsked = true
  try {
    await Notification.requestPermission()
  } catch {
    // ignore
  }
}

function tone(ctx: AudioContext, frequency: number, startAt: number, duration: number) {
  const oscillator = ctx.createOscillator()
  const gain = ctx.createGain()
  oscillator.type = 'sine'
  oscillator.frequency.value = frequency
  gain.gain.setValueAtTime(0.0001, startAt)
  gain.gain.exponentialRampToValueAtTime(0.18, startAt + 0.015)
  gain.gain.exponentialRampToValueAtTime(0.0001, startAt + duration)
  oscillator.connect(gain)
  gain.connect(ctx.destination)
  oscillator.start(startAt)
  oscillator.stop(startAt + duration + 0.02)
}

export async function playNewLeadSound(): Promise<boolean> {
  await unlockNewLeadSound()
  const ctx = getAudioContext()
  if (!ctx || ctx.state !== 'running') return false

  try {
    const t = ctx.currentTime + 0.01
    tone(ctx, 880, t, 0.14)
    tone(ctx, 1175, t + 0.15, 0.18)
    return true
  } catch {
    return false
  }
}

function showNewLeadNotification(title: string, body: string): void {
  if (typeof Notification === 'undefined') return
  if (Notification.permission !== 'granted') return

  try {
    const notification = new Notification(title, {
      body,
      tag: 'proclients-new-lead',
      silent: false,
    })
    notification.onclick = () => {
      window.focus()
      notification.close()
    }
  } catch {
    // ignore
  }
}

export type NewLeadNotifyInfo = {
  firstName?: string
  patronymic?: string
  trafficSource?: string
  leadNumber?: number
}

/**
 * Sound when tab is active; OS notification (with system sound) when Chrome is
 * minimized / another tab is focused. Sound is still attempted either way.
 */
export async function notifyNewLeadArrival(info: NewLeadNotifyInfo = {}): Promise<void> {
  const played = await playNewLeadSound()
  const tabHidden = typeof document !== 'undefined' && document.visibilityState !== 'visible'

  // Background / minimized Chrome: browsers often block page audio — use OS notification.
  if (!tabHidden && played) return

  const name = [info.firstName, info.patronymic].filter(Boolean).join(' ').trim()
  const source = (info.trafficSource ?? '').trim()
  const numberLabel =
    typeof info.leadNumber === 'number' && info.leadNumber > 0 ? `#${info.leadNumber}` : ''

  const title = numberLabel ? `Новый лид ${numberLabel}` : 'Новый лид'
  const bodyParts = [name || 'Без имени', source].filter(Boolean)
  showNewLeadNotification(title, bodyParts.join(' · ') || 'Появился новый лид в канбане')
}

export function isNewLeadSoundUnlocked(): boolean {
  return unlocked
}

/** Resume audio when user returns to the CRM tab. */
export function bindNewLeadSoundVisibility(): () => void {
  const onVisible = () => {
    if (document.visibilityState === 'visible') {
      void unlockNewLeadSound()
    }
  }
  document.addEventListener('visibilitychange', onVisible)
  return () => document.removeEventListener('visibilitychange', onVisible)
}
