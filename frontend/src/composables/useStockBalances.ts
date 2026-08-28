import { computed } from 'vue'
import { useIncomingInvoices } from '@/composables/useIncomingInvoices'
import { useOutgoingInvoices } from '@/composables/useOutgoingInvoices'
import { useProductsCatalog } from '@/composables/useProductsCatalog'
import type { StockBalance } from '@/types/stockBalance'
import {
  safeStockQuantity,
  stockMovementKey,
  type StockMovementItem,
  validateOutgoingInvoiceStock,
} from '@/utils/stockBalances'

export function useStockBalances() {
  const { invoices: incomingInvoices, isLoading: incomingLoading, loadInvoices: loadIncomingInvoices } =
    useIncomingInvoices()
  const { invoices: outgoingInvoices, isLoading: outgoingLoading, loadInvoices: loadOutgoingInvoices } =
    useOutgoingInvoices()
  const { getCatalogProductById, loadCatalog, isLoading: catalogLoading } = useProductsCatalog()

  function applyMovement(quantities: Map<string, StockMovementItem>, item: StockMovementItem, sign: 1 | -1) {
    const title = item.title.trim()
    if (!title && !item.catalogProductId) return

    const key = stockMovementKey(item)
    const current = quantities.get(key) ?? {
      catalogProductId: item.catalogProductId,
      title,
      quantity: 0,
    }
    current.quantity += sign * safeStockQuantity(item.quantity)
    if (title) current.title = title
    if (item.catalogProductId) current.catalogProductId = item.catalogProductId
    quantities.set(key, current)
  }

  function buildStockQuantities(excludeOutgoingInvoiceId?: string | null) {
    const quantities = new Map<string, StockMovementItem>()

    for (const invoice of incomingInvoices.value) {
      for (const item of invoice.items) applyMovement(quantities, item, 1)
    }

    for (const invoice of outgoingInvoices.value) {
      if (excludeOutgoingInvoiceId && invoice.id === excludeOutgoingInvoiceId) continue
      for (const item of invoice.items) applyMovement(quantities, item, -1)
    }

    return quantities
  }

  function resolveMovementTitle(item: StockMovementItem) {
    if (item.catalogProductId) {
      return getCatalogProductById(item.catalogProductId)?.name ?? item.title
    }
    return item.title
  }

  function getOutgoingInvoiceStockIssueKeys(
    items: StockMovementItem[],
    excludeOutgoingInvoiceId?: string | null,
  ) {
    const quantities = buildStockQuantities(excludeOutgoingInvoiceId)
    const availableQuantities = new Map<string, number>()

    for (const [key, item] of quantities) {
      availableQuantities.set(key, item.quantity)
    }

    const issues = validateOutgoingInvoiceStock(items, availableQuantities, resolveMovementTitle)
    return new Set(issues.map((issue) => issue.key))
  }

  const balances = computed(() => {
    const quantities = buildStockQuantities()

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

  async function loadStockData() {
    await Promise.all([loadCatalog(), loadIncomingInvoices(), loadOutgoingInvoices()])
  }

  const isLoading = computed(
    () => incomingLoading.value || outgoingLoading.value || catalogLoading.value,
  )

  return {
    balances,
    categoryGroups,
    uncategorizedItems,
    isLoading,
    getOutgoingInvoiceStockIssueKeys,
    loadCatalog,
    loadStockData,
  }
}
