import { ref, watch } from 'vue'
import {
  AnalyticsApiError,
  fetchDealTrafficMetrics,
  fetchLeadTrafficMetrics,
} from '@/api/analytics'
import { buildTrafficSourceMetrics } from '@/constants/analytics'
import { useAnalyticsPeriodContext } from '@/composables/useAnalyticsPeriod'
import type { AnalyticsDateRange, TrafficSourceCount, TrafficSourceMetric } from '@/types/analytics'

function useTrafficSourceAnalytics(
  fetchItems: (range: AnalyticsDateRange) => Promise<TrafficSourceCount[]>,
) {
  const { selectedRange } = useAnalyticsPeriodContext()

  const metrics = ref<TrafficSourceMetric[]>(buildTrafficSourceMetrics([]))
  const isLoading = ref(false)
  const errorMessage = ref('')

  let requestSeq = 0

  async function loadMetrics() {
    const seq = ++requestSeq
    isLoading.value = true
    errorMessage.value = ''

    try {
      const items = await fetchItems(selectedRange.value)
      if (seq !== requestSeq) return

      metrics.value = buildTrafficSourceMetrics(items)
    } catch (error) {
      if (seq !== requestSeq) return
      errorMessage.value =
        error instanceof AnalyticsApiError ? error.message : 'Не удалось загрузить аналитику'
    } finally {
      if (seq === requestSeq) {
        isLoading.value = false
      }
    }
  }

  watch(
    selectedRange,
    () => {
      void loadMetrics()
    },
    { immediate: true },
  )

  return {
    metrics,
    isLoading,
    errorMessage,
    reload: loadMetrics,
  }
}

export function useLeadTrafficAnalytics() {
  return useTrafficSourceAnalytics(fetchLeadTrafficMetrics)
}

export function useDealTrafficAnalytics() {
  return useTrafficSourceAnalytics(fetchDealTrafficMetrics)
}
