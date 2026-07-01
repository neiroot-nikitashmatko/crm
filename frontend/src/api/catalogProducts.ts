import type { CatalogProduct, NewCatalogProductInput } from '@/types/productCatalog'
import { ApiError, requestJson } from '@/api/httpClient'

interface CatalogProductsListResponse {
  items: CatalogProduct[]
}

interface CatalogProductItemResponse {
  item: CatalogProduct
}

export class CatalogProductsApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'CatalogProductsApiError'
  }
}

function normalizeCatalogProduct(raw: CatalogProduct): CatalogProduct {
  return {
    id: String(raw.id),
    name: String(raw.name ?? ''),
    sku: String(raw.sku ?? ''),
    category: String(raw.category ?? ''),
    cost: Number(raw.cost ?? 0),
    createdAt: Number(raw.createdAt ?? Date.now()),
  }
}

async function catalogRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new CatalogProductsApiError(error.message, error.status)
    }
    throw error
  }
}

export async function fetchCatalogProducts(): Promise<CatalogProduct[]> {
  const payload = await catalogRequestJson<CatalogProductsListResponse>('/api/v1/catalog-products', {
    method: 'GET',
  })
  return payload.items.map(normalizeCatalogProduct)
}

export async function createCatalogProduct(payload: NewCatalogProductInput): Promise<CatalogProduct> {
  const response = await catalogRequestJson<CatalogProductItemResponse>('/api/v1/catalog-products', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return normalizeCatalogProduct(response.item)
}

export async function updateCatalogProduct(
  productId: string,
  payload: NewCatalogProductInput,
): Promise<CatalogProduct> {
  const response = await catalogRequestJson<CatalogProductItemResponse>(
    `/api/v1/catalog-products/${productId}`,
    {
      method: 'PATCH',
      body: JSON.stringify(payload),
    },
  )
  return normalizeCatalogProduct(response.item)
}

export async function deleteCatalogProduct(productId: string): Promise<void> {
  await catalogRequestJson<{ ok: boolean }>(`/api/v1/catalog-products/${productId}`, {
    method: 'DELETE',
  })
}
