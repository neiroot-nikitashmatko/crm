/** New-lead alerts: in-tab sound + OS notification when window is not focused. */

let audioContext: AudioContext | null = null
let unlocked = false
let notificationPermissionAsked = false
let htmlAudio: HTMLAudioElement | null = null
let beepDataUri: string | null = null

function buildBeepDataUri(): string {
  const sampleRate = 22050
  const durationSec = 0.28
  const numSamples = Math.floor(sampleRate * durationSec)
  const dataSize = numSamples * 2
  const buffer = new ArrayBuffer(44 + dataSize)
  const view = new DataView(buffer)

  const writeString = (offset: number, value: string) => {
    for (let i = 0; i < value.length; i += 1) {
      view.setUint8(offset + i, value.charCodeAt(i))
    }
  }

  writeString(0, 'RIFF')
  view.setUint32(4, 36 + dataSize, true)
  writeString(8, 'WAVE')
  writeString(12, 'fmt ')
  view.setUint32(16, 16, true)
  view.setUint16(20, 1, true)
  view.setUint16(22, 1, true)
  view.setUint32(24, sampleRate, true)
  view.setUint32(28, sampleRate * 2, true)
  view.setUint16(32, 2, true)
  view.setUint16(34, 16, true)
  writeString(36, 'data')
  view.setUint32(40, dataSize, true)

  for (let i = 0; i < numSamples; i += 1) {
    const t = i / sampleRate
    const freq = t < 0.14 ? 880 : 1175
    const envelope = Math.min(1, t * 30) * Math.min(1, (durationSec - t) * 30)
    const sample = Math.sin(2 * Math.PI * freq * t) * envelope * 0.35
    view.setInt16(44 + i * 2, Math.max(-1, Math.min(1, sample)) * 0x7fff, true)
  }

  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (let i = 0; i < bytes.length; i += 1) {
    binary += String.fromCharCode(bytes[i])
  }
  return `data:audio/wav;base64,${btoa(binary)}`
}

function getBeepDataUri(): string {
  if (!beepDataUri) beepDataUri = buildBeepDataUri()
  return beepDataUri
}

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

function getHtmlAudio(): HTMLAudioElement {
  if (!htmlAudio) {
    htmlAudio = new Audio(getBeepDataUri())
    htmlAudio.preload = 'auto'
  }
  return htmlAudio
}

/** Call after a user gesture so browsers allow playback / notifications. */
export async function unlockNewLeadSound(): Promise<void> {
  const ctx = getAudioContext()
  if (ctx) {
    try {
      if (ctx.state === 'suspended') {
        await ctx.resume()
      }
      const buffer = ctx.createBuffer(1, 1, ctx.sampleRate)
      const source = ctx.createBufferSource()
      source.buffer = buffer
      source.connect(ctx.destination)
      source.start(0)
      unlocked = ctx.state === 'running'
    } catch {
      // ignore
    }
  }

  try {
    const audio = getHtmlAudio()
    audio.volume = 0.01
    await audio.play()
    audio.pause()
    audio.currentTime = 0
    audio.volume = 1
    unlocked = true
  } catch {
    // ignore
  }

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

async function playViaWebAudio(): Promise<boolean> {
  const ctx = getAudioContext()
  if (!ctx) return false
  try {
    if (ctx.state === 'suspended') {
      await ctx.resume()
    }
    if (ctx.state !== 'running') return false
    const t = ctx.currentTime + 0.01
    tone(ctx, 880, t, 0.14)
    tone(ctx, 1175, t + 0.15, 0.18)
    return true
  } catch {
    return false
  }
}

async function playViaHtmlAudio(): Promise<boolean> {
  try {
    const audio = getHtmlAudio()
    audio.currentTime = 0
    audio.volume = 1
    await audio.play()
    return true
  } catch {
    return false
  }
}

export async function playNewLeadSound(): Promise<boolean> {
  await unlockNewLeadSound()
  const webOk = await playViaWebAudio()
  const htmlOk = await playViaHtmlAudio()
  return webOk || htmlOk
}

function isCrmWindowInBackground(): boolean {
  if (typeof document === 'undefined') return false
  if (document.visibilityState !== 'visible') return true
  // Other app/browser focused, but Yandex still reports tab as "visible".
  if (typeof document.hasFocus === 'function' && !document.hasFocus()) return true
  return false
}

function showOsNotification(title: string, body: string, tagPrefix: string): void {
  if (typeof Notification === 'undefined') return
  if (Notification.permission !== 'granted') return

  try {
    const notification = new Notification(title, {
      body,
      tag: `${tagPrefix}-${Date.now()}`,
      silent: false,
      requireInteraction: false,
    })
    notification.onclick = () => {
      try {
        window.focus()
      } catch {
        // ignore
      }
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
 * Always try page sound. If CRM window is not focused (other tab/app/browser),
 * also show an OS notification — Yandex often mutes page audio in that case.
 */
export async function notifyNewLeadArrival(info: NewLeadNotifyInfo = {}): Promise<void> {
  const inBackground = isCrmWindowInBackground()
  const played = await playNewLeadSound()

  if (!inBackground && played) return

  const name = [info.firstName, info.patronymic].filter(Boolean).join(' ').trim()
  const source = (info.trafficSource ?? '').trim()
  const numberLabel =
    typeof info.leadNumber === 'number' && info.leadNumber > 0 ? `#${info.leadNumber}` : ''

  const title = numberLabel ? `Новый лид ${numberLabel}` : 'Новый лид'
  const bodyParts = [name || 'Без имени', source].filter(Boolean)
  showOsNotification(
    title,
    bodyParts.join(' · ') || 'Появился новый лид в канбане',
    'proclients-new-lead',
  )
}

export type NewChatMessageNotifyInfo = {
  nickname?: string
  preview?: string
}

/** Sound/OS alert for an incoming message in an already existing chat (not a brand-new lead). */
export async function notifyNewChatMessageArrival(
  info: NewChatMessageNotifyInfo = {},
): Promise<void> {
  const inBackground = isCrmWindowInBackground()
  const played = await playNewLeadSound()

  if (!inBackground && played) return

  const nickname = (info.nickname ?? '').trim() || 'Клиент Авито'
  const preview = (info.preview ?? '').trim()
  showOsNotification(
    'Новое сообщение в чате',
    preview ? `${nickname}: ${preview}` : nickname,
    'proclients-chat-message',
  )
}

export function isNewLeadSoundUnlocked(): boolean {
  return unlocked
}

/** Resume audio when user returns to the CRM tab/window. */
export function bindNewLeadSoundVisibility(): () => void {
  const resume = () => {
    if (document.visibilityState === 'visible') {
      void unlockNewLeadSound()
    }
  }
  document.addEventListener('visibilitychange', resume)
  window.addEventListener('focus', resume)
  return () => {
    document.removeEventListener('visibilitychange', resume)
    window.removeEventListener('focus', resume)
  }
}
