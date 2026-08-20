import { computed, ref } from 'vue'
import type { Supplier, SupplierInput } from '@/types/supplier'

const MOCK_SUPPLIERS: Supplier[] = [
  {
    id: 'mock-supplier-1',
    name: 'ООО «МеталлПром»',
    contactPerson: 'Иванов Сергей',
    phone: '+79031234567',
    inn: '7701234567',
    kpp: '770101001',
    ogrn: '1027700132195',
    legalAddress: 'г. Москва, ул. Лесная, д. 5',
    actualAddress: 'г. Москва, ул. Лесная, д. 5, склад 2',
    bik: '044525225',
    settlementAccount: '40702810900000012345',
    correspondentAccount: '30101810400000000225',
    createdAt: 1,
  },
  {
    id: 'mock-supplier-2',
    name: 'АО «СеверСнаб»',
    contactPerson: 'Петрова Анна',
    phone: '+79161112233',
    inn: '7812345678',
    kpp: '781201001',
    ogrn: '1027800000001',
    legalAddress: 'г. Санкт-Петербург, Невский пр., д. 28',
    actualAddress: 'г. Санкт-Петербург, ул. Салова, д. 61',
    bik: '044030653',
    settlementAccount: '40702810100000067890',
    correspondentAccount: '30101810500000000653',
    createdAt: 2,
  },
  {
    id: 'mock-supplier-3',
    name: 'ИП Козлов Дмитрий',
    contactPerson: 'Козлов Дмитрий',
    phone: '+79265554433',
    inn: '503012345678',
    kpp: '',
    ogrn: '304503012345671',
    legalAddress: 'Московская обл., г. Подольск, ул. Кирова, д. 12',
    actualAddress: 'Московская обл., г. Подольск, ул. Кирова, д. 12',
    bik: '044525974',
    settlementAccount: '40802810000000011223',
    correspondentAccount: '30101810100000000974',
    createdAt: 3,
  },
  {
    id: 'mock-supplier-4',
    name: 'ООО «СтройРесурс»',
    contactPerson: 'Сидорова Елена',
    phone: '+79876543210',
    inn: '5409876543',
    kpp: '540901001',
    ogrn: '1025400001122',
    legalAddress: 'г. Новосибирск, ул. Красный проспект, д. 17',
    actualAddress: 'г. Новосибирск, ул. Станционная, д. 40а',
    bik: '045004774',
    settlementAccount: '40702810700000044556',
    correspondentAccount: '30101810500000000774',
    createdAt: 4,
  },
]

const suppliers = ref<Supplier[]>([...MOCK_SUPPLIERS])

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

function createId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `supplier-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

export function useSuppliers() {
  const sortedSuppliers = computed(() =>
    [...suppliers.value].sort((left, right) => left.name.localeCompare(right.name, 'ru')),
  )

  function addSupplier(input: SupplierInput): Supplier {
    const normalized = normalizeInput(input)
    const supplier: Supplier = {
      id: createId(),
      ...normalized,
      createdAt: Date.now(),
    }
    suppliers.value = [...suppliers.value, supplier]
    return supplier
  }

  function updateSupplier(id: string, input: SupplierInput): Supplier | null {
    const index = suppliers.value.findIndex((item) => item.id === id)
    if (index < 0) return null

    const current = suppliers.value[index]
    const updated: Supplier = {
      ...current,
      ...normalizeInput(input),
    }
    const next = [...suppliers.value]
    next[index] = updated
    suppliers.value = next
    return updated
  }

  function removeSupplier(id: string) {
    suppliers.value = suppliers.value.filter((item) => item.id !== id)
  }

  function getSupplierById(id: string) {
    return suppliers.value.find((item) => item.id === id) ?? null
  }

  return {
    suppliers: sortedSuppliers,
    addSupplier,
    updateSupplier,
    removeSupplier,
    getSupplierById,
  }
}
