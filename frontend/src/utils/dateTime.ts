const DATE_FORMATTER = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
})

export function getTodayDateRange(referenceDate = new Date()): [number, number] {
  const start = new Date(referenceDate)
  start.setHours(0, 0, 0, 0)

  const end = new Date(referenceDate)
  end.setHours(23, 59, 59, 999)

  return [start.getTime(), end.getTime()]
}

export function getCurrentMonthDateRange(referenceDate = new Date()): [number, number] {
  const start = new Date(referenceDate.getFullYear(), referenceDate.getMonth(), 1)
  const end = new Date(referenceDate)
  end.setHours(23, 59, 59, 999)

  return [start.getTime(), end.getTime()]
}

export function formatDateRange([start, end]: [number, number]): string {
  const startDate = new Date(start)
  const endDate = new Date(end)

  if (
    startDate.getFullYear() === endDate.getFullYear() &&
    startDate.getMonth() === endDate.getMonth() &&
    startDate.getDate() === endDate.getDate()
  ) {
    return DATE_FORMATTER.format(startDate)
  }

  return `${DATE_FORMATTER.format(startDate)} — ${DATE_FORMATTER.format(endDate)}`
}
