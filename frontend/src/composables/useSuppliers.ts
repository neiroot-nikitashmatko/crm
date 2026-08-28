import { computed, ref } from 'vue'
import {
  createSupplier,
  deleteSupplier,
  fetchSuppliers,
  updateSupplier,
} from '@/api/suppliers'
import type { Supplier, SupplierInput } from '@/types/supplier'

const suppliers = ref<Supplier[]>([])
const isLoaded = ref(false)
const isLoading = ref(false)
let inFlight: Promise<void> | null = null

function normalizeInput(input: SupplierInput): SupplierInput {
  return {
    name: input.name.trim(),
    contactPerson: input.contactPerson.trim(),
    phone: input.phone.trim(),
    inn: input.inn.trim(),
    kpp: input.kpp.trim(),
    ogrn: input.ogrn.trim(),
    legalAddress: input.legalAddress.trim(),
    actualAddress: input.actualAddress.trim(),
    bik: input.bik.trim(),
    settlementAccount: input.settlementAccount.trim(),
    correspondentAccount: input.correspondentAccount.trim(),
  }
}

export function useSuppliers() {
  const sortedSuppliers = computed(() =>
    [...suppliers.value].sort((left, right) => left.name.localeCompare(right.name, 'ru')),
  )

  async function loadSuppliers(force = false) {
    if (isLoaded.value && !force) return
    if (inFlight) return inFlight

    inFlight = (async () => {
      isLoading.value = true
      try {
        suppliers.value = await fetchSuppliers()
        isLoaded.value = true
      } finally {
        isLoading.value = false
        inFlight = null
      }
    })()

    return inFlight
  }

  async function addSupplier(input: SupplierInput): Promise<Supplier> {
    const created = await createSupplier(normalizeInput(input))
    suppliers.value = [...suppliers.value, created]
    return created
  }

  async function updateSupplierById(id: string, input: SupplierInput): Promise<Supplier> {
    const updated = await updateSupplier(id, normalizeInput(input))
    suppliers.value = suppliers.value.map((item) => (item.id === id ? updated : item))
    return updated
  }

  async function removeSupplier(id: string) {
    await deleteSupplier(id)
    suppliers.value = suppliers.value.filter((item) => item.id !== id)
  }

  function getSupplierById(id: string) {
    return suppliers.value.find((item) => item.id === id) ?? null
  }

  return {
    suppliers: sortedSuppliers,
    isLoading,
    loadSuppliers,
    addSupplier,
    updateSupplier: updateSupplierById,
    removeSupplier,
    getSupplierById,
  }
}
