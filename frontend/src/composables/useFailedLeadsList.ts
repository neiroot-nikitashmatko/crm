import { ref } from 'vue'
import { fetchFailedLeadsList } from '@/api/analytics'
import { useAnalyticsPeriodContext } from '@/composables/useAnalyticsPeriod'
import type { FailedLeadListItem } from '@/types/analytics'

export function useFailedLeadsList() {
  const { selectedRange } = useAnalyticsPeriodContext()

  const leads = ref<FailedLeadListItem[]>([])
  const isLoading = ref(false)
  const errorMessage = ref('')

  let requestSeq = 0

  async function loadFailedLeads() {
    const seq = ++requestSeq
    isLoading.value = true
    errorMessage.value = ''

    try {
      const items = await fetchFailedLeadsList(selectedRange.value)
      if (seq !== requestSeq) return
      leads.value = items
    } catch {
      if (seq !== requestSeq) return
      leads.value = []
      errorMessage.value = 'Не удалось загрузить список лидов'
    } finally {
      if (seq === requestSeq) {
        isLoading.value = false
      }
    }
  }

  return {
    leads,
    isLoading,
    errorMessage,
    loadFailedLeads,
  }
}
