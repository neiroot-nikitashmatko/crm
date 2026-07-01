import { reactive, watch } from 'vue'
import type { DealProduct } from '@/types/deal'
import type { ProductRow } from '@/types/productRow'
import { useDeals } from '@/composables/useDeals'
import { useProductsCatalog } from '@/composables/useProductsCatalog'
import type { CatalogProduct } from '@/types/productCatalog'
import {
  createEmptyProductRow,
  ensureRowIds,
  productsToRows,
  rowsToDealProducts,
  serializeDealProducts,
  serializeProductRows,
} from '@/utils/products'

const dealProductRows = reactive<Record<string, ProductRow[]>>({})
const savedSnapshots = reactive<Record<string, string>>({})

function resolveCatalogIdByTitleAndPrice(
  title: string,
  unitPrice: number,
  catalog: CatalogProduct[],
): string | undefined {
  const normalizedTitle = title.trim()
  if (!normalizedTitle) return undefined

  const exact = catalog.find((item) => item.name === normalizedTitle && item.cost === unitPrice)
  if (exact) return exact.id

  const titleMatch = catalog.find((item) => item.name === normalizedTitle)
  return titleMatch?.id
}

export function useDealProductRows() {
  const { deals, updateDealProducts } = useDeals()
  const { products: catalogProducts } = useProductsCatalog()

  watch(
    catalogProducts,
    (catalog) => {
      if (!catalog || catalog.length === 0) return

      for (const dealId of Object.keys(dealProductRows)) {
        const rows = dealProductRows[dealId]
        if (!rows || rows.length === 0) continue

        let changed = false
        for (const row of rows) {
          if (row.catalogProductId) continue
          if (!row.title.trim()) continue

          const resolved = resolveCatalogIdByTitleAndPrice(row.title, row.unitPrice, catalog)
          if (resolved) {
            row.catalogProductId = resolved
            changed = true
          }
        }

        if (changed) {
          // триггерим реактивность для потребителей
          dealProductRows[dealId] = [...rows]
        }
      }
    },
    { deep: true },
  )

  function getDealRows(dealId: string): ProductRow[] {
    if (!dealProductRows[dealId]) {
      const deal = deals.value.find((item) => item.id === dealId)
      dealProductRows[dealId] = productsToRows(deal?.products ?? [], catalogProducts.value)
      ensureRowIds(dealProductRows[dealId])
      savedSnapshots[dealId] = serializeDealProducts(deal?.products ?? [])
    }
    return dealProductRows[dealId]
  }

  function hydrateDealRows(dealId: string, force = false) {
    const deal = deals.value.find((item) => item.id === dealId)
    if (!deal) return

    const apiSnapshot = serializeDealProducts(deal.products)
    const hasLocalState = Boolean(dealProductRows[dealId])
    const hasUnsavedChanges =
      hasLocalState && serializeProductRows(dealProductRows[dealId]) !== savedSnapshots[dealId]

    if (force || !hasLocalState || (!hasUnsavedChanges && savedSnapshots[dealId] !== apiSnapshot)) {
      dealProductRows[dealId] = productsToRows(deal.products, catalogProducts.value)
      ensureRowIds(dealProductRows[dealId])
      savedSnapshots[dealId] = apiSnapshot
    }
  }

  function resetDealRows(dealId: string) {
    delete dealProductRows[dealId]
    delete savedSnapshots[dealId]
  }

  async function saveDealProductRows(dealId: string) {
    const rows = dealProductRows[dealId] ?? []
    const products = rowsToDealProducts(rows)
    const currentSnapshot = serializeDealProducts(products)
    if (currentSnapshot === savedSnapshots[dealId]) {
      return deals.value.find((item) => item.id === dealId) ?? null
    }

    const updated = await updateDealProducts(dealId, products)
    dealProductRows[dealId] = productsToRows(updated.products, catalogProducts.value)
    ensureRowIds(dealProductRows[dealId])
    savedSnapshots[dealId] = serializeDealProducts(updated.products)
    return updated
  }

  function applySavedProducts(dealId: string, products: DealProduct[]) {
    dealProductRows[dealId] = productsToRows(products, catalogProducts.value)
    ensureRowIds(dealProductRows[dealId])
    savedSnapshots[dealId] = serializeDealProducts(products)
  }

  function setDealRows(dealId: string, rows: ProductRow[]) {
    dealProductRows[dealId] = rows
    ensureRowIds(dealProductRows[dealId])
  }

  return {
    getDealRows,
    setDealRows,
    hydrateDealRows,
    resetDealRows,
    saveDealProductRows,
    applySavedProducts,
    createEmptyProductRow,
  }
}
