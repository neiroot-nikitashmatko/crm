import { ref } from 'vue'
import { fetchClosedDealsList, fetchFailedDealsList } from '@/api/analytics'
import { useAnalyticsPeriodContext } from '@/composables/useAnalyticsPeriod'
import type { ClosedDealListItem } from '@/types/analytics'

export function useClosedDealsList() {
  const { selectedRange } = useAnalyticsPeriodContext()

  const deals = ref<ClosedDealListItem[]>([])
  const isLoading = ref(false)
  const errorMessage = ref('')

  let requestSeq = 0

  async function loadDeals(options?: { requireEmployee?: boolean; requireProduction?: boolean }) {
    const seq = ++requestSeq
    isLoading.value = true
    errorMessage.value = ''

    try {
      const items = await fetchClosedDealsList(selectedRange.value, options)
      if (seq !== requestSeq) return
      deals.value = items
    } catch {
      if (seq !== requestSeq) return
      deals.value = []
      errorMessage.value = 'Не удалось загрузить список сделок'
    } finally {
      if (seq === requestSeq) {
        isLoading.value = false
      }
    }
  }

  async function loadFailedDeals() {
    const seq = ++requestSeq
    isLoading.value = true
    errorMessage.value = ''

    try {
      const items = await fetchFailedDealsList(selectedRange.value)
      if (seq !== requestSeq) return
      deals.value = items
    } catch {
      if (seq !== requestSeq) return
      deals.value = []
      errorMessage.value = 'Не удалось загрузить список сделок'
    } finally {
      if (seq === requestSeq) {
        isLoading.value = false
      }
    }
  }

  return {
    deals,
    isLoading,
    errorMessage,
    loadDeals,
    loadFailedDeals,
  }
}
