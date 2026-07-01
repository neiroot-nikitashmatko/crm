import type { StoredAttachment, StoredActivity } from '@/types/attachment'

export type DealKanbanColumnId = 'production' | 'pickup' | 'delivery' | 'closed' | 'failed'

export type DealActivityType = 'system' | 'comment'

export interface DealProduct {
  catalogProductId?: string
  title: string
  quantity: number
  unitPrice: number
}

export interface PickupDelivery {
  pickupAddress: string
  pickupDate: number | null
  deliveryAddress: string
  deliveryDate: number | null
  courier: string
}

export type DealAttachment = StoredAttachment

export type DealActivityEntry = StoredActivity

export interface Deal {
  id: string
  leadId?: string
  dealNumber: number
  firstName: string
  patronymic: string
  phone: string
  trafficSource: string
  totalAmount: number
  dealComments: string
  failureReason: string
  createdAt: number
  createdBy: string
  products: DealProduct[]
  production: {
    nomenclature: string
    dueAt: number | null
    employee: string
  }
  productionDueAt: number | null
  pickupDelivery: PickupDelivery
  attachments: DealAttachment[]
  activities: DealActivityEntry[]
  columnId: DealKanbanColumnId
  status?: string
}
