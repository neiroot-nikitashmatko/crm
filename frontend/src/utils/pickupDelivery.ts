import type { PickupDelivery } from '@/types/deal'

export function hasPickupData(pickupDelivery: PickupDelivery): boolean {
  return (
    pickupDelivery.pickupAddress.trim() !== '' ||
    (pickupDelivery.pickupDate !== null && pickupDelivery.pickupDate !== undefined)
  )
}

export function hasDeliveryData(pickupDelivery: PickupDelivery): boolean {
  return (
    pickupDelivery.deliveryAddress.trim() !== '' ||
    (pickupDelivery.deliveryDate !== null && pickupDelivery.deliveryDate !== undefined) ||
    pickupDelivery.courier.trim() !== ''
  )
}

export function isPickupSectionLocked(pickupDelivery: PickupDelivery): boolean {
  return hasDeliveryData(pickupDelivery)
}

export function isDeliverySectionLocked(pickupDelivery: PickupDelivery): boolean {
  return hasPickupData(pickupDelivery)
}

export const PICKUP_SECTION_LOCKED_MESSAGE =
  'Перенос сделки в раздел «Самовывоз» недоступен, так как уже заполнен раздел «Доставка». Очистите поля в разделе «Доставка» и попробуйте снова.'

export const DELIVERY_SECTION_LOCKED_MESSAGE =
  'Перенос сделки в раздел «Доставка» недоступен, так как уже заполнен раздел «Самовывоз». Очистите поля в разделе «Самовывоз» и попробуйте снова.'
