import { computed, ref } from 'vue'
import type { OutgoingInvoice, OutgoingInvoiceInput } from '@/types/outgoingInvoice'

const invoices = ref<OutgoingInvoice[]>([])

function createId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `outgoing-invoice-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

function nextInvoiceNumber() {
  const maxNumber = invoices.value.reduce((max, invoice) => Math.max(max, invoice.invoiceNumber), 0)
  return maxNumber + 1
}

export function useOutgoingInvoices() {
  const sortedInvoices = computed(() =>
    [...invoices.value].sort((left, right) => right.date - left.date || right.createdAt - left.createdAt),
  )

  function addInvoice(input: OutgoingInvoiceInput): OutgoingInvoice {
    const invoice: OutgoingInvoice = {
      id: createId(),
      invoiceNumber: nextInvoiceNumber(),
      date: input.date,
      dealId: input.dealId,
      items: input.items.map((item) => ({ ...item })),
      total: input.total,
      comment: input.comment.trim(),
      createdAt: Date.now(),
    }
    invoices.value = [...invoices.value, invoice]
    return invoice
  }

  function updateInvoice(id: string, input: OutgoingInvoiceInput): OutgoingInvoice | null {
    const index = invoices.value.findIndex((item) => item.id === id)
    if (index < 0) return null

    const current = invoices.value[index]
    const updated: OutgoingInvoice = {
      ...current,
      date: input.date,
      dealId: input.dealId,
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
