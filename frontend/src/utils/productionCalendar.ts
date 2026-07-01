import type { Deal } from '@/types/deal'

export interface ProductionCalendarEntry {
  dealId: string
  nomenclature: string
  dueAt: number
  employee: string
}

export function getDateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function isDealProductionComplete(deal: Deal): boolean {
  const { nomenclature, dueAt, employee } = deal.production
  return (
    nomenclature.trim() !== '' &&
    dueAt !== null &&
    !Number.isNaN(dueAt) &&
    employee.trim() !== ''
  )
}

function isDealInProductionStatus(deal: Deal): boolean {
  if (deal.columnId === 'production') {
    return true
  }

  const status = String(deal.status ?? 'today').toLowerCase()
  return status === 'today'
}

export function shouldShowDealInProductionCalendar(deal: Deal): boolean {
  return isDealProductionComplete(deal) && isDealInProductionStatus(deal)
}

export function formatProductionTime(dueAt: number): string {
  return new Date(dueAt).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function buildProductionEntriesByDay(deals: Deal[]): Map<string, ProductionCalendarEntry[]> {
  const entriesByDay = new Map<string, ProductionCalendarEntry[]>()

  for (const deal of deals) {
    if (!shouldShowDealInProductionCalendar(deal)) continue

    const dueAt = deal.production.dueAt!
    const dayKey = getDateKey(new Date(dueAt))
    const entry: ProductionCalendarEntry = {
      dealId: deal.id,
      nomenclature: deal.production.nomenclature.trim(),
      dueAt,
      employee: deal.production.employee.trim(),
    }

    const dayEntries = entriesByDay.get(dayKey) ?? []
    dayEntries.push(entry)
    entriesByDay.set(dayKey, dayEntries)
  }

  return sortProductionEntriesByDay(entriesByDay)
}

function sortProductionEntriesByDay(entriesByDay: Map<string, ProductionCalendarEntry[]>) {
  for (const dayEntries of entriesByDay.values()) {
    dayEntries.sort((left, right) => left.dueAt - right.dueAt)
  }

  return entriesByDay
}
