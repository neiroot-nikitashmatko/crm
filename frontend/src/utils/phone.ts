export const PHONE_PREFIX = '+7'

export function normalizePhone(input: string): string {
  const digits = input.replace(/\D/g, '')
  const normalizedDigits = digits.startsWith('8')
    ? `7${digits.slice(1)}`
    : digits.startsWith('7')
      ? digits
      : `7${digits}`

  return `+${normalizedDigits.slice(0, 11)}`
}

export function isPhoneFilled(value: string): boolean {
  return /^\+7\d{10}$/.test(value)
}
