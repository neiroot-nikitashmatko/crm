import { ref } from 'vue'
import { fetchTradeProfitItems } from '@/api/analytics'
import { useAnalyticsPeriodContext } from '@/composables/useAnalyticsPeriod'
import type { TradeProfitItem } from '@/types/analytics'

export function useTradeProfitItems() {
  const { selectedRange } = useAnalyticsPeriodContext()

  const items = ref<TradeProfitItem[]>([])
  const isLoading = ref(false)
  const errorMessage = ref('')

  let requestSeq = 0

  async function loadItems() {
    const seq = ++requestSeq
    isLoading.value = true
    errorMessage.value = ''

    try {
      const nextItems = await fetchTradeProfitItems(selectedRange.value)
      if (seq !== requestSeq) return
      items.value = nextItems
    } catch {
      if (seq !== requestSeq) return
      items.value = []
      errorMessage.value = 'Не удалось загрузить список товаров'
    } finally {
      if (seq === requestSeq) {
        isLoading.value = false
      }
    }
  }

  return {
    items,
    isLoading,
    errorMessage,
    loadItems,
  }
}
