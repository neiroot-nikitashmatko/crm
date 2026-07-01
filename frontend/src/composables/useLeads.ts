import { ref } from 'vue'
import {
  createLead,
  deleteLead as deleteLeadRequest,
  fetchLeads,
  updateLeadColumn,
  updateLeadComment as patchLeadComment,
  updateLeadPickupDelivery as patchLeadPickupDelivery,
  updateLeadProducts as patchLeadProducts,
  updateLeadProduction as patchLeadProduction,
} from '@/api/leads'
import { useAuth } from '@/composables/useAuth'
import type { DealProduct, PickupDelivery } from '@/types/deal'
import type { Lead, LeadProduction, NewLeadForm } from '@/types/lead'

const leads = ref<Lead[]>([])
const isLoaded = ref(false)
const isLoading = ref(false)

const emptyPickupDelivery = (): PickupDelivery => ({
  pickupAddress: '',
  pickupDate: null,
  deliveryAddress: '',
  deliveryDate: null,
  courier: '',
})

const emptyProduction = (): LeadProduction => ({
  nomenclature: '',
  dueAt: null,
  employee: '',
})

function normalizeProducts(raw: any): DealProduct[] {
  const items = Array.isArray(raw?.products) ? raw.products : []
  return items.map((item: any) => ({
    title: String(item?.title ?? ''),
    quantity: Number(item?.quantity ?? 1),
    unitPrice: Number(item?.unitPrice ?? item?.unit_price ?? 0),
  }))
}

function normalizeProduction(raw: any): LeadProduction {
  const nested = raw?.production ?? raw
  return {
    nomenclature: String(nested?.nomenclature ?? nested?.production_nomenclature ?? ''),
    dueAt: resolveTimestamp(nested?.dueAt ?? nested?.due_at ?? nested?.production_due_at),
    employee: String(nested?.employee ?? nested?.production_employee ?? ''),
  }
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

export function normalizeLead(raw: any): Lead {
  return {
    id: String(raw.id),
    leadNumber: Number(raw.leadNumber ?? raw.lead_number ?? 0),
    firstName: String(raw.firstName ?? raw.first_name ?? ''),
    patronymic: String(raw.patronymic ?? ''),
    phone: String(raw.phone ?? ''),
    trafficSource: String(raw.trafficSource ?? raw.traffic_source ?? ''),
    columnId: String(raw.columnId ?? raw.column_id ?? 'new'),
    leadComments: String(raw.leadComments ?? raw.lead_comments ?? ''),
    failureReason: String(raw.failureReason ?? raw.failure_reason ?? ''),
    createdBy: String(raw.createdBy ?? raw.created_by ?? ''),
    createdAt: Number(raw.createdAt ?? raw.created_at ?? Date.now()),
    updatedAt: Number(raw.updatedAt ?? raw.updated_at ?? Date.now()),
    pickupDelivery: normalizePickupDelivery(raw),
    products: normalizeProducts(raw),
    production: normalizeProduction(raw),
  }
}

function applyLeadUpdate(lead: Lead, raw: any) {
  Object.assign(lead, normalizeLead({ ...raw, id: lead.id }))
}

export function useLeads() {
  const { user } = useAuth()

  async function loadLeads(force = false) {
    if (isLoading.value) return
    if (isLoaded.value && !force) return

    isLoading.value = true
    try {
      const items = await fetchLeads()
      leads.value = items.map(normalizeLead)
      isLoaded.value = true
    } finally {
      isLoading.value = false
    }
  }

  async function addLead(payload: NewLeadForm, columnId = 'new') {
    if (!user.value) {
      throw new Error('Пользователь не авторизован')
    }

    const createdLead = normalizeLead(await createLead(payload, columnId, user.value.id))
    leads.value.unshift(createdLead)
  }

  async function moveLeadToColumn(leadId: string, columnId: string, failureReason?: string) {
    const lead = leads.value.find((item) => item.id === leadId)
    if (!lead) return

    const updatedLead = normalizeLead(await updateLeadColumn(leadId, columnId, failureReason))
    if (lead) applyLeadUpdate(lead, updatedLead)
  }

  async function updateLeadComment(leadId: string, leadComments: string) {
    const lead = leads.value.find((item) => item.id === leadId)
    if (!lead) return

    const updatedLead = normalizeLead(await patchLeadComment(leadId, leadComments))
    applyLeadUpdate(lead, updatedLead)
  }

  async function updateLeadPickupDelivery(leadId: string, pickupDelivery: PickupDelivery) {
    const lead = leads.value.find((item) => item.id === leadId)
    if (!lead) return normalizeLead(await patchLeadPickupDelivery(leadId, pickupDelivery))

    const updatedLead = normalizeLead(await patchLeadPickupDelivery(leadId, pickupDelivery))
    applyLeadUpdate(lead, updatedLead)
    return updatedLead
  }

  async function updateLeadProducts(leadId: string, products: DealProduct[]) {
    const lead = leads.value.find((item) => item.id === leadId)
    if (!lead) return normalizeLead(await patchLeadProducts(leadId, products))

    const updatedLead = normalizeLead(await patchLeadProducts(leadId, products))
    applyLeadUpdate(lead, updatedLead)
    return updatedLead
  }

  async function updateLeadProduction(leadId: string, production: LeadProduction) {
    const lead = leads.value.find((item) => item.id === leadId)
    if (!lead) return normalizeLead(await patchLeadProduction(leadId, production))

    const updatedLead = normalizeLead(await patchLeadProduction(leadId, production))
    applyLeadUpdate(lead, updatedLead)
    return updatedLead
  }

  async function deleteLead(leadId: string) {
    await deleteLeadRequest(leadId)
    leads.value = leads.value.filter((lead) => lead.id !== leadId)
  }

  return {
    leads,
    isLoaded,
    isLoading,
    loadLeads,
    addLead,
    moveLeadToColumn,
    updateLeadComment,
    updateLeadPickupDelivery,
    updateLeadProducts,
    updateLeadProduction,
    deleteLead,
  }
}

export { emptyPickupDelivery, emptyProduction }
