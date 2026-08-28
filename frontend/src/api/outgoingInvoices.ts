import type { OutgoingInvoice, OutgoingInvoiceInput, OutgoingInvoiceItem } from '@/types/outgoingInvoice'
import { ApiError, requestJson } from '@/api/httpClient'

interface OutgoingInvoicesListResponse {
  items: OutgoingInvoice[]
}

interface OutgoingInvoiceItemResponse {
  item: OutgoingInvoice
}

export class OutgoingInvoicesApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'OutgoingInvoicesApiError'
  }
}

function normalizeInvoiceItem(item: OutgoingInvoiceItem): OutgoingInvoiceItem {
  const catalogProductId = String(item.catalogProductId ?? '').trim()
  return {
    ...(catalogProductId ? { catalogProductId } : {}),
    title: String(item.title ?? ''),
    quantity: Number(item.quantity ?? 0),
    unitPrice: Number(item.unitPrice ?? 0),
  }
}

function normalizeOutgoingInvoice(raw: OutgoingInvoice): OutgoingInvoice {
  return {
    id: String(raw.id ?? ''),
    invoiceNumber: Number(raw.invoiceNumber ?? 0),
    date: Number(raw.date ?? 0),
    dealId: String(raw.dealId ?? ''),
    items: Array.isArray(raw.items) ? raw.items.map(normalizeInvoiceItem) : [],
    total: Number(raw.total ?? 0),
    comment: String(raw.comment ?? ''),
    createdAt: Number(raw.createdAt ?? 0),
  }
}

async function outgoingInvoicesRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new OutgoingInvoicesApiError(error.message, error.status)
    }
    throw error
  }
}

export async function fetchOutgoingInvoices(): Promise<OutgoingInvoice[]> {
  const payload = await outgoingInvoicesRequestJson<OutgoingInvoicesListResponse>(
    '/api/v1/outgoing-invoices',
    { method: 'GET' },
  )
  if (!Array.isArray(payload.items)) return []
  return payload.items.map(normalizeOutgoingInvoice)
}

export async function createOutgoingInvoice(payload: OutgoingInvoiceInput): Promise<OutgoingInvoice> {
  const response = await outgoingInvoicesRequestJson<OutgoingInvoiceItemResponse>(
    '/api/v1/outgoing-invoices',
    {
      method: 'POST',
      body: JSON.stringify(payload),
    },
  )
  return normalizeOutgoingInvoice(response.item)
}

export async function updateOutgoingInvoice(
  invoiceId: string,
  payload: OutgoingInvoiceInput,
): Promise<OutgoingInvoice> {
  const response = await outgoingInvoicesRequestJson<OutgoingInvoiceItemResponse>(
    `/api/v1/outgoing-invoices/${invoiceId}`,
    {
      method: 'PATCH',
      body: JSON.stringify(payload),
    },
  )
  return normalizeOutgoingInvoice(response.item)
}

export async function deleteOutgoingInvoice(invoiceId: string): Promise<void> {
  await outgoingInvoicesRequestJson<{ ok: boolean }>(`/api/v1/outgoing-invoices/${invoiceId}`, {
    method: 'DELETE',
  })
}
