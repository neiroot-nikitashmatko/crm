export type PaymentCalendarViewMode = 'month' | 'week'

export const PAYMENT_CALENDAR_VIEW_OPTIONS: Array<{
  value: PaymentCalendarViewMode
  label: string
}> = [
  { value: 'month', label: 'Месяц' },
  { value: 'week', label: 'Неделя' },
]
