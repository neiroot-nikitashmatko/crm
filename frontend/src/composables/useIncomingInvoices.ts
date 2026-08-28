import { computed, ref } from 'vue'
import {
  createIncomingInvoice,
  deleteIncomingInvoice,
  fetchIncomingInvoices,
  updateIncomingInvoice,
} from '@/api/incomingInvoices'
import type { IncomingInvoice, IncomingInvoiceInput } from '@/types/incomingInvoice'

const invoices = ref<IncomingInvoice[]>([])
const isLoaded = ref(false)
const isLoading = ref(false)
let inFlight: Promise<void> | null = null

export function useIncomingInvoices() {
  const sortedInvoices = computed(() =>
    [...invoices.value].sort((left, right) => right.date - left.date || right.createdAt - left.createdAt),
  )

  async function loadInvoices(force = false) {
    if (isLoaded.value && !force) return
    if (inFlight) return inFlight

    inFlight = (async () => {
      isLoading.value = true
      try {
        invoices.value = await fetchIncomingInvoices()
        isLoaded.value = true
      } finally {
        isLoading.value = false
        inFlight = null
      }
    })()

    return inFlight
  }

  async function addInvoice(input: IncomingInvoiceInput): Promise<IncomingInvoice> {
    const created = await createIncomingInvoice({
      ...input,
      comment: input.comment.trim(),
    })
    invoices.value = [...invoices.value, created]
    return created
  }

  async function updateInvoice(id: string, input: IncomingInvoiceInput): Promise<IncomingInvoice> {
    const updated = await updateIncomingInvoice(id, {
      ...input,
      comment: input.comment.trim(),
    })
    invoices.value = invoices.value.map((item) => (item.id === id ? updated : item))
    return updated
  }

  async function removeInvoice(id: string) {
    await deleteIncomingInvoice(id)
    invoices.value = invoices.value.filter((item) => item.id !== id)
  }

  return {
    invoices: sortedInvoices,
    isLoading,
    loadInvoices,
    addInvoice,
    updateInvoice,
    removeInvoice,
  }
}
