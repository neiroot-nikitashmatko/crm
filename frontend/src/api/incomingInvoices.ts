import type { IncomingInvoice, IncomingInvoiceInput, IncomingInvoiceItem } from '@/types/incomingInvoice'
import { ApiError, requestJson } from '@/api/httpClient'

interface IncomingInvoicesListResponse {
  items: IncomingInvoice[]
}

interface IncomingInvoiceItemResponse {
  item: IncomingInvoice
}

export class IncomingInvoicesApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'IncomingInvoicesApiError'
  }
}

function normalizeInvoiceItem(item: IncomingInvoiceItem): IncomingInvoiceItem {
  const catalogProductId = String(item.catalogProductId ?? '').trim()
  return {
    ...(catalogProductId ? { catalogProductId } : {}),
    title: String(item.title ?? ''),
    quantity: Number(item.quantity ?? 0),
    unitPrice: Number(item.unitPrice ?? 0),
  }
}

function normalizeIncomingInvoice(raw: IncomingInvoice): IncomingInvoice {
  return {
    id: String(raw.id ?? ''),
    invoiceNumber: Number(raw.invoiceNumber ?? 0),
    date: Number(raw.date ?? 0),
    supplierId: String(raw.supplierId ?? ''),
    items: Array.isArray(raw.items) ? raw.items.map(normalizeInvoiceItem) : [],
    total: Number(raw.total ?? 0),
    comment: String(raw.comment ?? ''),
    createdAt: Number(raw.createdAt ?? 0),
  }
}

async function incomingInvoicesRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new IncomingInvoicesApiError(error.message, error.status)
    }
    throw error
  }
}

export async function fetchIncomingInvoices(): Promise<IncomingInvoice[]> {
  const payload = await incomingInvoicesRequestJson<IncomingInvoicesListResponse>(
    '/api/v1/incoming-invoices',
    { method: 'GET' },
  )
  if (!Array.isArray(payload.items)) return []
  return payload.items.map(normalizeIncomingInvoice)
}

export async function createIncomingInvoice(payload: IncomingInvoiceInput): Promise<IncomingInvoice> {
  const response = await incomingInvoicesRequestJson<IncomingInvoiceItemResponse>(
    '/api/v1/incoming-invoices',
    {
      method: 'POST',
      body: JSON.stringify(payload),
    },
  )
  return normalizeIncomingInvoice(response.item)
}

export async function updateIncomingInvoice(
  invoiceId: string,
  payload: IncomingInvoiceInput,
): Promise<IncomingInvoice> {
  const response = await incomingInvoicesRequestJson<IncomingInvoiceItemResponse>(
    `/api/v1/incoming-invoices/${invoiceId}`,
    {
      method: 'PATCH',
      body: JSON.stringify(payload),
    },
  )
  return normalizeIncomingInvoice(response.item)
}

export async function deleteIncomingInvoice(invoiceId: string): Promise<void> {
  await incomingInvoicesRequestJson<{ ok: boolean }>(`/api/v1/incoming-invoices/${invoiceId}`, {
    method: 'DELETE',
  })
}
