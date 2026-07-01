export type ProductionCalendarViewMode = 'month' | 'week'

export const PRODUCTION_CALENDAR_VIEW_OPTIONS: Array<{
  value: ProductionCalendarViewMode
  label: string
}> = [
  { value: 'month', label: 'Месяц' },
  { value: 'week', label: 'Неделя' },
]
