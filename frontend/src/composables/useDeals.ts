import { computed, ref } from 'vue'
import {
  createDealFromLead as createDealFromLeadRequest,
  deleteDeal as deleteDealRequest,
  fetchDeals,
  updateDealComment as updateDealCommentRequest,
  updateDealPickupDelivery as updateDealPickupDeliveryRequest,
  updateDealProduction as updateDealProductionRequest,
  updateDealProductionDueAt as updateDealProductionDueAtRequest,
  updateDealProducts as updateDealProductsRequest,
  updateDealProfile as updateDealProfileRequest,
  updateDealStatus as updateDealStatusRequest,
} from '@/api/deals'
import { deleteAttachment, uploadDealAttachments } from '@/api/attachments'
import type { Deal, DealAttachment, DealKanbanColumnId, DealProduct, PickupDelivery } from '@/types/deal'
import { normalizeStoredActivity, normalizeStoredAttachment, type StoredActivity } from '@/types/attachment'
import type { Lead } from '@/types/lead'
import { inferInitialDealColumnId, mapDealStatusToColumnId } from '@/utils/dealKanban'

const deals = ref<Deal[]>([])
const isLoaded = ref(false)
const isLoading = ref(false)

function resolveProductionDueAt(raw: any): number | null {
  const direct = raw?.productionDueAt ?? raw?.production_due_at
  if (direct !== null && direct !== undefined) {
    const value = Number(direct)
    return Number.isNaN(value) ? null : value
  }

  const nested = raw?.production?.dueAt ?? raw?.production?.due_at
  if (nested !== null && nested !== undefined) {
    const value = Number(nested)
    return Number.isNaN(value) ? null : value
  }

  return null
}

function resolveTimestamp(raw: unknown): number | null {
  if (raw === null || raw === undefined) return null
  const value = Number(raw)
  return Number.isNaN(value) ? null : value
}

function normalizePickupDelivery(raw: any): PickupDelivery {
  const nested = raw?.pickupDelivery ?? raw
  return {
    pickupAddress: String(nested?.pickupAddress ?? nested?.pickup_address ?? ''),
    pickupDate: resolveTimestamp(nested?.pickupDate ?? nested?.pickup_date),
    deliveryAddress: String(nested?.deliveryAddress ?? nested?.delivery_address ?? ''),
    deliveryDate: resolveTimestamp(nested?.deliveryDate ?? nested?.delivery_date),
    courier: String(nested?.courier ?? ''),
  }
}

function applyDealKanbanColumn(deal: Deal): Deal {
  return {
    ...deal,
    columnId: mapDealStatusToColumnId(deal.status),
  }
}

function normalizeDeal(raw: any): Deal {
  const rawStatus = String(raw?.status ?? raw?.column_id ?? raw?.columnId ?? 'today').toLowerCase()

  const dueAt = resolveProductionDueAt(raw)

  const createdAt = raw.createdAt ?? raw.created_at
  const firstName = raw.firstName ?? raw.first_name
  const trafficSource = raw.trafficSource ?? raw.traffic_source
  const dealComments = raw.dealComments ?? raw.deal_comments
  const failureReason = raw.failureReason ?? raw.failure_reason
  const dealNumber = raw.dealNumber ?? raw.deal_number
  const createdBy = raw.createdBy ?? raw.created_by
  const createdByName = raw.createdByName ?? raw.created_by_name
  const leadId = raw.leadId ?? raw.lead_id
  const totalAmount = raw.totalAmount ?? raw.total_amount
  const patronymic = raw.patronymic ?? ''
  const phone = raw.phone ?? ''
  const rawProducts = raw.products ?? []
  const rawProduction = raw.production ?? raw

  const deal: Deal = {
    id: String(raw.id),
    leadId: leadId ? String(leadId) : undefined,
    dealNumber: Number(dealNumber ?? 0),
    firstName: String(firstName ?? ''),
    patronymic: String(patronymic),
    phone: String(phone),
    trafficSource: String(trafficSource ?? ''),
    totalAmount: Number(totalAmount ?? 0),
    dealComments: String(dealComments ?? ''),
    failureReason: String(failureReason ?? ''),
    createdAt: Number(createdAt ?? Date.now()),
    createdBy: String(createdBy ?? ''),
    createdByName: String(createdByName ?? '').trim(),
    products: Array.isArray(rawProducts)
      ? rawProducts.map((item: any) => ({
          title: String(item?.title ?? item?.name ?? ''),
          quantity: Number(item?.quantity ?? item?.qty ?? 1),
          unitPrice: Number(item?.unitPrice ?? item?.unit_price ?? item?.price ?? 0),
        }))
      : [],
    production: {
      nomenclature: String(rawProduction?.nomenclature ?? rawProduction?.production_nomenclature ?? ''),
      dueAt,
      employee: String(rawProduction?.employee ?? rawProduction?.production_employee ?? ''),
    },
    productionDueAt: dueAt,
    pickupDelivery: normalizePickupDelivery(raw),
    attachments: Array.isArray(raw.attachments)
      ? raw.attachments.map((attachment: unknown) =>
          normalizeStoredAttachment(attachment, String(createdBy ?? '')),
        )
      : [],
    activities: Array.isArray(raw.activities)
      ? raw.activities.map((activity: unknown) =>
          normalizeStoredActivity(activity, String(createdBy ?? '')),
        )
      : [],
    columnId: 'pickup',
    status: rawStatus,
  }

  return applyDealKanbanColumn(deal)
}

function mergeDealLocalState(next: Deal, prev: Deal | undefined): Deal {
  if (!prev) return applyDealKanbanColumn(next)
  return applyDealKanbanColumn(next)
}

function prependDealActivity(dealId: string, activity: StoredActivity) {
  deals.value = deals.value.map((item) =>
    item.id === dealId ? { ...item, activities: [activity, ...item.activities] } : item,
  )
}

export function useDeals() {
  const hasLoadedOnce = computed(() => isLoaded.value)

  async function loadDeals(force = false) {
    if ((isLoaded.value && !force) || isLoading.value) return
    isLoading.value = true
    try {
      const items = await fetchDeals()
      deals.value = items.map(normalizeDeal)
      isLoaded.value = true
    } finally {
      isLoading.value = false
    }
  }

  function getLeadDeal(leadId: string) {
    return deals.value.find((item) => item.leadId === leadId)
  }

  function getActiveLeadDeal(leadId: string) {
    return deals.value.find(
      (item) =>
        item.leadId === leadId &&
        item.status !== 'failed' &&
        item.columnId !== 'failed',
    )
  }

  async function createDealFromLead(
    leadOrPayload: Lead | { leadId: string; createdBy: string },
    options?: {
      products?: DealProduct[]
      production?: { nomenclature: string; dueAt: number | null; employee: string }
      pickupDelivery?: PickupDelivery
    },
  ) {
    const pickupDelivery = options?.pickupDelivery ?? {
      pickupAddress: '',
      pickupDate: null,
      deliveryAddress: '',
      deliveryDate: null,
      courier: '',
    }
    const payload =
      'leadId' in leadOrPayload
        ? {
            leadId: leadOrPayload.leadId,
            createdBy: leadOrPayload.createdBy,
            products: options?.products ?? [],
            production: options?.production ?? { nomenclature: '', dueAt: null, employee: '' },
            pickupDelivery,
          }
        : {
            leadId: leadOrPayload.id,
            createdBy: leadOrPayload.createdBy,
            products: options?.products ?? [],
            production: options?.production ?? { nomenclature: '', dueAt: null, employee: '' },
            pickupDelivery,
          }

    const created = applyDealKanbanColumn(normalizeDeal(await createDealFromLeadRequest(payload)))

    const initialColumnId = inferInitialDealColumnId(created)
    const synced =
      initialColumnId === created.columnId
        ? created
        : applyDealKanbanColumn(
            normalizeDeal(await updateDealStatusRequest(created.id, initialColumnId)),
          )
    deals.value.unshift(synced)
    return synced
  }

  async function updateDealComment(dealId: string, dealComments: string) {
    const prev = deals.value.find((item) => item.id === dealId)
    const updated = mergeDealLocalState(normalizeDeal(await updateDealCommentRequest(dealId, dealComments)), prev)
    deals.value = deals.value.map((item) => (item.id === dealId ? updated : item))
    return updated
  }

  async function updateDealProfile(dealId: string, firstName: string, patronymic: string) {
    const prev = deals.value.find((item) => item.id === dealId)
    const updated = mergeDealLocalState(
      normalizeDeal(await updateDealProfileRequest(dealId, firstName, patronymic)),
      prev,
    )
    deals.value = deals.value.map((item) => (item.id === dealId ? updated : item))
    return updated
  }

  async function updateDealProductionDueAt(dealId: string, dueAt: number | null) {
    const prev = deals.value.find((item) => item.id === dealId)
    const updated = mergeDealLocalState(
      normalizeDeal(await updateDealProductionDueAtRequest(dealId, dueAt)),
      prev,
    )
    deals.value = deals.value.map((item) => (item.id === dealId ? updated : item))
    return updated
  }

  async function updateDealProduction(dealId: string, production: Deal['production']) {
    const prev = deals.value.find((item) => item.id === dealId)
    const updated = mergeDealLocalState(
      normalizeDeal(await updateDealProductionRequest(dealId, production)),
      prev,
    )
    deals.value = deals.value.map((item) => (item.id === dealId ? updated : item))
    return updated
  }

  async function updateDealPickupDelivery(dealId: string, pickupDelivery: PickupDelivery) {
    const prev = deals.value.find((item) => item.id === dealId)
    const updated = mergeDealLocalState(
      normalizeDeal(await updateDealPickupDeliveryRequest(dealId, pickupDelivery)),
      prev,
    )
    deals.value = deals.value.map((item) => (item.id === dealId ? updated : item))
    return updated
  }

  async function updateDealProducts(dealId: string, products: DealProduct[]) {
    const updated = applyDealKanbanColumn(normalizeDeal(await updateDealProductsRequest(dealId, products)))
    deals.value = deals.value.map((item) => (item.id === dealId ? updated : item))
    return updated
  }

  async function updateDealStatus(dealId: string, columnId: DealKanbanColumnId, failureReason?: string) {
    const prev = deals.value.find((item) => item.id === dealId)
    const updated = mergeDealLocalState(
      normalizeDeal(await updateDealStatusRequest(dealId, columnId, failureReason)),
      prev,
    )
    deals.value = deals.value.map((item) => (item.id === dealId ? updated : item))
    return updated
  }

  async function deleteDeal(dealId: string) {
    await deleteDealRequest(dealId)
    deals.value = deals.value.filter((item) => item.id !== dealId)
  }

  async function addDealAttachments(dealId: string, files: File[]): Promise<DealAttachment[]> {
    const { items, activity } = await uploadDealAttachments(dealId, files)
    deals.value = deals.value.map((item) =>
      item.id === dealId ? { ...item, attachments: [...items, ...item.attachments] } : item,
    )
    if (activity) {
      prependDealActivity(dealId, activity)
    }
    return items
  }

  async function removeDealAttachment(dealId: string, attachmentId: string): Promise<void> {
    await deleteAttachment(attachmentId)
    deals.value = deals.value.map((item) =>
      item.id === dealId
        ? { ...item, attachments: item.attachments.filter((attachment) => attachment.id !== attachmentId) }
        : item,
    )
  }

  return {
    deals,
    hasLoadedOnce,
    loadDeals,
    getLeadDeal,
    getActiveLeadDeal,
    createDealFromLead,
    updateDealComment,
    updateDealProfile,
    updateDealProduction,
    updateDealProductionDueAt,
    updateDealPickupDelivery,
    updateDealProducts,
    updateDealStatus,
    deleteDeal,
    addDealAttachments,
    removeDealAttachment,
  }
}
