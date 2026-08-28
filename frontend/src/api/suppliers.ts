import type { Supplier, SupplierInput } from '@/types/supplier'
import { ApiError, requestJson } from '@/api/httpClient'

interface SuppliersListResponse {
  items: Supplier[]
}

interface SupplierItemResponse {
  item: Supplier
}

export class SuppliersApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'SuppliersApiError'
  }
}

function normalizeSupplier(raw: Supplier): Supplier {
  return {
    id: String(raw.id ?? ''),
    name: String(raw.name ?? ''),
    contactPerson: String(raw.contactPerson ?? ''),
    phone: String(raw.phone ?? ''),
    inn: String(raw.inn ?? ''),
    kpp: String(raw.kpp ?? ''),
    ogrn: String(raw.ogrn ?? ''),
    legalAddress: String(raw.legalAddress ?? ''),
    actualAddress: String(raw.actualAddress ?? ''),
    bik: String(raw.bik ?? ''),
    settlementAccount: String(raw.settlementAccount ?? ''),
    correspondentAccount: String(raw.correspondentAccount ?? ''),
    createdAt: Number(raw.createdAt ?? 0),
  }
}

async function suppliersRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new SuppliersApiError(error.message, error.status)
    }
    throw error
  }
}

export async function fetchSuppliers(): Promise<Supplier[]> {
  const payload = await suppliersRequestJson<SuppliersListResponse>('/api/v1/suppliers', {
    method: 'GET',
  })
  if (!Array.isArray(payload.items)) return []
  return payload.items.map(normalizeSupplier)
}

export async function createSupplier(payload: SupplierInput): Promise<Supplier> {
  const response = await suppliersRequestJson<SupplierItemResponse>('/api/v1/suppliers', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return normalizeSupplier(response.item)
}

export async function updateSupplier(supplierId: string, payload: SupplierInput): Promise<Supplier> {
  const response = await suppliersRequestJson<SupplierItemResponse>(`/api/v1/suppliers/${supplierId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
  return normalizeSupplier(response.item)
}

export async function deleteSupplier(supplierId: string): Promise<void> {
  await suppliersRequestJson<{ ok: boolean }>(`/api/v1/suppliers/${supplierId}`, {
    method: 'DELETE',
  })
}
