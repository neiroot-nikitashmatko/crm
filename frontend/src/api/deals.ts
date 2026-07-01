import type { Deal, DealKanbanColumnId, DealProduct, PickupDelivery } from '@/types/deal'
import { mapColumnIdToDealStatus } from '@/utils/dealKanban'
import { ApiError, requestJson } from '@/api/httpClient'

interface DealsListResponse {
  items: Deal[]
}

interface DealItemResponse {
  item: Deal
}

export class DealsApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'DealsApiError'
  }
}

async function dealsRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new DealsApiError(error.message, error.status)
    }
    throw error
  }
}

function extractDealList(payload: any): Deal[] {
  if (Array.isArray(payload)) return payload as Deal[]
  if (Array.isArray(payload?.items)) return payload.items as Deal[]
  if (Array.isArray(payload?.deals)) return payload.deals as Deal[]
  if (Array.isArray(payload?.data)) return payload.data as Deal[]
  return []
}

export async function fetchDeals(): Promise<Deal[]> {
  const payload = await dealsRequestJson<DealsListResponse | Deal[]>('/api/v1/deals', { method: 'GET' })
  return extractDealList(payload)
}

export async function createDealFromLead(payload: {
  leadId: string
  createdBy: string
  products: DealProduct[]
  production: { nomenclature: string; dueAt: number | null; employee: string }
  pickupDelivery?: PickupDelivery
}): Promise<Deal> {
  const response = await dealsRequestJson<DealItemResponse>('/api/v1/deals/from-lead', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return response.item
}

export async function updateDealComment(dealId: string, dealComments: string): Promise<Deal> {
  const response = await dealsRequestJson<DealItemResponse>(`/api/v1/deals/${dealId}/comment`, {
    method: 'PATCH',
    body: JSON.stringify({ dealComments }),
  })
  return response.item
}

export async function updateDealStatus(
  dealId: string,
  columnId: DealKanbanColumnId,
  failureReason?: string,
): Promise<Deal> {
  const body: { columnId: string; failureReason?: string } = {
    columnId: mapColumnIdToDealStatus(columnId),
  }
  if (failureReason !== undefined) {
    body.failureReason = failureReason
  }
  const response = await dealsRequestJson<DealItemResponse>(`/api/v1/deals/${dealId}/status`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
  return response.item
}

export async function updateDealProductionDueAt(dealId: string, dueAt: number | null): Promise<Deal> {
  const response = await dealsRequestJson<DealItemResponse>(`/api/v1/deals/${dealId}/production-due-at`, {
    method: 'PATCH',
    body: JSON.stringify({ dueAt }),
  })
  return response.item
}

export async function updateDealProduction(
  dealId: string,
  production: Deal['production'],
): Promise<Deal> {
  const response = await dealsRequestJson<DealItemResponse>(`/api/v1/deals/${dealId}/production`, {
    method: 'PATCH',
    body: JSON.stringify({ production }),
  })
  return response.item
}

export async function updateDealPickupDelivery(dealId: string, pickupDelivery: PickupDelivery): Promise<Deal> {
  const response = await dealsRequestJson<DealItemResponse>(`/api/v1/deals/${dealId}/pickup-delivery`, {
    method: 'PATCH',
    body: JSON.stringify({ pickupDelivery }),
  })
  return response.item
}

export async function updateDealProducts(dealId: string, products: DealProduct[]): Promise<Deal> {
  const response = await dealsRequestJson<DealItemResponse>(`/api/v1/deals/${dealId}/products`, {
    method: 'PATCH',
    body: JSON.stringify({ products }),
  })
  return response.item
}

export async function deleteDeal(dealId: string): Promise<void> {
  await dealsRequestJson<{ ok: boolean }>(`/api/v1/deals/${dealId}`, { method: 'DELETE' })
}
