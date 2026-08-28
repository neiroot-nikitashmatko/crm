import { ref } from 'vue'
import { fetchDealsTrafficList } from '@/api/analytics'
import { useAnalyticsPeriodContext } from '@/composables/useAnalyticsPeriod'
import type { DealTrafficListItem } from '@/types/analytics'

export function useDealsTrafficList() {
  const { selectedRange } = useAnalyticsPeriodContext()

  const deals = ref<DealTrafficListItem[]>([])
  const isLoading = ref(false)
  const errorMessage = ref('')

  let requestSeq = 0

  async function loadDeals() {
    const seq = ++requestSeq
    isLoading.value = true
    errorMessage.value = ''

    try {
      const items = await fetchDealsTrafficList(selectedRange.value)
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

  function showEmpty() {
    requestSeq++
    deals.value = []
    errorMessage.value = ''
    isLoading.value = false
  }

  return {
    deals,
    isLoading,
    errorMessage,
    loadDeals,
    showEmpty,
  }
}
