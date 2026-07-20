/** Short notification chime via Web Audio (no asset file). */

let audioContext: AudioContext | null = null
let unlocked = false

function getAudioContext(): AudioContext | null {
  const Ctx = window.AudioContext || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
  if (!Ctx) return null
  if (!audioContext) {
    audioContext = new Ctx()
  }
  return audioContext
}

/** Call once after any user gesture so browsers allow playback. */
export async function unlockNewLeadSound(): Promise<void> {
  const ctx = getAudioContext()
  if (!ctx) return
  if (ctx.state === 'suspended') {
    try {
      await ctx.resume()
    } catch {
      return
    }
  }
  unlocked = ctx.state === 'running'
}

function tone(ctx: AudioContext, frequency: number, startAt: number, duration: number) {
  const oscillator = ctx.createOscillator()
  const gain = ctx.createGain()
  oscillator.type = 'sine'
  oscillator.frequency.value = frequency
  gain.gain.setValueAtTime(0.0001, startAt)
  gain.gain.exponentialRampToValueAtTime(0.12, startAt + 0.02)
  gain.gain.exponentialRampToValueAtTime(0.0001, startAt + duration)
  oscillator.connect(gain)
  gain.connect(ctx.destination)
  oscillator.start(startAt)
  oscillator.stop(startAt + duration + 0.02)
}

/** Plays when a new lead appears in column «Новый лид». */
export async function playNewLeadSound(): Promise<void> {
  const ctx = getAudioContext()
  if (!ctx) return

  if (ctx.state === 'suspended') {
    try {
      await ctx.resume()
    } catch {
      return
    }
  }
  if (ctx.state !== 'running') return
  unlocked = true

  const t = ctx.currentTime
  tone(ctx, 880, t, 0.12)
  tone(ctx, 1175, t + 0.14, 0.16)
}

export function isNewLeadSoundUnlocked(): boolean {
  return unlocked
}
