import { ref } from 'vue'
import { fetchClosedDealsList } from '@/api/analytics'
import { useAnalyticsPeriodContext } from '@/composables/useAnalyticsPeriod'
import {
  PRODUCTION_EMPLOYEE_NAMES,
  PRODUCTION_NOMENCLATURE_OPTIONS,
  productionCategoryForNomenclature,
} from '@/constants/production'
import type { ClosedDealListItem } from '@/types/analytics'

const MOCK_CLOSED_DEALS_COUNT = 30

const MOCK_CLIENT_NAMES = [
  { firstName: 'Иван', patronymic: 'Сергеевич' },
  { firstName: 'Алексей', patronymic: 'Петрович' },
  { firstName: 'Дмитрий', patronymic: 'Андреевич' },
  { firstName: 'Сергей', patronymic: 'Николаевич' },
  { firstName: 'Андрей', patronymic: 'Викторович' },
  { firstName: 'Максим', patronymic: 'Олегович' },
  { firstName: 'Павел', patronymic: 'Игоревич' },
  { firstName: 'Никита', patronymic: 'Александрович' },
  { firstName: 'Егор', patronymic: 'Дмитриевич' },
  { firstName: 'Роман', patronymic: 'Валерьевич' },
] as const

function buildMockClosedDeals(count: number, startNumber: number): ClosedDealListItem[] {
  const now = Date.now()

  return Array.from({ length: count }, (_, index) => {
    const client = MOCK_CLIENT_NAMES[index % MOCK_CLIENT_NAMES.length]
    const nomenclature = PRODUCTION_NOMENCLATURE_OPTIONS[index % PRODUCTION_NOMENCLATURE_OPTIONS.length]
    const employee = PRODUCTION_EMPLOYEE_NAMES[index % PRODUCTION_EMPLOYEE_NAMES.length]

    return {
      id: `mock-closed-deal-${index + 1}`,
      dealNumber: startNumber + index,
      firstName: client.firstName,
      patronymic: client.patronymic,
      phone: `+7 961 300-${String(10 + index).padStart(2, '0')}-${String(16 + index).padStart(2, '0')}`,
      nomenclature: nomenclature.value,
      category: productionCategoryForNomenclature(nomenclature.value),
      employee,
      createdAt: now - index * 36 * 60 * 60 * 1000,
    }
  })
}

function withMockClosedDeals(items: ClosedDealListItem[]): ClosedDealListItem[] {
  if (items.length >= MOCK_CLOSED_DEALS_COUNT) return items

  const lastDealNumber = items[items.length - 1]?.dealNumber ?? 1000
  return [
    ...items,
    ...buildMockClosedDeals(MOCK_CLOSED_DEALS_COUNT - items.length, lastDealNumber + 1),
  ]
}

export function useClosedDealsList() {
  const { selectedRange } = useAnalyticsPeriodContext()

  const deals = ref<ClosedDealListItem[]>([])
  const isLoading = ref(false)
  const errorMessage = ref('')

  let requestSeq = 0

  async function loadDeals(options?: { requireEmployee?: boolean }) {
    const seq = ++requestSeq
    isLoading.value = true
    errorMessage.value = ''

    try {
      const items = await fetchClosedDealsList(selectedRange.value, options)
      if (seq !== requestSeq) return
      deals.value = withMockClosedDeals(items)
    } catch {
      if (seq !== requestSeq) return
      deals.value = withMockClosedDeals([])
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
  }
}
