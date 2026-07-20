<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NDatePicker, NIcon, NSelect } from 'naive-ui'
import { TrashOutline } from '@vicons/ionicons5'
import { getApiBaseUrl } from '@/api/httpClient'
import { LeadsApiError } from '@/api/leads'
import { LEAD_KANBAN_COLUMNS } from '@/constants/leads'
import { PRODUCTION_NOMENCLATURE_OPTIONS } from '@/constants/production'
import { useAuth } from '@/composables/useAuth'
import { getAuthToken } from '@/api/session'
import { useDeals } from '@/composables/useDeals'
import { normalizeLead, useLeads, emptyPickupDelivery, emptyProduction } from '@/composables/useLeads'
import { useLeadsKanbanLayout } from '@/composables/useLeadsKanbanLayout'
import { useDealProductRows } from '@/composables/useDealProductRows'
import { useLeadProductRows } from '@/composables/useLeadProductRows'
import { useTasks } from '@/composables/useTasks'
import type { Lead, LeadProduction, NewLeadForm } from '@/types/lead'
import type { Task } from '@/types/task'
import type { PickupDelivery } from '@/types/deal'
import type { LeadKanbanColumn } from '@/constants/leads'
import {
  DELIVERY_SECTION_LOCKED_MESSAGE,
  isDeliverySectionLocked,
  isPickupSectionLocked,
  PICKUP_SECTION_LOCKED_MESSAGE,
} from '@/utils/pickupDelivery'
import LeadsKanbanColumn from './LeadsKanbanColumn.vue'
import LeadAvitoChatPanel from './LeadAvitoChatPanel.vue'
import TaskDetailsSheet from '../tasks/TaskDetailsSheet.vue'
import DateTimeField from '@/components/common/DateTimeField.vue'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import EntityAttachmentList from '@/components/attachments/EntityAttachmentList.vue'
import DealProductsEditor from '@/components/common/DealProductsEditor.vue'
import type { ProductRow } from '@/types/productRow'
import { rowsToDealProducts } from '@/utils/products'
import { playNewLeadSound, unlockNewLeadSound, isNewLeadSoundUnlocked } from '@/utils/newLeadSound'

const props = withDefaults(
  defineProps<{
    searchQuery?: string
  }>(),
  {
    searchQuery: '',
  },
)

type LeadDetailsSectionId =
  | 'lead-info'
  | 'chat'
  | 'task'
  | 'products'
  | 'pickup'
  | 'delivery'
  | 'production'

const LEAD_DETAILS_SECTIONS: Array<{ id: LeadDetailsSectionId; title: string }> = [
  { id: 'lead-info', title: 'Информация о лиде' },
  { id: 'chat', title: 'Чат' },
  { id: 'task', title: 'Задача' },
  { id: 'products', title: 'Услуги/Товары' },
  { id: 'pickup', title: 'Самовывоз' },
  { id: 'delivery', title: 'Доставка' },
  { id: 'production', title: 'Производство' },
]

function isAvitoChatLead(lead: Lead | null | undefined): boolean {
  return (lead?.trafficSource ?? '').trim() === 'Авито Чат'
}
const WORKING_HOURS = [9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]
const TIME_MINUTE_STEP = 5
const TIME_PICKER_PROPS = {
  format: 'HH:mm',
  hours: WORKING_HOURS,
  minutes: TIME_MINUTE_STEP,
  actions: ['confirm'] as Array<'confirm'>,
}

const router = useRouter()
const route = useRoute()
const { logout } = useAuth()
const {
  leads,
  addLead,
  moveLeadToColumn,
  updateLeadComment,
  updateLeadProfile,
  addLeadAttachments,
  removeLeadAttachment,
  updateLeadPickupDelivery,
  updateLeadProduction,
  loadLeads,
  deleteLead,
} = useLeads()
const { addTask, completeLeadTasks, getLeadTasks, loadTasks } = useTasks()
const { createDealFromLead, getActiveLeadDeal, getLeadDeal, loadDeals, deals } = useDeals()
const {
  getDealRows,
  setDealRows,
  hydrateDealRows,
  saveDealProductRows,
  applySavedProducts,
} = useDealProductRows()
const {
  getLeadRows,
  setLeadRows,
  hydrateLeadRows,
  saveLeadProductRows,
  resetLeadRows,
} = useLeadProductRows()
const { kanbanHeightPx, updateKanbanHeight } = useLeadsKanbanLayout()
const selectedLeadId = ref<string | null>(null)
const selectedLeadTask = ref<Task | null>(null)
const activeDetailsSection = ref<LeadDetailsSectionId>('lead-info')
const isCreateTaskModalOpen = ref(false)
const isTaskDateModalOpen = ref(false)
const isProductionDateModalOpen = ref(false)
const isPickupDateModalOpen = ref(false)
const isDeliveryDateModalOpen = ref(false)
const isFailureReasonModalOpen = ref(false)
const failureReasonDraft = ref('')
const taskForm = reactive({
  title: '',
  text: '',
  dueAt: null as number | null,
})
const persistedLeadPickupDeliveryJSON = ref('')
const persistedLeadProductionJSON = ref('')
let leadPickupDeliverySaveTimer: ReturnType<typeof setTimeout> | null = null
let leadProductsSaveTimer: ReturnType<typeof setTimeout> | null = null
let leadProductionSaveTimer: ReturnType<typeof setTimeout> | null = null
const productionNomenclatureOptions = [...PRODUCTION_NOMENCLATURE_OPTIONS]
const productionEmployeeOptions = [
  { label: 'Никита Хачересов', value: 'Никита Хачересов' },
  { label: 'Сергей Геворкян', value: 'Сергей Геворкян' },
]
const selectedLead = computed(() =>
  selectedLeadId.value
    ? leads.value.find((lead) => lead.id === selectedLeadId.value) ?? null
    : null,
)

const visibleLeadDetailsSections = computed(() =>
  LEAD_DETAILS_SECTIONS.filter(
    (section) => section.id !== 'chat' || isAvitoChatLead(selectedLead.value),
  ),
)
const canConfirmFailureReason = computed(() => failureReasonDraft.value.trim().length > 0)
const leadCommentDrafts = reactive<Record<string, string>>({})
const leadProfileDrafts = reactive<Record<string, { firstName: string; patronymic: string }>>({})
const currentLeadComment = computed({
  get: () => {
    if (!selectedLead.value) return ''
    return leadCommentDrafts[selectedLead.value.id] ?? selectedLead.value.leadComments ?? ''
  },
  set: (value: string) => {
    if (!selectedLead.value) return
    leadCommentDrafts[selectedLead.value.id] = value
  },
})
const currentLeadFirstName = computed({
  get: () => {
    if (!selectedLead.value) return ''
    return leadProfileDrafts[selectedLead.value.id]?.firstName ?? selectedLead.value.firstName ?? ''
  },
  set: (value: string) => {
    if (!selectedLead.value) return
    const current = leadProfileDrafts[selectedLead.value.id] ?? {
      firstName: selectedLead.value.firstName ?? '',
      patronymic: selectedLead.value.patronymic ?? '',
    }
    leadProfileDrafts[selectedLead.value.id] = { ...current, firstName: value }
  },
})
const currentLeadPatronymic = computed({
  get: () => {
    if (!selectedLead.value) return ''
    return leadProfileDrafts[selectedLead.value.id]?.patronymic ?? selectedLead.value.patronymic ?? ''
  },
  set: (value: string) => {
    if (!selectedLead.value) return
    const current = leadProfileDrafts[selectedLead.value.id] ?? {
      firstName: selectedLead.value.firstName ?? '',
      patronymic: selectedLead.value.patronymic ?? '',
    }
    leadProfileDrafts[selectedLead.value.id] = { ...current, patronymic: value }
  },
})
const currentLeadProduction = computed<LeadProduction>(() => {
  if (!selectedLead.value) {
    return emptyProduction()
  }
  return selectedLead.value.production
})
const currentLeadPickupDelivery = computed<PickupDelivery>(() => {
  if (!selectedLead.value) {
    return emptyPickupDelivery()
  }
  return selectedLead.value.pickupDelivery
})
const leadProductRows = computed<ProductRow[]>({
  get() {
    if (!selectedLead.value) return []

    if (activeLeadDeal.value) {
      return getDealRows(activeLeadDeal.value.id)
    }

    return getLeadRows(selectedLead.value.id)
  },
  set(value) {
    if (!selectedLead.value) return

    if (activeLeadDeal.value) {
      setDealRows(activeLeadDeal.value.id, value)
      return
    }

    setLeadRows(selectedLead.value.id, value)
  },
})
const linkedLeadDeal = computed(() =>
  selectedLead.value ? getLeadDeal(selectedLead.value.id) ?? null : null,
)
const activeLeadDeal = computed(() =>
  selectedLead.value ? getActiveLeadDeal(selectedLead.value.id) ?? null : null,
)
const isLeadLinkedToActiveDeal = computed(() => activeLeadDeal.value !== null)
const hasLinkedDeal = computed(() => linkedLeadDeal.value !== null)
const canCreateDeal = computed(() => selectedLead.value !== null && !hasLinkedDeal.value)
const leadTasks = computed<Task[]>(() => {
  if (!selectedLead.value) {
    return []
  }

  return [...getLeadTasks(selectedLead.value.id)].sort((left, right) => {
    if (left.status !== right.status) {
      return left.status === 'active' ? -1 : 1
    }

    const leftDue = left.dueAt ?? Number.MAX_SAFE_INTEGER
    const rightDue = right.dueAt ?? Number.MAX_SAFE_INTEGER
    if (leftDue !== rightDue) {
      return leftDue - rightDue
    }

    return right.createdAt.getTime() - left.createdAt.getTime()
  })
})
const sortedLeadActivities = computed(() => {
  if (!selectedLead.value) return []
  return [...selectedLead.value.activities].sort((a, b) => b.createdAt - a.createdAt)
})

function phoneDigits(value: string): string {
  return value.replace(/\D/g, '').replace(/^8/, '7')
}

function leadMatchesSearch(lead: Lead, query: string): boolean {
  const normalizedQuery = query.trim().toLowerCase()
  if (!normalizedQuery) {
    return true
  }

  const fullName = `${lead.firstName} ${lead.patronymic}`.trim().toLowerCase()
  if (fullName.includes(normalizedQuery)) {
    return true
  }

  const queryDigits = phoneDigits(normalizedQuery)
  if (queryDigits.length > 0 && phoneDigits(lead.phone).includes(queryDigits)) {
    return true
  }

  return false
}

function getColumnLeads(columnId: string) {
  return leads.value.filter(
    (lead) => lead.columnId === columnId && leadMatchesSearch(lead, props.searchQuery),
  )
}

async function handleAddLead(columnId: string, payload: NewLeadForm) {
  // Unlock audio while we still have the user-gesture stack (before await).
  void unlockNewLeadSound()
  try {
    await addLead(payload, columnId)
    if (columnId === 'new') {
      void playNewLeadSound()
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Не удалось создать лид'
    console.error('Не удалось создать лид', error)
    window.alert(message)
    if (error instanceof LeadsApiError && message.includes('Сессия устарела')) {
      logout()
      await router.push('/login')
    }
  }
}

function handleOpenLead(lead: Lead) {
  selectedLeadId.value = lead.id
  syncLeadCommentDraft(lead)
  syncLeadProfileDraft(lead)
  syncLeadPickupDeliverySnapshot()
  syncLeadProductionSnapshot()
  hydrateLeadRows(lead.id, true)
  activeDetailsSection.value = 'lead-info'
  isFailureReasonModalOpen.value = false
  failureReasonDraft.value = ''
}

function closeLeadDetails() {
  selectedLeadId.value = null
  selectedLeadTask.value = null
  isCreateTaskModalOpen.value = false
  isTaskDateModalOpen.value = false
  isFailureReasonModalOpen.value = false
  failureReasonDraft.value = ''

  if (route.query.leadId) {
    void router.replace({ name: 'leads' })
  }
}

async function openLeadFromRouteQuery() {
  const rawLeadId = route.query.leadId
  if (typeof rawLeadId !== 'string' || rawLeadId.trim() === '') {
    return
  }

  const leadId = rawLeadId.trim()
  if (!leads.value.some((lead) => lead.id === leadId)) {
    await loadLeads(true)
  }

  const lead = leads.value.find((item) => item.id === leadId)
  if (lead) {
    handleOpenLead(lead)
  }
}

async function handleDeleteLeadClick() {
  if (!selectedLead.value) {
    return
  }

  try {
    await deleteLead(selectedLead.value.id)
    closeLeadDetails()
  } catch (error) {
    console.error('Не удалось удалить лид', error)
  }
}

async function moveSelectedLead(columnId: string) {
  if (!selectedLead.value) return
  if (selectedLead.value.columnId === columnId) return

  if (columnId === 'failed') {
    failureReasonDraft.value = ''
    isFailureReasonModalOpen.value = true
    return
  }

  try {
    await moveLeadToColumn(selectedLead.value.id, columnId)
  } catch (error) {
    console.error('Не удалось обновить статус лида', error)
  }
}

function closeFailureReasonModal() {
  isFailureReasonModalOpen.value = false
  failureReasonDraft.value = ''
}

async function confirmFailureReason() {
  if (!selectedLead.value || !canConfirmFailureReason.value) return

  const leadId = selectedLead.value.id
  const reason = failureReasonDraft.value.trim()

  try {
    await moveLeadToColumn(leadId, 'failed', reason)
    closeFailureReasonModal()
  } catch (error) {
    console.error('Не удалось перевести лид в проваленные', error)
  }
}

function getStatusButtonStyle(column: LeadKanbanColumn) {
  return {
    '--status-color': column.style.countColor,
    '--status-bg': column.style.countBg,
    '--status-border': column.style.headerBorder,
    '--status-header-bg': column.style.headerBg,
  }
}

function getStatusButtonClass(column: LeadKanbanColumn) {
  return {
    'lead-details-sheet__status-btn--active': selectedLead.value?.columnId === column.id,
  }
}

function formatLeadCreatedAt(timestamp: number) {
  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(timestamp)
}

function formatDateTime(timestamp: number | null) {
  if (timestamp === null) return ''

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(timestamp)
}

function getColumnTitleById(columnId: string) {
  return LEAD_KANBAN_COLUMNS.find((column) => column.id === columnId)?.title ?? '—'
}

function openCreateTaskModal() {
  if (isLeadLinkedToActiveDeal.value) {
    return
  }
  resetTaskForm()
  isCreateTaskModalOpen.value = true
}

function closeCreateTaskModal() {
  isCreateTaskModalOpen.value = false
  isTaskDateModalOpen.value = false
  resetTaskForm()
}

function resetTaskForm() {
  taskForm.title = ''
  taskForm.text = ''
  taskForm.dueAt = null
}

function getLeadTaskStatusLabel(status: Task['status']) {
  return status === 'completed' ? 'Завершена' : 'Активна'
}

function openLeadTaskDetails(task: Task) {
  selectedLeadTask.value = task
}

function closeLeadTaskDetails() {
  selectedLeadTask.value = null
}

function schedulePersistLeadProducts() {
  if (!selectedLead.value) return

  // Для лида (до создания сделки) сохраняем сразу, чтобы refresh страницы не терял изменения.
  if (!activeLeadDeal.value) {
    void persistLeadProducts()
    return
  }

  if (leadProductsSaveTimer) {
    clearTimeout(leadProductsSaveTimer)
  }
  leadProductsSaveTimer = setTimeout(() => {
    leadProductsSaveTimer = null
    void persistLeadProducts()
  }, 500)
}

async function flushPersistLeadProducts() {
  if (leadProductsSaveTimer) {
    clearTimeout(leadProductsSaveTimer)
    leadProductsSaveTimer = null
  }
  await persistLeadProducts()
}

async function persistLeadProducts() {
  if (!selectedLead.value) return

  try {
    if (activeLeadDeal.value) {
      await saveDealProductRows(activeLeadDeal.value.id)
    } else {
      await saveLeadProductRows(selectedLead.value.id)
    }
  } catch (error) {
    console.error('Не удалось сохранить товары', error)
  }
}

function isTaskDateFilled(value: unknown) {
  if (value === null || value === undefined) return false
  if (typeof value === 'number') return !Number.isNaN(value)
  if (typeof value === 'string') return value.trim().length > 0
  return true
}

const canCreateTask = computed(
  () =>
    taskForm.title.trim().length > 0 &&
    taskForm.text.trim().length > 0 &&
    isTaskDateFilled(taskForm.dueAt),
)
const currentTaskDueAt = computed({
  get: () => taskForm.dueAt,
  set: (value: number | null) => {
    taskForm.dueAt = normalizeDateTimeValue(value)
  },
})
const currentLeadNomenclature = computed({
  get: () => currentLeadProduction.value.nomenclature,
  set: (value: string | null) => {
    currentLeadProduction.value.nomenclature = value ?? ''
  },
})
const currentLeadEmployee = computed({
  get: () => currentLeadProduction.value.employee,
  set: (value: string | null) => {
    currentLeadProduction.value.employee = value ?? ''
  },
})
const currentLeadProductionDueAt = computed({
  get: () => currentLeadProduction.value.dueAt,
  set: (value: number | null) => {
    currentLeadProduction.value.dueAt = normalizeDateTimeValue(value)
  },
})
const currentLeadPickupAddress = computed({
  get: () => currentLeadPickupDelivery.value.pickupAddress,
  set: (value: string) => {
    currentLeadPickupDelivery.value.pickupAddress = value
  },
})
const currentLeadPickupDate = computed({
  get: () => currentLeadPickupDelivery.value.pickupDate,
  set: (value: number | null) => {
    currentLeadPickupDelivery.value.pickupDate = normalizeDateTimeValue(value)
  },
})
const currentLeadDeliveryAddress = computed({
  get: () => currentLeadPickupDelivery.value.deliveryAddress,
  set: (value: string) => {
    currentLeadPickupDelivery.value.deliveryAddress = value
  },
})
const currentLeadDeliveryDate = computed({
  get: () => currentLeadPickupDelivery.value.deliveryDate,
  set: (value: number | null) => {
    currentLeadPickupDelivery.value.deliveryDate = normalizeDateTimeValue(value)
  },
})
const currentLeadCourier = computed({
  get: () => currentLeadPickupDelivery.value.courier,
  set: (value: string) => {
    currentLeadPickupDelivery.value.courier = value
  },
})
const isLeadPickupSectionLocked = computed(() =>
  isPickupSectionLocked(currentLeadPickupDelivery.value),
)
const isLeadDeliverySectionLocked = computed(() =>
  isDeliverySectionLocked(currentLeadPickupDelivery.value),
)
const isLeadPickupFieldsDisabled = computed(
  () => isLeadLinkedToActiveDeal.value || isLeadPickupSectionLocked.value,
)
const isLeadDeliveryFieldsDisabled = computed(
  () => isLeadLinkedToActiveDeal.value || isLeadDeliverySectionLocked.value,
)

function normalizeDateTimeValue(value: number | null) {
  if (value === null || Number.isNaN(value)) return null

  const date = new Date(value)
  let hours = date.getHours()
  let minutes = Math.round(date.getMinutes() / TIME_MINUTE_STEP) * TIME_MINUTE_STEP

  if (minutes === 60) {
    minutes = 0
    hours += 1
  }

  if (hours < WORKING_HOURS[0]) {
    hours = WORKING_HOURS[0]
    minutes = 0
  }

  if (hours > WORKING_HOURS[WORKING_HOURS.length - 1]) {
    hours = WORKING_HOURS[WORKING_HOURS.length - 1]
    minutes = 0
  }

  date.setHours(hours, minutes, 0, 0)
  return date.getTime()
}

function openTaskDateModal() {
  taskForm.dueAt = normalizeDateTimeValue(taskForm.dueAt)
  isTaskDateModalOpen.value = true
}

function handleTaskDateConfirm(onConfirm: () => void) {
  onConfirm()
  isTaskDateModalOpen.value = false
}

function openPickupDateModal() {
  if (isLeadPickupFieldsDisabled.value) return
  currentLeadPickupDate.value = normalizeDateTimeValue(currentLeadPickupDate.value)
  isPickupDateModalOpen.value = true
}

function handlePickupDateConfirm(onConfirm: () => void) {
  onConfirm()
  isPickupDateModalOpen.value = false
  flushPersistLeadPickupDelivery()
}

function openDeliveryDateModal() {
  if (isLeadDeliveryFieldsDisabled.value) return
  currentLeadDeliveryDate.value = normalizeDateTimeValue(currentLeadDeliveryDate.value)
  isDeliveryDateModalOpen.value = true
}

function handleDeliveryDateConfirm(onConfirm: () => void) {
  onConfirm()
  isDeliveryDateModalOpen.value = false
  flushPersistLeadPickupDelivery()
}

function openProductionDateModal() {
  currentLeadProductionDueAt.value = normalizeDateTimeValue(currentLeadProductionDueAt.value)
  isProductionDateModalOpen.value = true
}

function handleProductionDateConfirm(onConfirm: () => void) {
  onConfirm()
  isProductionDateModalOpen.value = false
  flushPersistLeadProduction()
}

function clearLeadPickupDate() {
  currentLeadPickupDate.value = null
  flushPersistLeadPickupDelivery()
}

function clearLeadDeliveryDate() {
  currentLeadDeliveryDate.value = null
  flushPersistLeadPickupDelivery()
}

function clearLeadProductionDueAt() {
  currentLeadProductionDueAt.value = null
  flushPersistLeadProduction()
}

async function handleCreateTaskClick() {
  if (!selectedLead.value || !canCreateTask.value || isLeadLinkedToActiveDeal.value) {
    return
  }

  try {
    await addTask({
      title: taskForm.title.trim(),
      text: taskForm.text.trim(),
      dueAt: taskForm.dueAt,
      leadId: selectedLead.value.id,
    })
    closeCreateTaskModal()
  } catch (error) {
    console.error('Не удалось создать задачу по лиду', error)
  }
}

async function handleGoToDealClick() {
  if (!linkedLeadDeal.value) {
    return
  }

  const dealId = linkedLeadDeal.value.id
  closeLeadDetails()
  await router.push({ name: 'deals', query: { dealId } })
}

async function handleCreateDealClick() {
  if (!selectedLead.value || !canCreateDeal.value) {
    return
  }

  const leadId = selectedLead.value.id

  try {
    await flushPersistLeadPickupDelivery()
    await flushPersistLeadProducts()
    await flushPersistLeadProduction()

    const createdDeal = await createDealFromLead(selectedLead.value, {
      products: rowsToDealProducts(leadProductRows.value),
      production: {
        nomenclature: currentLeadProduction.value.nomenclature,
        dueAt: currentLeadProduction.value.dueAt,
        employee: currentLeadProduction.value.employee,
      },
      pickupDelivery: { ...selectedLead.value.pickupDelivery },
    })
    applySavedProducts(createdDeal.id, createdDeal.products)
    resetLeadRows(leadId)
    await completeLeadTasks(leadId)
    await moveLeadToColumn(leadId, 'deal')
    closeLeadDetails()
    await router.push({ name: 'deals', query: { dealId: createdDeal.id } })
  } catch (error) {
    console.error('Не удалось создать сделку из лида', error)
  }
}

onMounted(() => {
  void Promise.all([loadLeads(true), loadDeals(), loadTasks()]).then(() => openLeadFromRouteQuery())
})

type LeadCreatedSSE = { lead: any }

let leadEventsAbortController: AbortController | null = null

function applyLeadCreatedEvent(rawLead: any) {
  const normalized = normalizeLead(rawLead)
  if (leads.value.some((item) => item.id === normalized.id)) return
  leads.value = [normalized, ...leads.value]
  if (normalized.columnId === 'new') {
    void playNewLeadSound()
  }
}

async function startLeadEventsStream() {
  if (leadEventsAbortController) return
  const token = getAuthToken()
  if (!token) return

  const controller = new AbortController()
  leadEventsAbortController = controller

  try {
    const response = await fetch(`${getApiBaseUrl()}/api/v1/events/leads`, {
      method: 'GET',
      headers: { Authorization: `Bearer ${token}` },
      signal: controller.signal,
    })
    if (!response.ok || !response.body) {
      leadEventsAbortController = null
      return
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let currentEvent = ''
    let currentData = ''

    while (true) {
      const { value, done } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })

      while (true) {
        const idx = buffer.indexOf('\n')
        if (idx === -1) break
        const line = buffer.slice(0, idx).replace(/\r$/, '')
        buffer = buffer.slice(idx + 1)

        if (line === '') {
          if (currentEvent === 'lead-created' && currentData.trim() !== '') {
            try {
              const parsed = JSON.parse(currentData) as LeadCreatedSSE
              if (parsed?.lead) applyLeadCreatedEvent(parsed.lead)
            } catch (error) {
              console.warn('Invalid lead SSE payload', error)
            }
          }
          currentEvent = ''
          currentData = ''
          continue
        }

        if (line.startsWith('event:')) {
          currentEvent = line.slice('event:'.length).trim()
          continue
        }
        if (line.startsWith('data:')) {
          const chunk = line.slice('data:'.length).trim()
          currentData = currentData ? `${currentData}\n${chunk}` : chunk
        }
      }
    }
  } catch {
    // ignore network errors (user may navigate away)
  } finally {
    leadEventsAbortController = null
  }
}

onMounted(() => {
  void startLeadEventsStream()

  const unlock = () => {
    void unlockNewLeadSound().then(() => {
      if (isNewLeadSoundUnlocked()) {
        window.removeEventListener('pointerdown', unlock)
        window.removeEventListener('keydown', unlock)
      }
    })
  }
  window.addEventListener('pointerdown', unlock)
  window.addEventListener('keydown', unlock)
})

onBeforeUnmount(() => {
  if (leadEventsAbortController) {
    leadEventsAbortController.abort()
    leadEventsAbortController = null
  }
})

watch(
  () => route.query.leadId,
  () => {
    void openLeadFromRouteQuery()
  },
)

watch(
  () => selectedLead.value?.id,
  () => {
    syncLeadPickupDeliverySnapshot()
    syncLeadProductionSnapshot()
    if (selectedLead.value && !activeLeadDeal.value) {
      hydrateLeadRows(selectedLead.value.id)
    }
  },
)

watch(
  () => activeLeadDeal.value?.id,
  (dealId) => {
    if (!dealId) return
    hydrateDealRows(dealId)
  },
)

watch(
  () => activeLeadDeal.value?.products,
  () => {
    if (!activeLeadDeal.value) return
    hydrateDealRows(activeLeadDeal.value.id)
  },
  { deep: true },
)

watch(
  () => selectedLead.value?.pickupDelivery,
  () => {
    if (isLeadPickupFieldsDisabled.value && isLeadDeliveryFieldsDisabled.value) return
    schedulePersistLeadPickupDelivery()
  },
  { deep: true },
)

watch(
  () => selectedLead.value?.production,
  () => {
    if (isLeadLinkedToActiveDeal.value) return
    schedulePersistLeadProduction()
  },
  { deep: true },
)

function serializeLeadPickupDelivery(value: PickupDelivery) {
  return JSON.stringify(value)
}

function syncLeadPickupDeliverySnapshot() {
  if (!selectedLead.value) {
    persistedLeadPickupDeliveryJSON.value = ''
    return
  }
  persistedLeadPickupDeliveryJSON.value = serializeLeadPickupDelivery(selectedLead.value.pickupDelivery)
}

function serializeLeadProduction(value: LeadProduction) {
  return JSON.stringify(value)
}

function syncLeadProductionSnapshot() {
  if (!selectedLead.value) {
    persistedLeadProductionJSON.value = ''
    return
  }
  persistedLeadProductionJSON.value = serializeLeadProduction(selectedLead.value.production)
}

function schedulePersistLeadProduction() {
  if (!selectedLead.value || isLeadLinkedToActiveDeal.value) return
  const current = serializeLeadProduction(selectedLead.value.production)
  if (current === persistedLeadProductionJSON.value) return

  // Для лида (до создания сделки) сохраняем сразу, чтобы refresh страницы не терял изменения.
  if (!activeLeadDeal.value) {
    void persistLeadProduction()
    return
  }

  if (leadProductionSaveTimer) {
    clearTimeout(leadProductionSaveTimer)
  }
  leadProductionSaveTimer = setTimeout(() => {
    leadProductionSaveTimer = null
    void persistLeadProduction()
  }, 500)
}

async function flushPersistLeadProduction() {
  if (leadProductionSaveTimer) {
    clearTimeout(leadProductionSaveTimer)
    leadProductionSaveTimer = null
  }
  await persistLeadProduction()
}

async function persistLeadProduction() {
  if (!selectedLead.value || isLeadLinkedToActiveDeal.value) return

  const leadId = selectedLead.value.id
  const production = { ...selectedLead.value.production }
  const current = serializeLeadProduction(production)
  if (current === persistedLeadProductionJSON.value) return

  try {
    await updateLeadProduction(leadId, production)
    persistedLeadProductionJSON.value = current
  } catch (error) {
    console.error('Не удалось сохранить производство лида', error)
  }
}

function schedulePersistLeadPickupDelivery() {
  if (!selectedLead.value || isLeadLinkedToActiveDeal.value) return
  const current = serializeLeadPickupDelivery(selectedLead.value.pickupDelivery)
  if (current === persistedLeadPickupDeliveryJSON.value) return

  if (leadPickupDeliverySaveTimer) {
    clearTimeout(leadPickupDeliverySaveTimer)
  }
  leadPickupDeliverySaveTimer = setTimeout(() => {
    leadPickupDeliverySaveTimer = null
    void persistLeadPickupDelivery()
  }, 500)
}

async function flushPersistLeadPickupDelivery() {
  if (leadPickupDeliverySaveTimer) {
    clearTimeout(leadPickupDeliverySaveTimer)
    leadPickupDeliverySaveTimer = null
  }
  await persistLeadPickupDelivery()
}

async function persistLeadPickupDelivery() {
  if (!selectedLead.value || isLeadLinkedToActiveDeal.value) return

  const leadId = selectedLead.value.id
  const pickupDelivery = { ...selectedLead.value.pickupDelivery }
  const current = serializeLeadPickupDelivery(pickupDelivery)
  if (current === persistedLeadPickupDeliveryJSON.value) return

  try {
    await updateLeadPickupDelivery(leadId, pickupDelivery)
    persistedLeadPickupDeliveryJSON.value = current
  } catch (error) {
    console.error('Не удалось сохранить самовывоз/доставку лида', error)
  }
}

function syncLeadCommentDraft(lead: Lead) {
  leadCommentDrafts[lead.id] = lead.leadComments ?? ''
}

function syncLeadProfileDraft(lead: Lead) {
  leadProfileDrafts[lead.id] = {
    firstName: lead.firstName ?? '',
    patronymic: lead.patronymic ?? '',
  }
}

function syncLinkedDealsProfile(leadId: string, firstName: string, patronymic: string) {
  deals.value = deals.value.map((deal) =>
    deal.leadId === leadId ? { ...deal, firstName, patronymic } : deal,
  )
}

async function persistLeadProfile() {
  if (!selectedLead.value) return

  const draft = leadProfileDrafts[selectedLead.value.id] ?? {
    firstName: selectedLead.value.firstName ?? '',
    patronymic: selectedLead.value.patronymic ?? '',
  }
  const firstName = draft.firstName.trim()
  const patronymic = draft.patronymic.trim()

  if (!firstName) {
    syncLeadProfileDraft(selectedLead.value)
    return
  }

  if (
    firstName === (selectedLead.value.firstName ?? '').trim() &&
    patronymic === (selectedLead.value.patronymic ?? '').trim()
  ) {
    return
  }

  try {
    const updated = await updateLeadProfile(selectedLead.value.id, firstName, patronymic)
    if (updated) {
      syncLeadProfileDraft(updated)
      syncLinkedDealsProfile(updated.id, updated.firstName, updated.patronymic)
    }
  } catch (error) {
    console.error('Не удалось сохранить имя/отчество лида', error)
    syncLeadProfileDraft(selectedLead.value)
  }
}

async function persistLeadComment() {
  if (!selectedLead.value) return

  const draft = leadCommentDrafts[selectedLead.value.id] ?? ''
  if (draft === selectedLead.value.leadComments) {
    return
  }

  selectedLead.value.leadComments = draft
  try {
    await updateLeadComment(selectedLead.value.id, draft)
  } catch (error) {
    console.error('Не удалось сохранить комментарий лида', error)
  }
}

function triggerLeadAttachmentPick() {
  ;(document.getElementById('lead-attachment-input') as HTMLInputElement | null)?.click()
}

function handleLeadAttachmentChange(event: Event) {
  if (!selectedLead.value) return
  const target = event.target as HTMLInputElement | null
  const files = target?.files ? Array.from(target.files) : []
  if (target) target.value = ''
  if (files.length === 0) return

  void uploadLeadFiles(selectedLead.value.id, files)
}

async function uploadLeadFiles(leadId: string, files: File[]) {
  try {
    await addLeadAttachments(leadId, files)
  } catch (error) {
    console.error('Не удалось загрузить вложения лида', error)
  }
}

async function removeLeadAttachmentFile(attachmentId: string) {
  if (!selectedLead.value) return

  try {
    await removeLeadAttachment(selectedLead.value.id, attachmentId)
  } catch (error) {
    console.error('Не удалось удалить файл лида', error)
  }
}

</script>

<template>
  <div
    class="leads-kanban"
    :style="{ height: `${kanbanHeightPx}px`, maxHeight: `${kanbanHeightPx}px` }"
  >
    <div
      class="leads-kanban__viewport"
      :style="{ height: `${kanbanHeightPx}px`, maxHeight: `${kanbanHeightPx}px` }"
    >
      <div class="leads-kanban__track">
        <LeadsKanbanColumn
          v-for="column in LEAD_KANBAN_COLUMNS"
          :key="column.id"
          :title="column.title"
          :column-style="column.style"
          :show-add-lead-button="column.showAddLeadButton"
          :leads="getColumnLeads(column.id)"
          :count="getColumnLeads(column.id).length"
          @add-lead="handleAddLead(column.id, $event)"
          @layout-change="updateKanbanHeight"
          @open-lead="handleOpenLead"
        />
      </div>
    </div>

    <Transition name="lead-details-backdrop">
      <button
        v-if="selectedLead"
        type="button"
        class="lead-details__backdrop"
        aria-label="Закрыть карточку лида"
        @click="closeLeadDetails"
      />
    </Transition>

    <Transition name="lead-details-sheet">
      <section v-if="selectedLead" class="lead-details-sheet" @click.stop>
        <header class="lead-details-sheet__header">
          <h2 class="lead-details-sheet__title">Лид #{{ selectedLead.leadNumber }}</h2>
          <div class="lead-details-sheet__header-actions">
            <button
              v-if="hasLinkedDeal"
              type="button"
              class="lead-details-sheet__action"
              @click="handleGoToDealClick"
            >
              Перейти в сделку
            </button>
            <button
              v-else
              type="button"
              class="lead-details-sheet__action"
              @click="handleCreateDealClick"
            >
              Создать сделку
            </button>
            <button
              type="button"
              class="lead-details-sheet__icon-action lead-details-sheet__icon-action--danger"
              aria-label="Удалить лид"
              @click="handleDeleteLeadClick"
            >
              <NIcon :size="16">
                <TrashOutline />
              </NIcon>
            </button>
            <button
              type="button"
              class="lead-details-sheet__close-btn"
              aria-label="Закрыть карточку лида"
              @click="closeLeadDetails"
            >
              <span aria-hidden="true">×</span>
            </button>
          </div>
        </header>

        <nav class="lead-details-sheet__statuses" aria-label="Статусы лида">
          <button
            v-for="column in LEAD_KANBAN_COLUMNS"
            :key="column.id"
            type="button"
            class="lead-details-sheet__status-btn"
            :class="getStatusButtonClass(column)"
            :style="getStatusButtonStyle(column)"
            @click="moveSelectedLead(column.id)"
          >
            {{ column.title }}
          </button>
        </nav>

        <div class="lead-details-sheet__body">
          <div class="lead-details-sheet__left">
            <aside class="lead-details-sheet__sections">
              <button
                v-for="section in visibleLeadDetailsSections"
                :key="section.id"
                type="button"
                class="lead-details-sheet__section-btn"
                :class="{
                  'lead-details-sheet__section-btn--active': activeDetailsSection === section.id,
                }"
                @click="activeDetailsSection = section.id"
              >
                {{ section.title }}
              </button>
            </aside>

            <section
              class="lead-details-sheet__content"
              :class="{ 'lead-details-sheet__content--chat': activeDetailsSection === 'chat' }"
            >
              <div v-if="activeDetailsSection === 'lead-info'" class="lead-details-sheet__panel">
              <h3 class="lead-details-sheet__panel-title">Информация о лиде</h3>

              <dl class="lead-details-sheet__info-list">
                <div class="lead-details-sheet__info-row">
                  <dt>Имя</dt>
                  <dd>
                    <input
                      v-model="currentLeadFirstName"
                      type="text"
                      class="lead-details-sheet__input lead-details-sheet__input--inline"
                      placeholder="Укажите имя"
                      @blur="persistLeadProfile"
                    />
                  </dd>
                </div>
                <div class="lead-details-sheet__info-row">
                  <dt>Отчество</dt>
                  <dd>
                    <input
                      v-model="currentLeadPatronymic"
                      type="text"
                      class="lead-details-sheet__input lead-details-sheet__input--inline"
                      placeholder="Укажите отчество"
                      @blur="persistLeadProfile"
                    />
                  </dd>
                </div>
                <div class="lead-details-sheet__info-row">
                  <dt>Телефон</dt>
                  <dd>
                    <span class="lead-details-sheet__value">{{ selectedLead.phone || '—' }}</span>
                  </dd>
                </div>
                <div class="lead-details-sheet__info-row">
                  <dt>Лид</dt>
                  <dd>
                    <span class="lead-details-sheet__value">#{{ selectedLead.leadNumber }}</span>
                  </dd>
                </div>
                <div class="lead-details-sheet__info-row">
                  <dt>Источник</dt>
                  <dd>
                    <span class="lead-details-sheet__value">{{ selectedLead.trafficSource || '—' }}</span>
                  </dd>
                </div>
                <div class="lead-details-sheet__info-row">
                  <dt>Статус</dt>
                  <dd>
                    <span class="lead-details-sheet__value">{{ getColumnTitleById(selectedLead.columnId) }}</span>
                  </dd>
                </div>
                <div class="lead-details-sheet__info-row">
                  <dt>Дата создания</dt>
                  <dd>
                    <span class="lead-details-sheet__value">{{ formatLeadCreatedAt(selectedLead.createdAt) }}</span>
                  </dd>
                </div>
                <div
                  v-if="selectedLead.columnId === 'failed' && selectedLead.failureReason"
                  class="lead-details-sheet__info-row"
                >
                  <dt>Причина провала</dt>
                  <dd>
                    <span class="lead-details-sheet__value">{{ selectedLead.failureReason }}</span>
                  </dd>
                </div>
              </dl>
              </div>

            <div
              v-else-if="selectedLead && activeDetailsSection === 'chat' && isAvitoChatLead(selectedLead)"
              class="lead-details-sheet__panel lead-details-sheet__panel--chat"
            >
              <LeadAvitoChatPanel :key="selectedLead.id" :lead-id="selectedLead.id" />
            </div>

            <div v-else-if="activeDetailsSection === 'task'" class="lead-details-sheet__panel">
              <h3 class="lead-details-sheet__panel-title">Задача</h3>
              <p v-if="isLeadLinkedToActiveDeal" class="lead-details-sheet__lock-note">
                Редактирование отключено: по лиду уже создана активная сделка.
              </p>

              <button
                type="button"
                class="lead-details-sheet__task-create-btn"
                :disabled="isLeadLinkedToActiveDeal"
                :title="
                  isLeadLinkedToActiveDeal
                    ? 'Редактирование отключено для лида с активной сделкой'
                    : ''
                "
                @click="openCreateTaskModal"
              >
                Создать задачу
              </button>

              <section class="lead-details-sheet__tasks-list-wrapper">
                <h4 class="lead-details-sheet__sub-title">Задачи по лиду</h4>

                <ul v-if="leadTasks.length > 0" class="lead-details-sheet__tasks-list">
                  <li
                    v-for="task in leadTasks"
                    :key="task.id"
                    class="lead-details-sheet__task-item"
                    :class="{ 'lead-details-sheet__task-item--clickable': task.status === 'active' }"
                    @click="openLeadTaskDetails(task)"
                  >
                    <div class="lead-details-sheet__task-item-header">
                      <p class="lead-details-sheet__task-item-title">{{ task.title }}</p>
                      <span
                        class="lead-details-sheet__task-item-status"
                        :class="{
                          'lead-details-sheet__task-item-status--completed': task.status === 'completed',
                        }"
                      >
                        {{ getLeadTaskStatusLabel(task.status) }}
                      </span>
                    </div>
                    <p class="lead-details-sheet__task-item-text">{{ task.text }}</p>
                    <p class="lead-details-sheet__task-item-meta">
                      Срок: {{ formatDateTime(task.dueAt) || 'Не указан' }}
                    </p>
                  </li>
                </ul>
                <p v-else class="lead-details-sheet__task-empty">По этому лиду пока нет задач</p>
              </section>
            </div>

            <div v-else-if="activeDetailsSection === 'products'" class="lead-details-sheet__panel">
              <h3 class="lead-details-sheet__panel-title">Услуги/Товары</h3>
              <DealProductsEditor
                v-model="leadProductRows"
                @persist="schedulePersistLeadProducts"
              />
            </div>

            <div v-else-if="activeDetailsSection === 'pickup'" class="lead-details-sheet__panel">
              <h3 class="lead-details-sheet__panel-title">Самовывоз</h3>
              <p v-if="isLeadLinkedToActiveDeal" class="lead-details-sheet__lock-note">
                Редактирование отключено: данные перенесены в активную сделку.
              </p>
              <p v-else-if="isLeadPickupSectionLocked" class="lead-details-sheet__lock-note">
                {{ PICKUP_SECTION_LOCKED_MESSAGE }}
              </p>

              <label class="lead-details-sheet__field">
                <span class="lead-details-sheet__label">Адрес самовывоза</span>
                <input
                  v-model="currentLeadPickupAddress"
                  type="text"
                  class="lead-details-sheet__input"
                  placeholder="Укажите адрес самовывоза"
                  :disabled="isLeadPickupFieldsDisabled"
                  @blur="flushPersistLeadPickupDelivery"
                />
              </label>

              <label class="lead-details-sheet__field">
                <span class="lead-details-sheet__label">Дата и время самовывоза</span>
                <DateTimeField
                  :display-value="formatDateTime(currentLeadPickupDate) || 'Выберите дату и время'"
                  :has-value="currentLeadPickupDate !== null"
                  :disabled="isLeadPickupFieldsDisabled"
                  @open="openPickupDateModal"
                  @clear="clearLeadPickupDate"
                />
              </label>
            </div>

            <div v-else-if="activeDetailsSection === 'delivery'" class="lead-details-sheet__panel">
              <h3 class="lead-details-sheet__panel-title">Доставка</h3>
              <p v-if="isLeadLinkedToActiveDeal" class="lead-details-sheet__lock-note">
                Редактирование отключено: данные перенесены в активную сделку.
              </p>
              <p v-else-if="isLeadDeliverySectionLocked" class="lead-details-sheet__lock-note">
                {{ DELIVERY_SECTION_LOCKED_MESSAGE }}
              </p>

              <label class="lead-details-sheet__field">
                <span class="lead-details-sheet__label">Адрес доставки</span>
                <input
                  v-model="currentLeadDeliveryAddress"
                  type="text"
                  class="lead-details-sheet__input"
                  placeholder="Укажите адрес доставки"
                  :disabled="isLeadDeliveryFieldsDisabled"
                  @blur="flushPersistLeadPickupDelivery"
                />
              </label>

              <label class="lead-details-sheet__field">
                <span class="lead-details-sheet__label">Дата и время доставки</span>
                <DateTimeField
                  :display-value="formatDateTime(currentLeadDeliveryDate) || 'Выберите дату и время'"
                  :has-value="currentLeadDeliveryDate !== null"
                  :disabled="isLeadDeliveryFieldsDisabled"
                  @open="openDeliveryDateModal"
                  @clear="clearLeadDeliveryDate"
                />
              </label>

              <label class="lead-details-sheet__field">
                <span class="lead-details-sheet__label">Курьер</span>
                <input
                  v-model="currentLeadCourier"
                  type="text"
                  class="lead-details-sheet__input"
                  placeholder="Укажите курьера"
                  :disabled="isLeadDeliveryFieldsDisabled"
                  @blur="flushPersistLeadPickupDelivery"
                />
              </label>
            </div>

            <div v-else-if="activeDetailsSection === 'production'" class="lead-details-sheet__panel">
              <h3 class="lead-details-sheet__panel-title">Производство</h3>
              <p v-if="isLeadLinkedToActiveDeal" class="lead-details-sheet__lock-note">
                Редактирование отключено: данные перенесены в активную сделку.
              </p>

              <label class="lead-details-sheet__field">
                <span class="lead-details-sheet__label">Номенклатура</span>
                <NSelect
                  v-model:value="currentLeadNomenclature"
                  :options="productionNomenclatureOptions"
                  class="lead-details-sheet__select"
                  placeholder="Выберите номенклатуру"
                  :disabled="isLeadLinkedToActiveDeal"
                />
              </label>

              <label class="lead-details-sheet__field">
                <span class="lead-details-sheet__label">Дата и время начала производства</span>
                <DateTimeField
                  :display-value="formatDateTime(currentLeadProductionDueAt) || 'Выберите дату и время'"
                  :has-value="currentLeadProductionDueAt !== null"
                  :disabled="isLeadLinkedToActiveDeal"
                  @open="openProductionDateModal"
                  @clear="clearLeadProductionDueAt"
                />
              </label>

              <label class="lead-details-sheet__field">
                <span class="lead-details-sheet__label">Сотрудник</span>
                <NSelect
                  v-model:value="currentLeadEmployee"
                  :options="productionEmployeeOptions"
                  class="lead-details-sheet__select"
                  placeholder="Выберите сотрудника"
                  :disabled="isLeadLinkedToActiveDeal"
                />
              </label>
            </div>
            </section>
          </div>

          <aside class="lead-details-sheet__communication">
            <section class="lead-details-sheet__side-block">
              <h3 class="lead-details-sheet__side-title">Комментарий</h3>
              <div class="lead-details-sheet__comment-box">
                <textarea
                  v-model="currentLeadComment"
                  class="lead-details-sheet__comment-area"
                  rows="4"
                  placeholder="Комментарий менеджера по разговору с клиентом"
                  @blur="persistLeadComment"
                />

                <div class="lead-details-sheet__comment-box-foot">
                  <input id="lead-attachment-input" type="file" class="lead-details-sheet__file-input" multiple @change="handleLeadAttachmentChange" />
                  <button type="button" class="lead-details-sheet__attachment-btn" @click="triggerLeadAttachmentPick">
                    Прикрепить файл
                  </button>

                  <EntityAttachmentList
                    :attachments="selectedLead.attachments"
                    @remove="removeLeadAttachmentFile"
                  />
                </div>
              </div>
            </section>

            <section class="lead-details-sheet__side-block lead-details-sheet__side-block--timeline">
              <h3 class="lead-details-sheet__side-title">Таймлайн</h3>

              <ul v-if="sortedLeadActivities.length > 0" class="lead-details-sheet__timeline">
                <li
                  v-for="entry in sortedLeadActivities"
                  :key="entry.id"
                  class="lead-details-sheet__timeline-entry"
                  :class="{
                    'lead-details-sheet__timeline-entry--comment': entry.type === 'comment',
                    'lead-details-sheet__timeline-entry--system': entry.type === 'system',
                  }"
                >
                  <div class="lead-details-sheet__timeline-entry-body">
                    <p class="lead-details-sheet__timeline-text">{{ entry.text }}</p>
                    <p class="lead-details-sheet__timeline-meta">
                      {{ formatDateTime(entry.createdAt) }}
                    </p>
                  </div>
                </li>
              </ul>
              <p v-else class="lead-details-sheet__timeline-empty">Таймлайн пока пуст</p>
            </section>
          </aside>
        </div>
      </section>
    </Transition>

    <AppModal
      v-model:show="isCreateTaskModalOpen"
      title="Новая задача по лиду"
      width="wide"
      actions-align="end"
      close-label="Закрыть окно создания задачи"
      @close="closeCreateTaskModal"
    >
      <div class="lead-details-sheet__modal-fields">
        <label class="lead-details-sheet__field">
          <span class="lead-details-sheet__label">Заголовок</span>
          <input
            v-model="taskForm.title"
            type="text"
            class="lead-details-sheet__input"
            placeholder="Введите заголовок задачи"
          />
        </label>

        <label class="lead-details-sheet__field">
          <span class="lead-details-sheet__label">Текст задачи</span>
          <textarea
            v-model="taskForm.text"
            class="lead-details-sheet__textarea"
            rows="5"
            placeholder="Введите текст задачи"
          />
        </label>

        <label class="lead-details-sheet__field">
          <span class="lead-details-sheet__label">Крайний срок</span>
          <button
            type="button"
            class="lead-details-sheet__date-trigger"
            @click="openTaskDateModal"
          >
            {{ formatDateTime(taskForm.dueAt) || 'Выберите дату и время' }}
          </button>
        </label>
      </div>

      <template #actions>
        <AppModalButton
          :disabled="!canCreateTask"
          :title="canCreateTask ? '' : 'Заполните заголовок, текст задачи и крайний срок'"
          @click="handleCreateTaskClick"
        >
          Создать задачу
        </AppModalButton>
      </template>
    </AppModal>

    <AppModal
      v-model:show="isTaskDateModalOpen"
      title="Дата и время задачи"
      width="wide"
      body-variant="date"
      close-label="Закрыть выбор даты и времени"
    >
      <NDatePicker
        v-model:value="currentTaskDueAt"
        class="app-modal__date-panel"
        type="datetime"
        panel
        clearable
        default-time="09:00:00"
        date-format="dd.MM.yyyy"
        time-picker-format="HH:mm"
        format="dd.MM.yyyy HH:mm"
        :time-picker-props="TIME_PICKER_PROPS"
        :actions="['confirm']"
      >
        <template #confirm="{ text, onConfirm }">
          <div class="app-modal__date-confirm">
            <NButton class="app-modal__date-confirm-btn" type="primary" size="small" @click="handleTaskDateConfirm(onConfirm)">
              {{ text }}
            </NButton>
          </div>
        </template>
      </NDatePicker>
    </AppModal>

    <AppModal
      v-model:show="isProductionDateModalOpen"
      title="Дата и время производства"
      width="wide"
      body-variant="date"
      close-label="Закрыть выбор даты и времени"
    >
      <NDatePicker
        v-model:value="currentLeadProductionDueAt"
        class="app-modal__date-panel"
        type="datetime"
        panel
        clearable
        default-time="09:00:00"
        date-format="dd.MM.yyyy"
        time-picker-format="HH:mm"
        format="dd.MM.yyyy HH:mm"
        :time-picker-props="TIME_PICKER_PROPS"
        :actions="['confirm']"
      >
        <template #confirm="{ text, onConfirm }">
          <div class="app-modal__date-confirm">
            <NButton class="app-modal__date-confirm-btn" type="primary" size="small" @click="handleProductionDateConfirm(onConfirm)">
              {{ text }}
            </NButton>
          </div>
        </template>
      </NDatePicker>
    </AppModal>

    <AppModal
      v-model:show="isPickupDateModalOpen"
      title="Дата и время самовывоза"
      width="wide"
      body-variant="date"
      close-label="Закрыть выбор даты и времени"
    >
      <NDatePicker
        v-model:value="currentLeadPickupDate"
        class="app-modal__date-panel"
        type="datetime"
        panel
        clearable
        default-time="09:00:00"
        date-format="dd.MM.yyyy"
        time-picker-format="HH:mm"
        format="dd.MM.yyyy HH:mm"
        :time-picker-props="TIME_PICKER_PROPS"
        :actions="['confirm']"
      >
        <template #confirm="{ text, onConfirm }">
          <div class="app-modal__date-confirm">
            <NButton class="app-modal__date-confirm-btn" type="primary" size="small" @click="handlePickupDateConfirm(onConfirm)">
              {{ text }}
            </NButton>
          </div>
        </template>
      </NDatePicker>
    </AppModal>

    <AppModal
      v-model:show="isDeliveryDateModalOpen"
      title="Дата и время доставки"
      width="wide"
      body-variant="date"
      close-label="Закрыть выбор даты и времени"
    >
      <NDatePicker
        v-model:value="currentLeadDeliveryDate"
        class="app-modal__date-panel"
        type="datetime"
        panel
        clearable
        default-time="09:00:00"
        date-format="dd.MM.yyyy"
        time-picker-format="HH:mm"
        format="dd.MM.yyyy HH:mm"
        :time-picker-props="TIME_PICKER_PROPS"
        :actions="['confirm']"
      >
        <template #confirm="{ text, onConfirm }">
          <div class="app-modal__date-confirm">
            <NButton class="app-modal__date-confirm-btn" type="primary" size="small" @click="handleDeliveryDateConfirm(onConfirm)">
              {{ text }}
            </NButton>
          </div>
        </template>
      </NDatePicker>
    </AppModal>

    <AppModal
      v-model:show="isFailureReasonModalOpen"
      title="Причина провала"
      @close="closeFailureReasonModal"
    >
      <textarea
        v-model="failureReasonDraft"
        class="app-modal__textarea"
        rows="5"
        placeholder="Укажите пожалуйста причину провала лида"
      />

      <template #actions>
        <AppModalButton :disabled="!canConfirmFailureReason" @click="confirmFailureReason">
          Подтвердить
        </AppModalButton>
      </template>
    </AppModal>

    <TaskDetailsSheet :task-id="selectedLeadTask?.id ?? null" :task="selectedLeadTask" @close="closeLeadTaskDetails" />
  </div>
</template>

<style scoped>
.leads-kanban {
  position: relative;
  overflow: hidden;
  flex-shrink: 0;
}

.leads-kanban__viewport {
  display: flex;
  flex-direction: column;
  overflow-x: auto;
  overflow-y: hidden;
  padding-top: 16px;
  box-sizing: border-box;
}

.leads-kanban__viewport::-webkit-scrollbar {
  height: 8px;
}

.leads-kanban__viewport::-webkit-scrollbar-track {
  background: transparent;
}

.leads-kanban__viewport::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 4px;
}

.leads-kanban__track {
  display: flex;
  gap: 10px;
  flex: 1;
  min-height: 0;
  width: max-content;
  margin: 0 auto;
  padding: 0 24px;
  box-sizing: border-box;
  align-items: stretch;
}

.lead-details__backdrop {
  position: fixed;
  inset: 0;
  border: 0;
  background: rgba(15, 23, 42, 0.2);
  z-index: 180;
  cursor: default;
}

.lead-details-sheet {
  position: fixed;
  top: 15px;
  right: 15px;
  bottom: 0;
  left: 15px;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border-radius: 12px 12px 0 0;
  box-shadow: 0 -8px 24px rgba(15, 23, 42, 0.15);
  z-index: 190;
  overflow: hidden;
}

.lead-details-sheet__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid #e2e8f0;
}

.lead-details-sheet__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.lead-details-sheet__action {
  padding: 8px 14px;
  border: 1px solid #1f883d;
  border-radius: 8px;
  background: #1f883d;
  color: #ffffff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.lead-details-sheet__action:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.lead-details-sheet__header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.lead-details-sheet__icon-action {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #64748b;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.lead-details-sheet__icon-action:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.lead-details-sheet__icon-action--danger:hover {
  color: #dc2626;
}

.lead-details-sheet__close-btn {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #475569;
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.lead-details-sheet__close-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1f2937;
}

.lead-details-sheet__statuses {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid #e2e8f0;
}

.lead-details-sheet__status-btn {
  width: 100%;
  min-width: 0;
  padding: 8px 10px;
  border: 1px solid #dbe3ed;
  border-radius: 8px;
  background: #f8fafc;
  color: #4a5568;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.3;
  text-align: center;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease,
    box-shadow 0.15s ease;
}

.lead-details-sheet__status-btn {
  box-shadow: inset 0 -2px 0 0 var(--status-color);
}

.lead-details-sheet__status-btn--active {
  background: var(--status-bg);
  border-color: var(--status-border);
  color: var(--status-color);
  box-shadow:
    inset 0 0 0 1px var(--status-color),
    inset 0 -2px 0 0 var(--status-color);
}

.lead-details-sheet__body {
  flex: 1 1 auto;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 16px;
  padding: 16px;
}

.lead-details-sheet__left {
  min-width: 0;
  min-height: 0;
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
}

.lead-details-sheet__sections {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.lead-details-sheet__section-btn {
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: #334155;
  font-size: 14px;
  text-align: left;
  padding: 9px 10px;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.lead-details-sheet__section-btn--active {
  background: #ffffff;
  border-color: #d1d9e2;
  color: #0f172a;
  font-weight: 600;
}

.lead-details-sheet__content {
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 16px;
}

.lead-details-sheet__content--chat {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 10px;
}

.lead-details-sheet__panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.lead-details-sheet__panel--chat {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  gap: 0;
  overflow: hidden;
}

.lead-details-sheet__panel--chat :deep(.lead-avito-chat) {
  flex: 1 1 auto;
  min-height: 0;
}

.lead-details-sheet__panel-title {
  margin: 0;
  font-size: 18px;
  color: #0f172a;
}

.lead-details-sheet__info-list {
  margin: 0;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 16px;
  row-gap: 12px;
  align-items: start;
}

.lead-details-sheet__info-row {
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.lead-details-sheet__info-row dt {
  font-size: 12px;
  color: #64748b;
}

.lead-details-sheet__info-row dd {
  margin: 0;
  min-width: 0;
}

.lead-details-sheet__value,
.lead-details-sheet__input--inline {
  display: block;
  width: 100%;
  box-sizing: border-box;
  min-height: 36px;
  padding: 8px 10px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  line-height: 1.25;
  color: #0f172a;
}

.lead-details-sheet__value {
  border: 1px solid transparent;
  background: #f8fafc;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.lead-details-sheet__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.lead-details-sheet__label {
  font-size: 13px;
  color: #475569;
}

.lead-details-sheet__input,
.lead-details-sheet__textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #0f172a;
  padding: 8px 10px;
  font-size: 14px;
  font-family: inherit;
}

.lead-details-sheet__input--inline {
  margin: 0;
  border: 1px solid #cbd5e1;
  background: #ffffff;
  font-family: inherit;
}

.lead-details-sheet__textarea {
  resize: vertical;
  min-height: 110px;
}

.lead-details-sheet__input:focus,
.lead-details-sheet__textarea:focus {
  outline: none;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.lead-details-sheet__date-picker {
  width: 100%;
}

.lead-details-sheet__date-trigger {
  width: 100%;
  min-height: 34px;
  appearance: none;
  box-sizing: border-box;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #0f172a;
  padding: 8px 10px;
  font-size: 14px;
  font-family: inherit;
  line-height: 1.3;
  text-align: left;
  text-decoration: none;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: pointer;
}

.lead-details-sheet__date-trigger:hover {
  border-color: #93c5fd;
}

.lead-details-sheet__date-trigger:disabled {
  background: #f8fafc;
  color: #94a3b8;
  cursor: not-allowed;
}

.lead-details-sheet__modal-fields {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.lead-details-sheet__select {
  width: 100%;
}

.lead-details-sheet__date-picker :deep(.n-input-wrapper) {
  border-radius: 8px;
}

:deep(.n-date-panel) {
  max-width: calc(100vw - 32px);
  max-height: calc(100vh - 24px);
  overflow-y: auto;
  border-radius: 14px;
  border: 1px solid #d9e5f2;
  box-shadow: 0 18px 36px rgba(15, 23, 42, 0.16);
  box-sizing: border-box;
}

:deep(.n-date-panel-calendar) {
  padding: 10px 12px;
}

:deep(.n-date-panel-month) {
  padding: 6px 0 10px;
}

:deep(.n-date-panel-weekdays) {
  margin-bottom: 8px;
}

:deep(.n-date-panel-date) {
  border-radius: 999px;
}

:deep(.n-date-panel-date.n-date-panel-date--selected::after) {
  background-color: #4a5568 !important;
}

:deep(.n-date-panel-date.n-date-panel-date--selected) {
  color: #ffffff !important;
}

:deep(.n-date-panel-date.n-date-panel-date--current .n-date-panel-date__sup) {
  background-color: #4a5568;
}

:deep(.n-date-panel-actions__suffix) {
  width: 100%;
}

.lead-details-sheet__task-create-btn {
  align-self: flex-start;
  border: 1px solid #1f883d;
  border-radius: 8px;
  background: #1f883d;
  color: #ffffff;
  font-size: 13px;
  font-weight: 600;
  padding: 8px 14px;
  cursor: pointer;
}

.lead-details-sheet__task-create-btn:disabled {
  border-color: #2aa04c;
  background: #2aa04c;
  color: #f8fff9;
  cursor: not-allowed;
}

.lead-details-sheet__sub-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
}

.lead-details-sheet__tasks-list-wrapper {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.lead-details-sheet__tasks-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.lead-details-sheet__task-item {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #ffffff;
  padding: 8px 9px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.lead-details-sheet__task-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.lead-details-sheet__task-item-title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #1a202c;
}

.lead-details-sheet__task-item-status {
  border: 1px solid #d1d9e2;
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 12px;
  font-weight: 600;
  color: #1f883d;
  background: #f0fff4;
  white-space: nowrap;
}

.lead-details-sheet__task-item-status--completed {
  color: #4a5568;
  background: #f8fafc;
}

.lead-details-sheet__task-item-text {
  margin: 0;
  font-size: 12px;
  color: #334155;
  white-space: pre-wrap;
  word-break: break-word;
}

.lead-details-sheet__task-item-meta {
  margin: 0;
  font-size: 11px;
  color: #64748b;
}

.lead-details-sheet__task-item--clickable {
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.lead-details-sheet__task-item--clickable:hover {
  border-color: #cbd5e1;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.06);
}

.lead-details-sheet__task-empty {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.lead-details-sheet__lock-note,
.lead-details-sheet__catalog-note {
  margin: 0;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
  color: #475569;
  font-size: 13px;
}

.lead-details-sheet__placeholder {
  margin: 0;
  font-size: 14px;
  color: #4a5568;
}

.lead-details-sheet__communication {
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 0 0 0 20px;
  border-left: 1px solid #e2e8f0;
  scrollbar-gutter: stable;
}

.lead-details-sheet__side-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.lead-details-sheet__side-block--timeline {
  flex: 1 1 auto;
  min-height: 0;
}

.lead-details-sheet__side-title {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  color: #1a202c;
}

.lead-details-sheet__comment-box {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  overflow: hidden;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.lead-details-sheet__comment-box:focus-within {
  border-color: #cbd5e1;
  box-shadow: 0 0 0 3px rgba(31, 136, 61, 0.08);
}

.lead-details-sheet__comment-area {
  width: 100%;
  box-sizing: border-box;
  border: 0;
  background: #ffffff;
  color: #1a202c;
  font: inherit;
  font-size: 14px;
  line-height: 1.45;
  padding: 10px 12px;
  resize: vertical;
  min-height: 88px;
}

.lead-details-sheet__comment-area:focus {
  outline: none;
}

.lead-details-sheet__comment-box-foot {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 12px;
  border-top: 1px solid #e2e8f0;
  background: #f8fafc;
}

.lead-details-sheet__file-input {
  display: none;
}

.lead-details-sheet__attachment-btn {
  align-self: flex-start;
  padding: 8px 12px;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #4a5568;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease,
    color 0.15s ease;
}

.lead-details-sheet__attachment-btn:hover {
  border-color: #cbd5e1;
  background: #f8fafc;
  color: #1a202c;
}

.lead-details-sheet__timeline {
  margin: 0;
  padding: 4px 0 0 14px;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0;
  border-left: 1px solid #e2e8f0;
}

.lead-details-sheet__timeline-entry {
  position: relative;
  padding: 0 0 14px 12px;
}

.lead-details-sheet__timeline-entry::before {
  content: '';
  position: absolute;
  left: -18px;
  top: 6px;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #cbd5e1;
  box-shadow: 0 0 0 3px #ffffff;
}

.lead-details-sheet__timeline-entry--comment::before {
  background: #1f883d;
  box-shadow: 0 0 0 3px rgba(31, 136, 61, 0.14);
}

.lead-details-sheet__timeline-entry--system::before {
  background: #cbd5e1;
}

.lead-details-sheet__timeline-entry-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.lead-details-sheet__timeline-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.45;
  color: #1a202c;
  white-space: pre-wrap;
  word-break: break-word;
}

.lead-details-sheet__timeline-entry--system .lead-details-sheet__timeline-text {
  color: #4a5568;
}

.lead-details-sheet__timeline-meta {
  margin: 0;
  font-size: 12px;
  color: #718096;
}

.lead-details-sheet__timeline-empty {
  margin: 0;
  font-size: 13px;
  color: #718096;
}

@media (max-width: 1200px) {
  .lead-details-sheet__body {
    grid-template-columns: minmax(0, 1fr);
  }

  .lead-details-sheet__left {
    grid-template-columns: minmax(0, 1fr);
  }
}

.lead-details-sheet-enter-active,
.lead-details-sheet-leave-active {
  transition:
    transform 0.28s ease,
    opacity 0.28s ease;
}

.lead-details-sheet-enter-from,
.lead-details-sheet-leave-to {
  transform: translateY(100%);
  opacity: 0.98;
}

.lead-details-backdrop-enter-active,
.lead-details-backdrop-leave-active {
  transition: opacity 0.2s ease;
}

.lead-details-backdrop-enter-from,
.lead-details-backdrop-leave-to {
  opacity: 0;
}
</style>
