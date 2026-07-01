import type { Deal, DealKanbanColumnId } from '@/types/deal'
import { isDealProductionComplete } from '@/utils/productionCalendar'
import {
  DELIVERY_SECTION_LOCKED_MESSAGE,
  isDeliverySectionLocked,
  isPickupSectionLocked,
  PICKUP_SECTION_LOCKED_MESSAGE,
} from '@/utils/pickupDelivery'

export const WORKFLOW_DEAL_COLUMN_IDS = ['production', 'pickup', 'delivery'] as const

export type WorkflowDealColumnId = (typeof WORKFLOW_DEAL_COLUMN_IDS)[number]

export function isDealPickupComplete(deal: Deal): boolean {
  const { pickupAddress, pickupDate } = deal.pickupDelivery
  return (
    pickupAddress.trim() !== '' &&
    pickupDate !== null &&
    pickupDate !== undefined &&
    !Number.isNaN(pickupDate)
  )
}

export function isDealDeliveryComplete(deal: Deal): boolean {
  const { deliveryAddress, deliveryDate, courier } = deal.pickupDelivery
  return (
    deliveryAddress.trim() !== '' &&
    deliveryDate !== null &&
    deliveryDate !== undefined &&
    !Number.isNaN(deliveryDate) &&
    courier.trim() !== ''
  )
}

export function mapDealStatusToColumnId(status: string | undefined): DealKanbanColumnId {
  switch (String(status ?? 'today').toLowerCase()) {
    case 'closed':
      return 'closed'
    case 'failed':
      return 'failed'
    case 'tomorrow':
      return 'pickup'
    case 'later':
      return 'delivery'
    case 'today':
    default:
      return 'production'
  }
}

export function mapColumnIdToDealStatus(columnId: DealKanbanColumnId): string {
  switch (columnId) {
    case 'closed':
      return 'closed'
    case 'failed':
      return 'failed'
    case 'pickup':
      return 'tomorrow'
    case 'delivery':
      return 'later'
    case 'production':
    default:
      return 'today'
  }
}

export function resolveDealKanbanColumnId(deal: Deal): DealKanbanColumnId {
  return mapDealStatusToColumnId(deal.status)
}

export function inferInitialDealColumnId(deal: Deal): DealKanbanColumnId {
  if (isDealProductionComplete(deal)) return 'production'
  if (isDealDeliveryComplete(deal)) return 'delivery'
  if (isDealPickupComplete(deal)) return 'pickup'
  return 'production'
}

export interface DealColumnValidationResult {
  message: string
  targetSection: 'production' | 'pickup' | 'delivery'
}

export function getDealColumnValidationResult(
  deal: Deal,
  columnId: DealKanbanColumnId,
): DealColumnValidationResult | null {
  if (columnId === 'production') {
    if (!isDealProductionComplete(deal)) {
      return {
        message:
          'Чтобы перевести сделку в «Производство», заполните номенклатуру, дату и время производства и сотрудника в разделе «Производство».',
        targetSection: 'production',
      }
    }
    return null
  }

  if (columnId === 'pickup') {
    if (isPickupSectionLocked(deal.pickupDelivery)) {
      return {
        message: PICKUP_SECTION_LOCKED_MESSAGE,
        targetSection: 'delivery',
      }
    }
    if (!isDealPickupComplete(deal)) {
      return {
        message:
          'Чтобы перевести сделку в «Самовывоз», заполните пожалуйста все поля данного раздела.',
        targetSection: 'pickup',
      }
    }
    return null
  }

  if (columnId === 'delivery') {
    if (isDeliverySectionLocked(deal.pickupDelivery)) {
      return {
        message: DELIVERY_SECTION_LOCKED_MESSAGE,
        targetSection: 'pickup',
      }
    }
    if (!isDealDeliveryComplete(deal)) {
      return {
        message:
          'Чтобы перевести сделку в «Доставка», заполните пожалуйста все поля данного раздела.',
        targetSection: 'delivery',
      }
    }
    return null
  }

  return null
}

export function getDealColumnValidationMessage(deal: Deal, columnId: DealKanbanColumnId): string | null {
  return getDealColumnValidationResult(deal, columnId)?.message ?? null
}

export function canMoveDealToColumn(deal: Deal, columnId: DealKanbanColumnId): boolean {
  return getDealColumnValidationMessage(deal, columnId) === null
}

export function mapColumnIdToDealSection(columnId: DealKanbanColumnId): 'production' | 'pickup' | 'delivery' | null {
  if (columnId === 'production' || columnId === 'pickup' || columnId === 'delivery') {
    return columnId
  }
  return null
}
