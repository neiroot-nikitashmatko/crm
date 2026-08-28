import { computed, ref, watch } from 'vue'
import { AnalyticsApiError, fetchTradeProfit } from '@/api/analytics'
import { useAnalyticsPeriodContext } from '@/composables/useAnalyticsPeriod'
import type { TradeProfit } from '@/types/analytics'

const emptyProfit = (): TradeProfit => ({
  profit: 0,
  revenue: 0,
  cost: 0,
  invoicesCount: 0,
})

const emptyPeriodHint = 'Нет продаж за выбранный период'

export function useTradeProfit() {
  const { selectedRange } = useAnalyticsPeriodContext()

  const profit = ref<TradeProfit>(emptyProfit())
  const isLoading = ref(false)
  const errorMessage = ref('')

  const hasSales = computed(() => profit.value.invoicesCount > 0)

  const profitAmount = computed(() => (hasSales.value ? profit.value.profit : null))
  const revenueAmount = computed(() => (hasSales.value ? profit.value.revenue : null))

  const marginPercent = computed(() => {
    if (!hasSales.value || profit.value.revenue === 0) return null
    return (profit.value.profit / profit.value.revenue) * 100
  })

  const markupPercent = computed(() => {
    if (!hasSales.value || profit.value.cost === 0) return null
    return (profit.value.profit / profit.value.cost) * 100
  })

  const emptyHint = computed(() => (hasSales.value ? '' : emptyPeriodHint))

  const markupHint = computed(() => {
    if (!hasSales.value) return emptyPeriodHint
    if (profit.value.cost === 0) return 'Нет закупочной цены'
    return ''
  })

  let requestSeq = 0

  async function loadProfit() {
    const seq = ++requestSeq
    isLoading.value = true
    errorMessage.value = ''

    try {
      const item = await fetchTradeProfit(selectedRange.value)
      if (seq !== requestSeq) return
      profit.value = item
    } catch (error) {
      if (seq !== requestSeq) return
      profit.value = emptyProfit()
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
      void loadProfit()
    },
    { immediate: true },
  )

  return {
    profit,
    profitAmount,
    revenueAmount,
    marginPercent,
    markupPercent,
    emptyHint,
    markupHint,
    isLoading,
    errorMessage,
  }
}
