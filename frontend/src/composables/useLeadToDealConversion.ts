import { ref, watch } from 'vue'
import { AnalyticsApiError, fetchLeadToDealConversion } from '@/api/analytics'
import { useAnalyticsPeriodContext } from '@/composables/useAnalyticsPeriod'
import type { LeadToDealConversion } from '@/types/analytics'

const emptyConversion = (): LeadToDealConversion => ({
  leadsCount: 0,
  convertedCount: 0,
  percent: 0,
})

export function useLeadToDealConversion() {
  const { selectedRange } = useAnalyticsPeriodContext()

  const conversion = ref<LeadToDealConversion>(emptyConversion())
  const isLoading = ref(false)
  const errorMessage = ref('')

  let requestSeq = 0

  async function loadConversion() {
    const seq = ++requestSeq
    isLoading.value = true
    errorMessage.value = ''

    try {
      const item = await fetchLeadToDealConversion(selectedRange.value)
      if (seq !== requestSeq) return
      conversion.value = item
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
      void loadConversion()
    },
    { immediate: true },
  )

  return {
    conversion,
    isLoading,
    errorMessage,
  }
}
