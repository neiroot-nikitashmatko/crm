import type { PickupDelivery } from '@/types/deal'
import type { Lead, NewLeadForm } from '@/types/lead'
import { ApiError, requestJson } from '@/api/httpClient'

interface LeadsListResponse {
  items: Lead[]
}

interface LeadItemResponse {
  item: Lead
}

export class LeadsApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'LeadsApiError'
  }
}

async function leadsRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new LeadsApiError(error.message, error.status)
    }
    throw error
  }
}

export async function fetchLeads(): Promise<Lead[]> {
  const payload = await leadsRequestJson<LeadsListResponse>('/api/v1/leads', { method: 'GET' })
  return payload.items
}

export async function createLead(payload: NewLeadForm, columnId: string, createdBy: string): Promise<Lead> {
  const response = await leadsRequestJson<LeadItemResponse>('/api/v1/leads', {
    method: 'POST',
    body: JSON.stringify({
      ...payload,
      columnId,
      createdBy,
    }),
  })
  return response.item
}

export async function updateLeadColumn(
  leadId: string,
  columnId: string,
  failureReason?: string,
): Promise<Lead> {
  const body: { columnId: string; failureReason?: string } = { columnId }
  if (failureReason !== undefined) {
    body.failureReason = failureReason
  }
  const response = await leadsRequestJson<LeadItemResponse>(`/api/v1/leads/${leadId}/status`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
  return response.item
}

export async function updateLeadComment(leadId: string, leadComments: string): Promise<Lead> {
  const response = await leadsRequestJson<LeadItemResponse>(`/api/v1/leads/${leadId}/comment`, {
    method: 'PATCH',
    body: JSON.stringify({ leadComments }),
  })
  return response.item
}

export async function updateLeadProfile(
  leadId: string,
  firstName: string,
  patronymic: string,
  phone: string,
): Promise<Lead> {
  const response = await leadsRequestJson<LeadItemResponse>(`/api/v1/leads/${leadId}/profile`, {
    method: 'PATCH',
    body: JSON.stringify({ firstName, patronymic, phone }),
  })
  return response.item
}

export async function updateLeadPickupDelivery(
  leadId: string,
  pickupDelivery: PickupDelivery,
): Promise<Lead> {
  const response = await leadsRequestJson<LeadItemResponse>(`/api/v1/leads/${leadId}/pickup-delivery`, {
    method: 'PATCH',
    body: JSON.stringify({ pickupDelivery }),
  })
  return response.item
}

export async function updateLeadProducts(leadId: string, products: Lead['products']): Promise<Lead> {
  const response = await leadsRequestJson<LeadItemResponse>(`/api/v1/leads/${leadId}/products`, {
    method: 'PATCH',
    body: JSON.stringify({ products }),
  })
  return response.item
}

export async function updateLeadProduction(
  leadId: string,
  production: Lead['production'],
): Promise<Lead> {
  const response = await leadsRequestJson<LeadItemResponse>(`/api/v1/leads/${leadId}/production`, {
    method: 'PATCH',
    body: JSON.stringify({ production }),
  })
  return response.item
}

export async function deleteLead(leadId: string): Promise<void> {
  await leadsRequestJson<{ ok: boolean }>(`/api/v1/leads/${leadId}`, {
    method: 'DELETE',
  })
}
