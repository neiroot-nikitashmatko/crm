import type { DealProduct } from '@/types/deal'
import type { ProductRow } from '@/types/productRow'
import type { CatalogProduct } from '@/types/productCatalog'

let productRowSeq = 0

export function createEmptyProductRow(): ProductRow {
  productRowSeq += 1
  return {
    rowId: `product-row-${productRowSeq}`,
    catalogProductId: undefined,
    title: '',
    quantity: 1,
    unitPrice: 0,
  }
}

export function calculateProductsTotal(products: DealProduct[]): number {
  return products.reduce((sum, product) => {
    if (!product.title.trim()) return sum

    const quantity = Number(product.quantity)
    const unitPrice = normalizeUnitPrice(product.unitPrice)
    const safeQuantity = Number.isFinite(quantity) && quantity > 0 ? quantity : 0

    return sum + safeQuantity * unitPrice
  }, 0)
}

export function normalizeUnitPrice(value: unknown): number {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed < 0) return 0
  return Math.round(parsed * 100) / 100
}

export function productsToRows(products: DealProduct[], catalog: CatalogProduct[] = []): ProductRow[] {
  return products.map((product) => {
    productRowSeq += 1
    return {
      rowId: `product-row-${productRowSeq}`,
      catalogProductId: resolveCatalogProductId(product, catalog),
      title: product.title,
      quantity: product.quantity,
      unitPrice: product.unitPrice,
    }
  })
}

export function rowsToDealProducts(rows: ProductRow[]): DealProduct[] {
  return rows
    .map((row) => ({
      catalogProductId: row.catalogProductId,
      title: row.title.trim(),
      quantity: Number.isFinite(Number(row.quantity)) && Number(row.quantity) > 0 ? Number(row.quantity) : 1,
      unitPrice: normalizeUnitPrice(row.unitPrice),
    }))
    .filter((product) => product.title.length > 0)
}

export function resolveCatalogProductId(product: DealProduct, catalog: CatalogProduct[]): string | undefined {
  if (product.catalogProductId) return product.catalogProductId

  const title = product.title.trim()
  if (!title) return undefined

  const exactMatch = catalog.find(
    (item) => item.name === title && item.cost === normalizeUnitPrice(product.unitPrice),
  )
  if (exactMatch) return exactMatch.id

  const titleMatch = catalog.find((item) => item.name === title)
  return titleMatch?.id
}

export function ensureRowIds(rows: ProductRow[]): ProductRow[] {
  for (const row of rows) {
    if (!row.rowId) {
      productRowSeq += 1
      row.rowId = `product-row-${productRowSeq}`
    }
  }
  return rows
}

export function serializeDealProducts(products: DealProduct[]): string {
  return JSON.stringify(
    products.map((product) => ({
      title: product.title,
      quantity: product.quantity,
      unitPrice: normalizeUnitPrice(product.unitPrice),
    })),
  )
}

export function serializeProductRows(rows: ProductRow[]): string {
  return serializeDealProducts(rowsToDealProducts(rows))
}
