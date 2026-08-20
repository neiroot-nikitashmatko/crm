import { computed } from 'vue'
import { useIncomingInvoices } from '@/composables/useIncomingInvoices'
import { useOutgoingInvoices } from '@/composables/useOutgoingInvoices'
import { useProductsCatalog } from '@/composables/useProductsCatalog'
import type { StockBalance } from '@/types/stockBalance'

interface StockMovementItem {
  catalogProductId?: string
  title: string
  quantity: number
}

function movementKey(item: StockMovementItem) {
  if (item.catalogProductId) return `id:${item.catalogProductId}`
  return `title:${item.title.trim().toLowerCase()}`
}

function safeQuantity(value: number) {
  const quantity = Number(value)
  return Number.isFinite(quantity) ? quantity : 0
}

export function useStockBalances() {
  const { invoices: incomingInvoices } = useIncomingInvoices()
  const { invoices: outgoingInvoices } = useOutgoingInvoices()
  const { getCatalogProductById, loadCatalog } = useProductsCatalog()

  const balances = computed(() => {
    const quantities = new Map<string, StockMovementItem>()

    function applyMovement(item: StockMovementItem, sign: 1 | -1) {
      const title = item.title.trim()
      if (!title && !item.catalogProductId) return

      const key = movementKey(item)
      const current = quantities.get(key) ?? {
        catalogProductId: item.catalogProductId,
        title,
        quantity: 0,
      }
      current.quantity += sign * safeQuantity(item.quantity)
      if (title) current.title = title
      if (item.catalogProductId) current.catalogProductId = item.catalogProductId
      quantities.set(key, current)
    }

    for (const invoice of incomingInvoices.value) {
      for (const item of invoice.items) applyMovement(item, 1)
    }

    for (const invoice of outgoingInvoices.value) {
      for (const item of invoice.items) applyMovement(item, -1)
    }

    return Array.from(quantities.entries())
      .map(([key, item]) => {
        const catalogProduct = item.catalogProductId ? getCatalogProductById(item.catalogProductId) : null

        return {
          key,
          catalogProductId: item.catalogProductId,
          title: catalogProduct?.name ?? item.title,
          sku: catalogProduct?.sku ?? '',
          category: catalogProduct?.category ?? '',
          quantity: item.quantity,
        } satisfies StockBalance
      })
      .filter((item) => item.quantity !== 0)
      .sort((left, right) => left.title.localeCompare(right.title, 'ru'))
  })

  const categoryGroups = computed(() => {
    const groupsMap = new Map<string, StockBalance[]>()

    for (const item of balances.value) {
      const category = item.category.trim()
      if (!category) continue

      if (!groupsMap.has(category)) {
        groupsMap.set(category, [])
      }
      groupsMap.get(category)?.push(item)
    }

    return Array.from(groupsMap.entries())
      .map(([category, items]) => ({
        category,
        items,
      }))
      .sort((left, right) => left.category.localeCompare(right.category, 'ru'))
  })

  const uncategorizedItems = computed(() =>
    balances.value.filter((item) => item.category.trim().length === 0),
  )

  return {
    balances,
    categoryGroups,
    uncategorizedItems,
    loadCatalog,
  }
}
