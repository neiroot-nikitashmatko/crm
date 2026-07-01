import { reactive, watch } from 'vue'
import type { DealProduct } from '@/types/deal'
import type { ProductRow } from '@/types/productRow'
import { useLeads } from '@/composables/useLeads'
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

const leadProductRows = reactive<Record<string, ProductRow[]>>({})
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

export function useLeadProductRows() {
  const { leads, updateLeadProducts } = useLeads()
  const { products: catalogProducts } = useProductsCatalog()

  watch(
    catalogProducts,
    (catalog) => {
      if (!catalog || catalog.length === 0) return

      for (const leadId of Object.keys(leadProductRows)) {
        const rows = leadProductRows[leadId]
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
          // trigger reactive updates for consumers
          leadProductRows[leadId] = [...rows]
        }
      }
    },
    { deep: true },
  )

  function getLeadRows(leadId: string): ProductRow[] {
    if (!leadProductRows[leadId]) {
      const lead = leads.value.find((item) => item.id === leadId)
      const rows = productsToRows(lead?.products ?? [], catalogProducts.value)
      leadProductRows[leadId] = rows.length > 0 ? rows : [createEmptyProductRow()]
      ensureRowIds(leadProductRows[leadId])
      savedSnapshots[leadId] = serializeDealProducts(lead?.products ?? [])
    }
    return leadProductRows[leadId]
  }

  function hydrateLeadRows(leadId: string, force = false) {
    const lead = leads.value.find((item) => item.id === leadId)
    if (!lead) return

    const apiSnapshot = serializeDealProducts(lead.products)
    const hasLocalState = Boolean(leadProductRows[leadId])
    const hasUnsavedChanges =
      hasLocalState && serializeProductRows(leadProductRows[leadId]) !== savedSnapshots[leadId]

    if (force || !hasLocalState || (!hasUnsavedChanges && savedSnapshots[leadId] !== apiSnapshot)) {
      const rows = productsToRows(lead.products, catalogProducts.value)
      leadProductRows[leadId] = rows.length > 0 ? rows : [createEmptyProductRow()]
      ensureRowIds(leadProductRows[leadId])
      savedSnapshots[leadId] = apiSnapshot
    }
  }

  function resetLeadRows(leadId: string) {
    delete leadProductRows[leadId]
    delete savedSnapshots[leadId]
  }

  async function saveLeadProductRows(leadId: string) {
    const rows = leadProductRows[leadId] ?? []
    const products = rowsToDealProducts(rows)
    const currentSnapshot = serializeDealProducts(products)
    if (currentSnapshot === savedSnapshots[leadId]) {
      return leads.value.find((item) => item.id === leadId) ?? null
    }

    const updated = await updateLeadProducts(leadId, products)
    const nextRows = productsToRows(updated.products, catalogProducts.value)
    leadProductRows[leadId] = nextRows.length > 0 ? nextRows : [createEmptyProductRow()]
    ensureRowIds(leadProductRows[leadId])
    savedSnapshots[leadId] = serializeDealProducts(updated.products)
    return updated
  }

  function applySavedProducts(leadId: string, products: DealProduct[]) {
    const rows = productsToRows(products, catalogProducts.value)
    leadProductRows[leadId] = rows.length > 0 ? rows : [createEmptyProductRow()]
    ensureRowIds(leadProductRows[leadId])
    savedSnapshots[leadId] = serializeDealProducts(products)
  }

  function setLeadRows(leadId: string, rows: ProductRow[]) {
    leadProductRows[leadId] = rows
    ensureRowIds(leadProductRows[leadId])
  }

  return {
    getLeadRows,
    setLeadRows,
    hydrateLeadRows,
    resetLeadRows,
    saveLeadProductRows,
    applySavedProducts,
    createEmptyProductRow,
  }
}
