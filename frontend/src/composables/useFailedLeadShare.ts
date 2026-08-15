import { ref, watch } from 'vue'
import { AnalyticsApiError, fetchFailedLeadShare } from '@/api/analytics'
import { useAnalyticsPeriodContext } from '@/composables/useAnalyticsPeriod'
import type { FailedLeadShare } from '@/types/analytics'

const emptyShare = (): FailedLeadShare => ({
  leadsCount: 0,
  failedCount: 0,
  percent: 0,
})

export function useFailedLeadShare() {
  const { selectedRange } = useAnalyticsPeriodContext()

  const share = ref<FailedLeadShare>(emptyShare())
  const isLoading = ref(false)
  const errorMessage = ref('')

  let requestSeq = 0

  async function loadShare() {
    const seq = ++requestSeq
    isLoading.value = true
    errorMessage.value = ''

    try {
      const item = await fetchFailedLeadShare(selectedRange.value)
      if (seq !== requestSeq) return
      share.value = item
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
      void loadShare()
    },
    { immediate: true },
  )

  return {
    share,
    isLoading,
    errorMessage,
  }
}
