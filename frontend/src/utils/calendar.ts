export interface CalendarDayCell {
  date: Date
  dayNumber: number
  isCurrentMonth: boolean
  isToday: boolean
}

const WEEKDAY_LABELS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

function getMondayBasedWeekday(date: Date) {
  const day = date.getDay()
  return day === 0 ? 6 : day - 1
}

function isSameDay(left: Date, right: Date) {
  return (
    left.getFullYear() === right.getFullYear() &&
    left.getMonth() === right.getMonth() &&
    left.getDate() === right.getDate()
  )
}

function getStartOfWeek(date: Date) {
  const monday = new Date(date)
  monday.setDate(date.getDate() - getMondayBasedWeekday(date))
  monday.setHours(0, 0, 0, 0)
  return monday
}

export function buildCurrentWeekGrid(referenceDate = new Date()): CalendarDayCell[] {
  const monday = getStartOfWeek(referenceDate)
  const referenceMonth = referenceDate.getMonth()

  return Array.from({ length: 7 }, (_, index) => {
    const date = new Date(monday)
    date.setDate(monday.getDate() + index)

    return {
      date,
      dayNumber: date.getDate(),
      isCurrentMonth: date.getMonth() === referenceMonth,
      isToday: isSameDay(date, referenceDate),
    }
  })
}

function formatRussianMonthGenitive(date: Date) {
  return (
    new Intl.DateTimeFormat('ru-RU', {
      day: 'numeric',
      month: 'long',
    })
      .formatToParts(date)
      .find((part) => part.type === 'month')?.value ?? ''
  )
}

function formatRussianDayMonth(date: Date) {
  return new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
  }).format(date)
}

function formatRussianDayMonthYear(date: Date) {
  return new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(date)
}

export function formatWeekTitle(weekDays: CalendarDayCell[]) {
  const start = weekDays[0]?.date
  const end = weekDays[weekDays.length - 1]?.date
  if (!start || !end) return ''

  const startDay = start.getDate()
  const endDay = end.getDate()

  if (start.getMonth() === end.getMonth() && start.getFullYear() === end.getFullYear()) {
    const month = formatRussianMonthGenitive(end)
    const year = new Intl.DateTimeFormat('ru-RU', { year: 'numeric' }).format(end)
    const label = `${startDay}–${endDay} ${month} ${year} г.`

    return label.charAt(0).toUpperCase() + label.slice(1)
  }

  return `${formatRussianDayMonth(start)} — ${formatRussianDayMonthYear(end)}`
}

export function buildMonthGrid(year: number, month: number, today = new Date()): CalendarDayCell[] {
  const firstDayOfMonth = new Date(year, month, 1)
  const daysInMonth = new Date(year, month + 1, 0).getDate()
  const leadingEmptyCells = getMondayBasedWeekday(firstDayOfMonth)

  const cells: CalendarDayCell[] = []

  for (let index = leadingEmptyCells - 1; index >= 0; index -= 1) {
    const date = new Date(year, month, -index)
    cells.push({
      date,
      dayNumber: date.getDate(),
      isCurrentMonth: false,
      isToday: isSameDay(date, today),
    })
  }

  for (let day = 1; day <= daysInMonth; day += 1) {
    const date = new Date(year, month, day)
    cells.push({
      date,
      dayNumber: day,
      isCurrentMonth: true,
      isToday: isSameDay(date, today),
    })
  }

  const trailingCells = (7 - (cells.length % 7)) % 7
  for (let day = 1; day <= trailingCells; day += 1) {
    const date = new Date(year, month + 1, day)
    cells.push({
      date,
      dayNumber: day,
      isCurrentMonth: false,
      isToday: isSameDay(date, today),
    })
  }

  return cells
}

export function formatMonthTitle(year: number, month: number) {
  const label = new Intl.DateTimeFormat('ru-RU', {
    month: 'long',
    year: 'numeric',
  }).format(new Date(year, month, 1))

  return label.charAt(0).toUpperCase() + label.slice(1)
}

export function getWeekdayLabels() {
  return WEEKDAY_LABELS
}

export function splitIntoWeeks(days: CalendarDayCell[]) {
  const weeks: CalendarDayCell[][] = []
  for (let index = 0; index < days.length; index += 7) {
    weeks.push(days.slice(index, index + 7))
  }
  return weeks
}
