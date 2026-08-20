export type PaymentPayerId = 'ip-panov-nikolay' | 'ip-shmatko-nikita' | 'ip-panov-dmitry'

export const PAYMENT_PAYER_OPTIONS: Array<{ label: string; value: PaymentPayerId }> = [
  { label: 'ИП Панов Николай Константинович', value: 'ip-panov-nikolay' },
  { label: 'ИП Шматко Никита Сергеевич', value: 'ip-shmatko-nikita' },
  { label: 'ИП Панов Дмитрий Константинович', value: 'ip-panov-dmitry' },
]

export function getPaymentPayerLabel(payerId: string | null) {
  if (!payerId) return ''
  return PAYMENT_PAYER_OPTIONS.find((option) => option.value === payerId)?.label ?? payerId
}

export const PAYMENT_SHORT_TITLE_MAX_LENGTH = 22
