export interface StockMovementItem {
  catalogProductId?: string
  title: string
  quantity: number
}

export function stockMovementKey(item: Pick<StockMovementItem, 'catalogProductId' | 'title'>) {
  if (item.catalogProductId) return `id:${item.catalogProductId}`
  return `title:${item.title.trim().toLowerCase()}`
}

export function safeStockQuantity(value: number) {
  const quantity = Number(value)
  return Number.isFinite(quantity) ? quantity : 0
}

export interface OutgoingStockValidationIssue {
  key: string
  title: string
  requested: number
  available: number
}

export function validateOutgoingInvoiceStock(
  items: StockMovementItem[],
  quantities: Map<string, number>,
  resolveTitle?: (item: StockMovementItem) => string,
): OutgoingStockValidationIssue[] {
  const issues: OutgoingStockValidationIssue[] = []
  const requestedByKey = new Map<string, { title: string; quantity: number }>()

  for (const item of items) {
    const title = (resolveTitle?.(item) ?? item.title).trim()
    if (!title && !item.catalogProductId) continue

    const key = stockMovementKey(item)
    const current = requestedByKey.get(key) ?? { title: title || 'Товар', quantity: 0 }
    current.quantity += safeStockQuantity(item.quantity)
    if (title) current.title = title
    requestedByKey.set(key, current)
  }

  for (const [key, requested] of requestedByKey) {
    const available = quantities.get(key) ?? 0
    if (requested.quantity > available) {
      issues.push({
        key,
        title: requested.title,
        requested: requested.quantity,
        available,
      })
    }
  }

  return issues
}
