import { computed, ref } from 'vue'
import {
  createPayment,
  deletePaymentRequest,
  fetchPayments,
  setPaymentClosedRequest,
  updatePayment,
} from '@/api/payments'
import { PAYMENT_SHORT_TITLE_MAX_LENGTH } from '@/constants/payments'
import type { Payment, PaymentInput } from '@/types/payment'

const payments = ref<Payment[]>([])
const isLoaded = ref(false)
const isLoading = ref(false)

function sortPayments(items: Payment[]) {
  return [...items].sort((left, right) => left.date - right.date || left.createdAt - right.createdAt)
}

function normalizeInput(input: PaymentInput): PaymentInput {
  return {
    date: input.date,
    remindAt: input.remindAt,
    payerId: input.payerId,
    counterparty: input.counterparty.trim(),
    amount: Number.isFinite(input.amount) ? Math.max(0, input.amount) : 0,
    shortTitle: input.shortTitle.trim().slice(0, PAYMENT_SHORT_TITLE_MAX_LENGTH),
    comment: input.comment.trim(),
  }
}

export function usePayments() {
  const sortedPayments = computed(() => sortPayments(payments.value))

  async function loadPayments(force = false) {
    if (isLoading.value) return
    if (isLoaded.value && !force) return

    isLoading.value = true
    try {
      payments.value = await fetchPayments()
      isLoaded.value = true
    } finally {
      isLoading.value = false
    }
  }

  async function addPayment(input: PaymentInput): Promise<Payment> {
    const created = await createPayment(normalizeInput(input))
    payments.value = sortPayments([...payments.value.filter((item) => item.id !== created.id), created])
    isLoaded.value = true
    return created
  }

  async function savePayment(paymentId: string, input: PaymentInput): Promise<Payment> {
    const updated = await updatePayment(paymentId, normalizeInput(input))
    payments.value = sortPayments(
      payments.value.map((payment) => (payment.id === updated.id ? updated : payment)),
    )
    return updated
  }

  async function deletePayment(paymentId: string): Promise<void> {
    await deletePaymentRequest(paymentId)
    payments.value = payments.value.filter((payment) => payment.id !== paymentId)
  }

  async function setPaymentClosed(paymentId: string, isClosed: boolean): Promise<Payment> {
    const updated = await setPaymentClosedRequest(paymentId, isClosed)
    payments.value = payments.value.map((payment) => (payment.id === updated.id ? updated : payment))
    return updated
  }

  return {
    payments: sortedPayments,
    isLoaded,
    isLoading,
    loadPayments,
    addPayment,
    savePayment,
    deletePayment,
    setPaymentClosed,
  }
}
