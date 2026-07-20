/** Short notification chime via Web Audio (no asset file). */

let audioContext: AudioContext | null = null
let unlocked = false

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

/** Call after a user gesture so Chrome allows playback. */
export async function unlockNewLeadSound(): Promise<void> {
  const ctx = getAudioContext()
  if (!ctx) return

  try {
    if (ctx.state === 'suspended') {
      await ctx.resume()
    }
    // Chrome often needs an actual node start inside a gesture to fully unlock.
    const buffer = ctx.createBuffer(1, 1, ctx.sampleRate)
    const source = ctx.createBufferSource()
    source.buffer = buffer
    source.connect(ctx.destination)
    source.start(0)
  } catch {
    return
  }

  unlocked = ctx.state === 'running'
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

/** Plays when a new lead appears in column «Новый лид». */
export async function playNewLeadSound(): Promise<void> {
  await unlockNewLeadSound()
  const ctx = getAudioContext()
  if (!ctx || ctx.state !== 'running') return

  const t = ctx.currentTime + 0.01
  tone(ctx, 880, t, 0.14)
  tone(ctx, 1175, t + 0.15, 0.18)
}

export function isNewLeadSoundUnlocked(): boolean {
  return unlocked
}
