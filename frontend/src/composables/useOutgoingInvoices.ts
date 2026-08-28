import { computed, ref } from 'vue'
import {
  createOutgoingInvoice,
  deleteOutgoingInvoice,
  fetchOutgoingInvoices,
  updateOutgoingInvoice,
} from '@/api/outgoingInvoices'
import type { OutgoingInvoice, OutgoingInvoiceInput } from '@/types/outgoingInvoice'

const invoices = ref<OutgoingInvoice[]>([])
const isLoaded = ref(false)
const isLoading = ref(false)
let inFlight: Promise<void> | null = null

function findInvoiceByDealId(dealId: string, excludeInvoiceId?: string | null): OutgoingInvoice | null {
  return (
    invoices.value.find(
      (item) => item.dealId === dealId && (!excludeInvoiceId || item.id !== excludeInvoiceId),
    ) ?? null
  )
}

export function useOutgoingInvoices() {
  const sortedInvoices = computed(() =>
    [...invoices.value].sort((left, right) => right.date - left.date || right.createdAt - left.createdAt),
  )

  async function loadInvoices(force = false) {
    if (isLoaded.value && !force) return
    if (inFlight) return inFlight

    inFlight = (async () => {
      isLoading.value = true
      try {
        invoices.value = await fetchOutgoingInvoices()
        isLoaded.value = true
      } finally {
        isLoading.value = false
        inFlight = null
      }
    })()

    return inFlight
  }

  function getInvoiceByDealId(dealId: string): OutgoingInvoice | null {
    return findInvoiceByDealId(dealId)
  }

  function hasInvoiceForDeal(dealId: string): boolean {
    return findInvoiceByDealId(dealId) !== null
  }

  async function addInvoice(input: OutgoingInvoiceInput): Promise<OutgoingInvoice> {
    const created = await createOutgoingInvoice({
      ...input,
      comment: input.comment.trim(),
    })
    invoices.value = [...invoices.value, created]
    return created
  }

  async function updateInvoice(id: string, input: OutgoingInvoiceInput): Promise<OutgoingInvoice> {
    const updated = await updateOutgoingInvoice(id, {
      ...input,
      comment: input.comment.trim(),
    })
    invoices.value = invoices.value.map((item) => (item.id === id ? updated : item))
    return updated
  }

  async function removeInvoice(id: string) {
    await deleteOutgoingInvoice(id)
    invoices.value = invoices.value.filter((item) => item.id !== id)
  }

  return {
    invoices: sortedInvoices,
    isLoading,
    loadInvoices,
    getInvoiceByDealId,
    hasInvoiceForDeal,
    addInvoice,
    updateInvoice,
    removeInvoice,
  }
}
