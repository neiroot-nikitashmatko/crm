import { ref, watch } from 'vue'
import { AnalyticsApiError, fetchClosedDealsEmployeeShare } from '@/api/analytics'
import { buildEmployeeShareMetrics } from '@/constants/analytics'
import { useAnalyticsPeriodContext } from '@/composables/useAnalyticsPeriod'
import type { TrafficSourceMetric } from '@/types/analytics'

export function useEmployeeShareAnalytics() {
  const { selectedRange } = useAnalyticsPeriodContext()

  const metrics = ref<TrafficSourceMetric[]>(buildEmployeeShareMetrics([]))
  const isLoading = ref(false)
  const errorMessage = ref('')

  let requestSeq = 0

  async function loadMetrics() {
    const seq = ++requestSeq
    isLoading.value = true
    errorMessage.value = ''

    try {
      const items = await fetchClosedDealsEmployeeShare(selectedRange.value)
      if (seq !== requestSeq) return

      metrics.value = buildEmployeeShareMetrics(items)
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
