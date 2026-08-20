import { computed, ref } from 'vue'
import type { IncomingInvoice, IncomingInvoiceInput } from '@/types/incomingInvoice'

const invoices = ref<IncomingInvoice[]>([])

function createId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `incoming-invoice-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

function nextInvoiceNumber() {
  const maxNumber = invoices.value.reduce((max, invoice) => Math.max(max, invoice.invoiceNumber), 0)
  return maxNumber + 1
}

export function useIncomingInvoices() {
  const sortedInvoices = computed(() =>
    [...invoices.value].sort((left, right) => right.date - left.date || right.createdAt - left.createdAt),
  )

  function addInvoice(input: IncomingInvoiceInput): IncomingInvoice {
    const invoice: IncomingInvoice = {
      id: createId(),
      invoiceNumber: nextInvoiceNumber(),
      date: input.date,
      supplierId: input.supplierId,
      items: input.items.map((item) => ({ ...item })),
      total: input.total,
      comment: input.comment.trim(),
      createdAt: Date.now(),
    }
    invoices.value = [...invoices.value, invoice]
    return invoice
  }

  function updateInvoice(id: string, input: IncomingInvoiceInput): IncomingInvoice | null {
    const index = invoices.value.findIndex((item) => item.id === id)
    if (index < 0) return null

    const current = invoices.value[index]
    const updated: IncomingInvoice = {
      ...current,
      date: input.date,
      supplierId: input.supplierId,
      items: input.items.map((item) => ({ ...item })),
      total: input.total,
      comment: input.comment.trim(),
    }
    const next = [...invoices.value]
    next[index] = updated
    invoices.value = next
    return updated
  }

  function removeInvoice(id: string) {
    invoices.value = invoices.value.filter((item) => item.id !== id)
  }

  return {
    invoices: sortedInvoices,
    addInvoice,
    updateInvoice,
    removeInvoice,
  }
}
