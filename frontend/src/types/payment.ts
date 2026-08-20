import type { PaymentPayerId } from '@/constants/payments'

export interface Payment {
  id: string
  date: number
  remindAt: number | null
  payerId: PaymentPayerId | null
  counterparty: string
  amount: number
  shortTitle: string
  comment: string
  isClosed: boolean
  createdAt: number
}

export type PaymentInput = Omit<Payment, 'id' | 'createdAt' | 'isClosed'>
