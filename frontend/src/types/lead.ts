import type { DealProduct, PickupDelivery } from '@/types/deal'
import type { StoredAttachment, StoredActivity } from '@/types/attachment'

export type LeadAttachment = StoredAttachment
export type LeadActivityEntry = StoredActivity

export interface LeadProduction {
  nomenclature: string
  dueAt: number | null
  employee: string
}

export interface NewLeadForm {
  firstName: string
  patronymic: string
  phone: string
  trafficSource: string
}

export interface Lead extends NewLeadForm {
  id: string
  leadNumber: number
  columnId: string
  leadComments: string
  failureReason: string
  createdBy: string
  createdAt: number
  updatedAt: number
  pickupDelivery: PickupDelivery
  products: DealProduct[]
  production: LeadProduction
  attachments: LeadAttachment[]
  activities: LeadActivityEntry[]
}
