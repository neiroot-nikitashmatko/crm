import { computed, inject, ref, type InjectionKey } from 'vue'
import {
  formatDateRange,
  getCurrentMonthDateRange,
  getTodayDateRange,
} from '@/utils/dateTime'
import type { AnalyticsDateRange, AnalyticsPeriodPreset } from '@/types/analytics'

export function useAnalyticsPeriod() {
  const selectedRange = ref<AnalyticsDateRange>(getCurrentMonthDateRange())
  const selectedPreset = ref<AnalyticsPeriodPreset>('month')

  const periodLabel = computed(() => {
    if (selectedPreset.value === 'today') return 'За сегодня'
    if (selectedPreset.value === 'month') return 'За месяц'
    return formatDateRange(selectedRange.value)
  })

  function setPeriod(range: AnalyticsDateRange, preset: AnalyticsPeriodPreset) {
    const [start, end] = range
    selectedRange.value = start <= end ? [start, end] : [end, start]
    selectedPreset.value = preset
  }

  function selectToday() {
    setPeriod(getTodayDateRange(), 'today')
  }

  function selectMonth() {
    setPeriod(getCurrentMonthDateRange(), 'month')
  }

  function selectCustomRange(range: AnalyticsDateRange) {
    const [start, end] = range
    const startDate = new Date(start)
    const endDate = new Date(end)
    startDate.setHours(0, 0, 0, 0)
    endDate.setHours(23, 59, 59, 999)

    setPeriod([startDate.getTime(), endDate.getTime()], 'custom')
  }

  function updateCustomPeriodBoundary(boundary: 0 | 1, value: number | null) {
    if (value === null) return

    const nextRange: AnalyticsDateRange = [selectedRange.value[0], selectedRange.value[1]]
    nextRange[boundary] = value

    if (boundary === 0 && nextRange[0] > nextRange[1]) {
      nextRange[1] = nextRange[0]
    }
    if (boundary === 1 && nextRange[1] < nextRange[0]) {
      nextRange[0] = nextRange[1]
    }

    selectCustomRange(nextRange)
  }

  const periodStart = computed<number | null>({
    get: () => selectedRange.value[0],
    set: (value) => updateCustomPeriodBoundary(0, value),
  })

  const periodEnd = computed<number | null>({
    get: () => selectedRange.value[1],
    set: (value) => updateCustomPeriodBoundary(1, value),
  })

  return {
    selectedRange,
    selectedPreset,
    periodLabel,
    periodStart,
    periodEnd,
    selectToday,
    selectMonth,
    selectCustomRange,
  }
}

export type AnalyticsPeriodContext = ReturnType<typeof useAnalyticsPeriod>

export const ANALYTICS_PERIOD_KEY: InjectionKey<AnalyticsPeriodContext> = Symbol('analyticsPeriod')

export function useAnalyticsPeriodContext(): AnalyticsPeriodContext {
  const context = inject(ANALYTICS_PERIOD_KEY)

  if (!context) {
    throw new Error('Контекст периода аналитики не предоставлен')
  }

  return context
}
