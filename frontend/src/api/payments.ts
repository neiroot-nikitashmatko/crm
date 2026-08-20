import { ApiError, requestJson } from '@/api/httpClient'
import { PAYMENT_PAYER_OPTIONS, type PaymentPayerId } from '@/constants/payments'
import type { Payment, PaymentInput } from '@/types/payment'

interface PaymentsListResponse {
  items: Array<Record<string, unknown>>
}

interface PaymentItemResponse {
  item: Record<string, unknown>
}

const PAYER_IDS = new Set(PAYMENT_PAYER_OPTIONS.map((option) => option.value))

export class PaymentsApiError extends ApiError {
  constructor(message: string, status: number) {
    super(message, status)
    this.name = 'PaymentsApiError'
  }
}

function asOptionalMillis(value: unknown): number | null {
  if (value === null || value === undefined || value === '') return null
  const numeric = Number(value)
  return Number.isFinite(numeric) && numeric > 0 ? numeric : null
}

function asPayerId(value: unknown): PaymentPayerId | null {
  if (typeof value !== 'string') return null
  const trimmed = value.trim()
  return PAYER_IDS.has(trimmed as PaymentPayerId) ? (trimmed as PaymentPayerId) : null
}

function normalizePayment(raw: Record<string, unknown>): Payment {
  return {
    id: String(raw.id ?? ''),
    date: Number(raw.date ?? 0),
    remindAt: asOptionalMillis(raw.remindAt),
    payerId: asPayerId(raw.payerId),
    counterparty: String(raw.counterparty ?? ''),
    amount: Number(raw.amount ?? 0),
    shortTitle: String(raw.shortTitle ?? ''),
    comment: String(raw.comment ?? ''),
    isClosed: Boolean(raw.isClosed),
    createdAt: Number(raw.createdAt ?? 0),
  }
}

async function paymentsRequestJson<T>(path: string, init?: RequestInit): Promise<T> {
  try {
    return await requestJson<T>(path, init)
  } catch (error) {
    if (error instanceof ApiError) {
      throw new PaymentsApiError(error.message, error.status)
    }
    throw error
  }
}

function paymentPayload(input: PaymentInput) {
  return {
    date: input.date,
    remindAt: input.remindAt,
    payerId: input.payerId,
    counterparty: input.counterparty,
    amount: input.amount,
    shortTitle: input.shortTitle,
    comment: input.comment,
  }
}

export async function fetchPayments(): Promise<Payment[]> {
  const payload = await paymentsRequestJson<PaymentsListResponse>('/api/v1/payments', {
    method: 'GET',
  })
  return (payload.items ?? []).map(normalizePayment)
}

export async function createPayment(input: PaymentInput): Promise<Payment> {
  const response = await paymentsRequestJson<PaymentItemResponse>('/api/v1/payments', {
    method: 'POST',
    body: JSON.stringify(paymentPayload(input)),
  })
  return normalizePayment(response.item)
}

export async function updatePayment(paymentId: string, input: PaymentInput): Promise<Payment> {
  const response = await paymentsRequestJson<PaymentItemResponse>(`/api/v1/payments/${paymentId}`, {
    method: 'PATCH',
    body: JSON.stringify(paymentPayload(input)),
  })
  return normalizePayment(response.item)
}

export async function setPaymentClosedRequest(paymentId: string, isClosed: boolean): Promise<Payment> {
  const response = await paymentsRequestJson<PaymentItemResponse>(`/api/v1/payments/${paymentId}`, {
    method: 'PATCH',
    body: JSON.stringify({ isClosed }),
  })
  return normalizePayment(response.item)
}

export async function deletePaymentRequest(paymentId: string): Promise<void> {
  await paymentsRequestJson<{ ok: boolean }>(`/api/v1/payments/${paymentId}`, {
    method: 'DELETE',
  })
}
