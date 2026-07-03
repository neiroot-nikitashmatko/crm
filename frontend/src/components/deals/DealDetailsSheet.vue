<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NDatePicker, NIcon, NSelect } from 'naive-ui'
import { TrashOutline } from '@vicons/ionicons5'
import { DEAL_KANBAN_COLUMNS } from '@/constants/deals'
import { useDeals } from '@/composables/useDeals'
import { useTasks } from '@/composables/useTasks'
import {
  getDealColumnValidationResult,
  resolveDealKanbanColumnId,
} from '@/utils/dealKanban'
import {
  DELIVERY_SECTION_LOCKED_MESSAGE,
  isDeliverySectionLocked,
  isPickupSectionLocked,
  PICKUP_SECTION_LOCKED_MESSAGE,
} from '@/utils/pickupDelivery'
import type { DealKanbanColumnId, PickupDelivery } from '@/types/deal'
import type { Task } from '@/types/task'
import TaskDetailsSheet from '@/components/tasks/TaskDetailsSheet.vue'
import DateTimeField from '@/components/common/DateTimeField.vue'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import EntityAttachmentList from '@/components/attachments/EntityAttachmentList.vue'
import DealProductsEditor from '@/components/common/DealProductsEditor.vue'
import { useDealProductRows } from '@/composables/useDealProductRows'
import type { ProductRow } from '@/types/productRow'

type DealSectionId = 'general' | 'task' | 'products' | 'pickup' | 'delivery' | 'production'

const DEAL_SECTION_TITLES: Record<'production' | 'pickup' | 'delivery', string> = {
  production: 'Производство',
  pickup: 'Самовывоз',
  delivery: 'Доставка',
}

const DEAL_SECTIONS: Array<{ id: DealSectionId; title: string }> = [
  { id: 'general', title: 'Общая информация' },
  { id: 'task', title: 'Задача' },
  { id: 'products', title: 'Услуги/Товары' },
  { id: 'pickup', title: 'Самовывоз' },
  { id: 'delivery', title: 'Доставка' },
  { id: 'production', title: 'Производство' },
]

const WORKING_HOURS = [9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]
const TIME_MINUTE_STEP = 5
const TIME_PICKER_PROPS = {
  format: 'HH:mm',
  hours: WORKING_HOURS,
  minutes: TIME_MINUTE_STEP,
  actions: ['confirm'] as Array<'confirm'>,
}

const props = defineProps<{
  dealId: string | null
}>()

const emit = defineEmits<{
  close: []
}>()

const router = useRouter()
const { deals, deleteDeal, updateDealComment, updateDealPickupDelivery, updateDealProduction, updateDealStatus, addDealAttachments, removeDealAttachment } = useDeals()
const { getDealRows, setDealRows, hydrateDealRows, saveDealProductRows } = useDealProductRows()
const { addTask, getDealTasks, loadTasks } = useTasks()

const activeSection = ref<DealSectionId>('general')
const isProductionDateModalOpen = ref(false)
const isPickupDateModalOpen = ref(false)
const isDeliveryDateModalOpen = ref(false)
const isCreateTaskModalOpen = ref(false)
const isTaskDateModalOpen = ref(false)
const selectedDealTask = ref<Task | null>(null)
const commentDraft = ref('')
const isStatusValidationModalOpen = ref(false)
const statusValidationMessage = ref('')
const statusValidationTargetSection = ref<'production' | 'pickup' | 'delivery' | null>(null)
const isFailureReasonModalOpen = ref(false)
const failureReasonDraft = ref('')
const persistedPickupDeliveryJSON = ref('')
let pickupDeliverySaveTimer: ReturnType<typeof setTimeout> | null = null
let dealProductsSaveTimer: ReturnType<typeof setTimeout> | null = null
let productionSaveTimer: ReturnType<typeof setTimeout> | null = null

const productionNomenclatureOptions = [
  { label: 'Перетяжка руля', value: 'Перетяжка руля' },
  { label: 'Установка чехлов', value: 'Установка чехлов' },
  { label: 'Ремонт стёкол', value: 'Ремонт стёкол' },
  { label: 'Пошив ковриков', value: 'Пошив ковриков' },
]
const productionEmployeeOptions = [
  { label: 'Никита Хачересов', value: 'Никита Хачересов' },
  { label: 'Сергей Геворкян', value: 'Сергей Геворкян' },
]

const canConfirmFailureReason = computed(() => failureReasonDraft.value.trim().length > 0)

const statusValidationSectionTitle = computed(() => {
  if (!statusValidationTargetSection.value) return ''
  return DEAL_SECTION_TITLES[statusValidationTargetSection.value]
})
const taskForm = reactive({
  title: '',
  text: '',
  dueAt: null as number | null,
})

const selectedDeal = computed(() =>
  props.dealId ? deals.value.find((deal) => deal.id === props.dealId) ?? null : null,
)

const dealProductRows = computed<ProductRow[]>({
  get() {
    if (!selectedDeal.value) return []
    return getDealRows(selectedDeal.value.id)
  },
  set(value) {
    if (!selectedDeal.value) return
    setDealRows(selectedDeal.value.id, value)
  },
})

const hasLinkedLead = computed(() => Boolean(selectedDeal.value?.leadId))

const dealTasks = computed<Task[]>(() => {
  if (!selectedDeal.value) {
    return []
  }

  return [...getDealTasks(selectedDeal.value.id)].sort((left, right) => {
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

async function handleGoToLeadClick() {
  const leadId = selectedDeal.value?.leadId
  if (!leadId) {
    return
  }

  emit('close')
  await router.push({ name: 'leads', query: { leadId } })
}

const currentProductionDueAt = computed({
  get: () => selectedDeal.value?.production.dueAt ?? null,
  set: (value: number | null) => {
    if (!selectedDeal.value) return
    selectedDeal.value.production.dueAt = value
    selectedDeal.value.productionDueAt = value
  },
})

const currentProductionNomenclature = computed({
  get: () => selectedDeal.value?.production.nomenclature ?? '',
  set: (value: string | null) => {
    if (!selectedDeal.value) return
    selectedDeal.value.production.nomenclature = value ?? ''
  },
})

const currentProductionEmployee = computed({
  get: () => selectedDeal.value?.production.employee ?? '',
  set: (value: string | null) => {
    if (!selectedDeal.value) return
    selectedDeal.value.production.employee = value ?? ''
  },
})

const currentPickupAddress = computed({
  get: () => selectedDeal.value?.pickupDelivery.pickupAddress ?? '',
  set: (value: string) => {
    if (!selectedDeal.value || isDealPickupFieldsDisabled.value) return
    selectedDeal.value.pickupDelivery.pickupAddress = value
  },
})

const currentPickupDate = computed({
  get: () => selectedDeal.value?.pickupDelivery.pickupDate ?? null,
  set: (value: number | null) => {
    if (!selectedDeal.value || isDealPickupFieldsDisabled.value) return
    selectedDeal.value.pickupDelivery.pickupDate = normalizeDateTimeValue(value)
  },
})

const currentDeliveryAddress = computed({
  get: () => selectedDeal.value?.pickupDelivery.deliveryAddress ?? '',
  set: (value: string) => {
    if (!selectedDeal.value || isDealDeliveryFieldsDisabled.value) return
    selectedDeal.value.pickupDelivery.deliveryAddress = value
  },
})

const currentDeliveryDate = computed({
  get: () => selectedDeal.value?.pickupDelivery.deliveryDate ?? null,
  set: (value: number | null) => {
    if (!selectedDeal.value || isDealDeliveryFieldsDisabled.value) return
    selectedDeal.value.pickupDelivery.deliveryDate = normalizeDateTimeValue(value)
  },
})

const currentCourier = computed({
  get: () => selectedDeal.value?.pickupDelivery.courier ?? '',
  set: (value: string) => {
    if (!selectedDeal.value || isDealDeliveryFieldsDisabled.value) return
    selectedDeal.value.pickupDelivery.courier = value
  },
})

const isDealPickupSectionLocked = computed(() =>
  selectedDeal.value ? isPickupSectionLocked(selectedDeal.value.pickupDelivery) : false,
)
const isDealDeliverySectionLocked = computed(() =>
  selectedDeal.value ? isDeliverySectionLocked(selectedDeal.value.pickupDelivery) : false,
)
const isDealPickupFieldsDisabled = computed(() => isDealPickupSectionLocked.value)
const isDealDeliveryFieldsDisabled = computed(() => isDealDeliverySectionLocked.value)

const sortedActivities = computed(() => {
  if (!selectedDeal.value) return []
  return [...selectedDeal.value.activities].sort((a, b) => b.createdAt - a.createdAt)
})

const resolvedKanbanColumnId = computed(() =>
  selectedDeal.value ? resolveDealKanbanColumnId(selectedDeal.value) : null,
)

watch(
  () => selectedDeal.value?.id,
  () => {
    if (!selectedDeal.value) return
    commentDraft.value = selectedDeal.value.dealComments ?? ''
    isStatusValidationModalOpen.value = false
    statusValidationMessage.value = ''
    statusValidationTargetSection.value = null
    isFailureReasonModalOpen.value = false
    failureReasonDraft.value = ''
    activeSection.value = 'general'
    closeCreateTaskModal()
    closeDealTaskDetails()
    syncPickupDeliverySnapshot()
    if (selectedDeal.value) {
      hydrateDealRows(selectedDeal.value.id)
    }
  },
  { immediate: true },
)

watch(
  () => selectedDeal.value?.pickupDelivery,
  () => {
    schedulePersistDealPickupDelivery()
  },
  { deep: true },
)

watch(
  () => selectedDeal.value?.products,
  () => {
    if (!selectedDeal.value) return
    hydrateDealRows(selectedDeal.value.id)
  },
  { deep: true },
)

watch(
  () => props.dealId,
  (dealId) => {
    if (!dealId) {
      closeCreateTaskModal()
      closeDealTaskDetails()
    }
  },
)

onMounted(() => {
  void loadTasks()
})

function formatDateTime(timestamp: number | null) {
  if (!timestamp) return '—'
  return new Date(timestamp).toLocaleString('ru-RU')
}

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

function isTaskDateFilled(value: unknown) {
  if (value === null || value === undefined) return false
  if (typeof value === 'number') return !Number.isNaN(value)
  if (typeof value === 'string') return value.trim().length > 0
  return true
}

function getDealTaskStatusLabel(status: Task['status']) {
  return status === 'completed' ? 'Завершена' : 'Активна'
}

function openCreateTaskModal() {
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

function openDealTaskDetails(task: Task) {
  selectedDealTask.value = task
}

function closeDealTaskDetails() {
  selectedDealTask.value = null
}

function openTaskDateModal() {
  taskForm.dueAt = normalizeDateTimeValue(taskForm.dueAt)
  isTaskDateModalOpen.value = true
}

function handleTaskDateConfirm(onConfirm: () => void) {
  onConfirm()
  isTaskDateModalOpen.value = false
}

async function handleCreateTaskClick() {
  if (!selectedDeal.value || !canCreateTask.value) {
    return
  }

  try {
    await addTask({
      title: taskForm.title.trim(),
      text: taskForm.text.trim(),
      dueAt: taskForm.dueAt,
      dealId: selectedDeal.value.id,
    })
    closeCreateTaskModal()
  } catch (error) {
    console.error('Не удалось создать задачу по сделке', error)
  }
}

function getStatusButtonStyle(columnId: DealKanbanColumnId) {
  const column = DEAL_KANBAN_COLUMNS.find((item) => item.id === columnId)
  return {
    '--status-color': column?.style.countColor ?? '#4a5568',
    '--status-bg': column?.style.countBg ?? 'rgba(74,85,104,0.12)',
    '--status-border': column?.style.headerBorder ?? '#d1d9e2',
  }
}

async function handleStatusChange(columnId: DealKanbanColumnId) {
  if (!selectedDeal.value) return
  if (resolvedKanbanColumnId.value === columnId) return

  const validationResult = getDealColumnValidationResult(selectedDeal.value, columnId)
  if (validationResult) {
    statusValidationMessage.value = validationResult.message
    statusValidationTargetSection.value = validationResult.targetSection
    isStatusValidationModalOpen.value = true
    return
  }

  isStatusValidationModalOpen.value = false

  if (columnId === 'failed') {
    failureReasonDraft.value = ''
    isFailureReasonModalOpen.value = true
    return
  }

  try {
    await updateDealStatus(selectedDeal.value.id, columnId)
  } catch (error) {
    console.error('Не удалось обновить статус сделки', error)
  }
}

function closeFailureReasonModal() {
  isFailureReasonModalOpen.value = false
  failureReasonDraft.value = ''
}

async function confirmFailureReason() {
  if (!selectedDeal.value || !canConfirmFailureReason.value) return

  const dealId = selectedDeal.value.id
  const reason = failureReasonDraft.value.trim()

  try {
    await updateDealStatus(dealId, 'failed', reason)
    closeFailureReasonModal()
  } catch (error) {
    console.error('Не удалось перевести сделку в проваленные', error)
  }
}

function closeStatusValidationModal() {
  isStatusValidationModalOpen.value = false
}

function handleStatusValidationGoToSection() {
  if (statusValidationTargetSection.value) {
    activeSection.value = statusValidationTargetSection.value
  }
  closeStatusValidationModal()
}

async function handleDeleteDeal() {
  if (!selectedDeal.value) return
  try {
    await deleteDeal(selectedDeal.value.id)
    emit('close')
  } catch (error) {
    console.error('Не удалось удалить сделку', error)
  }
}

async function persistDealComment() {
  if (!selectedDeal.value) return
  if (commentDraft.value === selectedDeal.value.dealComments) return
  selectedDeal.value.dealComments = commentDraft.value
  try {
    await updateDealComment(selectedDeal.value.id, commentDraft.value)
  } catch (error) {
    console.error('Не удалось сохранить комментарий сделки', error)
  }
}

function schedulePersistDealProduction() {
  if (!selectedDeal.value) return
  if (productionSaveTimer) {
    clearTimeout(productionSaveTimer)
  }
  productionSaveTimer = setTimeout(() => {
    productionSaveTimer = null
    void persistDealProduction()
  }, 300)
}

function flushPersistDealProduction() {
  if (productionSaveTimer) {
    clearTimeout(productionSaveTimer)
    productionSaveTimer = null
  }
  void persistDealProduction()
}

async function persistDealProduction() {
  if (!selectedDeal.value) return
  try {
    await updateDealProduction(selectedDeal.value.id, { ...selectedDeal.value.production })
  } catch (error) {
    console.error('Не удалось сохранить производство', error)
  }
}

function serializePickupDelivery(value: PickupDelivery) {
  return JSON.stringify(value)
}

function syncPickupDeliverySnapshot() {
  if (!selectedDeal.value) {
    persistedPickupDeliveryJSON.value = ''
    return
  }
  persistedPickupDeliveryJSON.value = serializePickupDelivery(selectedDeal.value.pickupDelivery)
}

function schedulePersistDealPickupDelivery() {
  if (!selectedDeal.value) return
  const current = serializePickupDelivery(selectedDeal.value.pickupDelivery)
  if (current === persistedPickupDeliveryJSON.value) return

  if (pickupDeliverySaveTimer) {
    clearTimeout(pickupDeliverySaveTimer)
  }
  pickupDeliverySaveTimer = setTimeout(() => {
    pickupDeliverySaveTimer = null
    void persistDealPickupDelivery()
  }, 500)
}

function flushPersistDealPickupDelivery() {
  if (pickupDeliverySaveTimer) {
    clearTimeout(pickupDeliverySaveTimer)
    pickupDeliverySaveTimer = null
  }
  void persistDealPickupDelivery()
}

async function persistDealPickupDelivery() {
  if (!selectedDeal.value) return
  const dealId = selectedDeal.value.id
  const pickupDelivery = { ...selectedDeal.value.pickupDelivery }
  const current = serializePickupDelivery(pickupDelivery)
  if (current === persistedPickupDeliveryJSON.value) return

  try {
    await updateDealPickupDelivery(dealId, pickupDelivery)
    persistedPickupDeliveryJSON.value = current
  } catch (error) {
    console.error('Не удалось сохранить самовывоз/доставку', error)
  }
}

function schedulePersistDealProducts() {
  if (!selectedDeal.value) return

  if (dealProductsSaveTimer) {
    clearTimeout(dealProductsSaveTimer)
  }
  dealProductsSaveTimer = setTimeout(() => {
    dealProductsSaveTimer = null
    void persistDealProducts()
  }, 500)
}

async function persistDealProducts() {
  if (!selectedDeal.value) return

  try {
    await saveDealProductRows(selectedDeal.value.id)
  } catch (error) {
    console.error('Не удалось сохранить товары', error)
  }
}

function openProductionDateModal() {
  if (!selectedDeal.value) return
  currentProductionDueAt.value = currentProductionDueAt.value ?? Date.now()
  isProductionDateModalOpen.value = true
}

async function handleProductionDateConfirm(onConfirm: () => void) {
  if (!selectedDeal.value) return

  const dealId = selectedDeal.value.id
  const previousDueAt = selectedDeal.value.production.dueAt

  onConfirm()
  isProductionDateModalOpen.value = false

  const nextDueAt = currentProductionDueAt.value
  if (previousDueAt === nextDueAt) return

  try {
    await updateDealProduction(dealId, { ...selectedDeal.value.production, dueAt: nextDueAt })
  } catch (error) {
    console.error('Не удалось сохранить дату производства', error)
    currentProductionDueAt.value = previousDueAt
  }
}

function openPickupDateModal() {
  if (!selectedDeal.value || isDealPickupFieldsDisabled.value) return
  currentPickupDate.value = normalizeDateTimeValue(currentPickupDate.value ?? Date.now())
  isPickupDateModalOpen.value = true
}

function handlePickupDateConfirm(onConfirm: () => void) {
  onConfirm()
  isPickupDateModalOpen.value = false
  flushPersistDealPickupDelivery()
}

function openDeliveryDateModal() {
  if (!selectedDeal.value || isDealDeliveryFieldsDisabled.value) return
  currentDeliveryDate.value = normalizeDateTimeValue(currentDeliveryDate.value ?? Date.now())
  isDeliveryDateModalOpen.value = true
}

function handleDeliveryDateConfirm(onConfirm: () => void) {
  onConfirm()
  isDeliveryDateModalOpen.value = false
  flushPersistDealPickupDelivery()
}

function clearPickupDate() {
  currentPickupDate.value = null
  flushPersistDealPickupDelivery()
}

function clearDeliveryDate() {
  currentDeliveryDate.value = null
  flushPersistDealPickupDelivery()
}

async function clearProductionDueAt() {
  if (!selectedDeal.value) return

  const dealId = selectedDeal.value.id
  const previousDueAt = selectedDeal.value.production.dueAt
  if (previousDueAt === null) return

  currentProductionDueAt.value = null

  try {
    await updateDealProduction(dealId, { ...selectedDeal.value.production, dueAt: null })
  } catch (error) {
    console.error('Не удалось очистить дату производства', error)
    currentProductionDueAt.value = previousDueAt
  }
}

function triggerAttachmentPick() {
  ;(document.getElementById('deal-attachment-input') as HTMLInputElement | null)?.click()
}

function handleAttachmentChange(event: Event) {
  if (!selectedDeal.value) return
  const target = event.target as HTMLInputElement | null
  const files = target?.files ? Array.from(target.files) : []
  if (files.length === 0) return

  void uploadAttachments(files)
  if (target) target.value = ''
}

async function uploadAttachments(files: File[]) {
  if (!selectedDeal.value) return

  try {
    await addDealAttachments(selectedDeal.value.id, files)
  } catch (error) {
    console.error('Не удалось загрузить вложения', error)
  }
}

async function removeAttachment(attachmentId: string) {
  if (!selectedDeal.value) return

  try {
    await removeDealAttachment(selectedDeal.value.id, attachmentId)
  } catch (error) {
    console.error('Не удалось удалить файл', error)
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="deal-details-backdrop">
      <button
        v-if="selectedDeal"
        type="button"
        class="deal-details__backdrop"
        aria-label="Закрыть карточку сделки"
        @click="emit('close')"
      />
    </Transition>

    <Transition name="deal-details-sheet">
      <section v-if="selectedDeal" class="deal-details-sheet" @click.stop>
      <header class="deal-details-sheet__header">
        <h2 class="deal-details-sheet__title">Сделка #{{ selectedDeal.dealNumber }}</h2>
        <div class="deal-details-sheet__header-actions">
          <button
            v-if="hasLinkedLead"
            type="button"
            class="deal-details-sheet__primary-action"
            @click="handleGoToLeadClick"
          >
            Перейти в лид
          </button>
          <button type="button" class="deal-details-sheet__icon-action deal-details-sheet__icon-action--danger" @click="handleDeleteDeal">
            <NIcon :size="16">
              <TrashOutline />
            </NIcon>
          </button>
          <button type="button" class="deal-details-sheet__close-btn" @click="emit('close')">
            <span aria-hidden="true">×</span>
          </button>
        </div>
      </header>

      <nav class="deal-details-sheet__statuses">
        <button
          v-for="column in DEAL_KANBAN_COLUMNS"
          :key="column.id"
          type="button"
          class="deal-details-sheet__status-btn"
          :class="{ 'deal-details-sheet__status-btn--active': resolvedKanbanColumnId === column.id }"
          :style="getStatusButtonStyle(column.id)"
          @click="handleStatusChange(column.id)"
        >
          {{ column.title }}
        </button>
      </nav>

      <div class="deal-details-sheet__body">
        <div class="deal-details-sheet__left">
          <aside class="deal-details-sheet__sections">
            <button
              v-for="section in DEAL_SECTIONS"
              :key="section.id"
              type="button"
              class="deal-details-sheet__section-btn"
              :class="{ 'deal-details-sheet__section-btn--active': activeSection === section.id }"
              @click="activeSection = section.id"
            >
              {{ section.title }}
            </button>
          </aside>

          <section class="deal-details-sheet__content">
            <div v-if="activeSection === 'general'" class="deal-details-sheet__panel">
              <h3 class="deal-details-sheet__panel-title">Общая информация</h3>
              <dl class="deal-details-sheet__info-list">
                <div class="deal-details-sheet__info-row"><dt>Имя</dt><dd>{{ selectedDeal.firstName || '—' }}</dd></div>
                <div class="deal-details-sheet__info-row"><dt>Отчество</dt><dd>{{ selectedDeal.patronymic || '—' }}</dd></div>
                <div class="deal-details-sheet__info-row"><dt>Телефон</dt><dd>{{ selectedDeal.phone }}</dd></div>
                <div class="deal-details-sheet__info-row"><dt>Источник</dt><dd>{{ selectedDeal.trafficSource || '—' }}</dd></div>
                <div class="deal-details-sheet__info-row"><dt>Сумма</dt><dd>{{ selectedDeal.totalAmount }} ₽</dd></div>
                <div class="deal-details-sheet__info-row"><dt>Создатель</dt><dd>{{ selectedDeal.createdBy }}</dd></div>
                <div class="deal-details-sheet__info-row"><dt>Создана</dt><dd>{{ formatDateTime(selectedDeal.createdAt) }}</dd></div>
                <div
                  v-if="resolvedKanbanColumnId === 'failed' && selectedDeal.failureReason"
                  class="deal-details-sheet__info-row"
                >
                  <dt>Причина провала</dt>
                  <dd>{{ selectedDeal.failureReason }}</dd>
                </div>
              </dl>
            </div>

            <div v-else-if="activeSection === 'task'" class="deal-details-sheet__panel">
              <h3 class="deal-details-sheet__panel-title">Задача</h3>

              <button
                type="button"
                class="deal-details-sheet__task-create-btn"
                @click="openCreateTaskModal"
              >
                Создать задачу
              </button>

              <section class="deal-details-sheet__tasks-list-wrapper">
                <h4 class="deal-details-sheet__sub-title">Задачи по сделке</h4>

                <ul v-if="dealTasks.length > 0" class="deal-details-sheet__tasks-list">
                  <li
                    v-for="task in dealTasks"
                    :key="task.id"
                    class="deal-details-sheet__task-item"
                    :class="{ 'deal-details-sheet__task-item--clickable': task.status === 'active' }"
                    @click="openDealTaskDetails(task)"
                  >
                    <div class="deal-details-sheet__task-item-header">
                      <p class="deal-details-sheet__task-item-title">{{ task.title }}</p>
                      <span
                        class="deal-details-sheet__task-item-status"
                        :class="{
                          'deal-details-sheet__task-item-status--completed': task.status === 'completed',
                        }"
                      >
                        {{ getDealTaskStatusLabel(task.status) }}
                      </span>
                    </div>
                    <p class="deal-details-sheet__task-item-text">{{ task.text }}</p>
                    <p class="deal-details-sheet__task-item-meta">
                      Срок: {{ formatDateTime(task.dueAt) || 'Не указан' }}
                    </p>
                  </li>
                </ul>
                <p v-else class="deal-details-sheet__task-empty">По этой сделке пока нет задач</p>
              </section>
            </div>

            <div v-else-if="activeSection === 'products'" class="deal-details-sheet__panel">
              <h3 class="deal-details-sheet__panel-title">Услуги/Товары</h3>
              <DealProductsEditor v-model="dealProductRows" @persist="schedulePersistDealProducts" />
            </div>

            <div v-else-if="activeSection === 'pickup'" class="deal-details-sheet__panel">
              <h3 class="deal-details-sheet__panel-title">Самовывоз</h3>
              <p v-if="isDealPickupSectionLocked" class="deal-details-sheet__lock-note">
                {{ PICKUP_SECTION_LOCKED_MESSAGE }}
              </p>

              <label class="deal-details-sheet__field">
                <span class="deal-details-sheet__label">Адрес самовывоза</span>
                <input
                  v-model="currentPickupAddress"
                  type="text"
                  class="deal-details-sheet__input"
                  placeholder="Укажите адрес самовывоза"
                  :disabled="isDealPickupFieldsDisabled"
                  @blur="flushPersistDealPickupDelivery"
                />
              </label>

              <label class="deal-details-sheet__field">
                <span class="deal-details-sheet__label">Дата и время самовывоза</span>
                <DateTimeField
                  :display-value="formatDateTime(currentPickupDate)"
                  :has-value="currentPickupDate !== null"
                  :disabled="isDealPickupFieldsDisabled"
                  @open="openPickupDateModal"
                  @clear="clearPickupDate"
                />
              </label>
            </div>

            <div v-else-if="activeSection === 'delivery'" class="deal-details-sheet__panel">
              <h3 class="deal-details-sheet__panel-title">Доставка</h3>
              <p v-if="isDealDeliverySectionLocked" class="deal-details-sheet__lock-note">
                {{ DELIVERY_SECTION_LOCKED_MESSAGE }}
              </p>

              <label class="deal-details-sheet__field">
                <span class="deal-details-sheet__label">Адрес доставки</span>
                <input
                  v-model="currentDeliveryAddress"
                  type="text"
                  class="deal-details-sheet__input"
                  placeholder="Укажите адрес доставки"
                  :disabled="isDealDeliveryFieldsDisabled"
                  @blur="flushPersistDealPickupDelivery"
                />
              </label>

              <label class="deal-details-sheet__field">
                <span class="deal-details-sheet__label">Дата и время доставки</span>
                <DateTimeField
                  :display-value="formatDateTime(currentDeliveryDate)"
                  :has-value="currentDeliveryDate !== null"
                  :disabled="isDealDeliveryFieldsDisabled"
                  @open="openDeliveryDateModal"
                  @clear="clearDeliveryDate"
                />
              </label>

              <label class="deal-details-sheet__field">
                <span class="deal-details-sheet__label">Курьер</span>
                <input
                  v-model="currentCourier"
                  type="text"
                  class="deal-details-sheet__input"
                  placeholder="Укажите курьера"
                  :disabled="isDealDeliveryFieldsDisabled"
                  @blur="flushPersistDealPickupDelivery"
                />
              </label>
            </div>

            <div v-else-if="activeSection === 'production'" class="deal-details-sheet__panel">
              <h3 class="deal-details-sheet__panel-title">Производство</h3>
              <label class="deal-details-sheet__field">
                <span class="deal-details-sheet__label">Номенклатура</span>
                <NSelect
                  v-model:value="currentProductionNomenclature"
                  :options="productionNomenclatureOptions"
                  filterable
                  clearable
                  placeholder="Выберите номенклатуру"
                  class="deal-details-sheet__select"
                  @update:value="schedulePersistDealProduction"
                />
              </label>
              <label class="deal-details-sheet__field">
                <span class="deal-details-sheet__label">Дата и время</span>
                <DateTimeField
                  :display-value="formatDateTime(currentProductionDueAt)"
                  :has-value="currentProductionDueAt !== null"
                  @open="openProductionDateModal"
                  @clear="clearProductionDueAt"
                />
              </label>
              <label class="deal-details-sheet__field">
                <span class="deal-details-sheet__label">Сотрудник</span>
                <NSelect
                  v-model:value="currentProductionEmployee"
                  :options="productionEmployeeOptions"
                  filterable
                  clearable
                  placeholder="Выберите сотрудника"
                  class="deal-details-sheet__select"
                  @update:value="schedulePersistDealProduction"
                  @blur="flushPersistDealProduction"
                />
              </label>
            </div>
          </section>
        </div>

        <aside class="deal-details-sheet__communication">
          <section class="deal-details-sheet__side-block">
            <h3 class="deal-details-sheet__side-title">Комментарий</h3>
            <div class="deal-details-sheet__comment-box">
              <textarea
                v-model="commentDraft"
                class="deal-details-sheet__comment-area"
                rows="4"
                placeholder="Заметки по сделке..."
                @blur="persistDealComment"
              />

              <div class="deal-details-sheet__comment-box-foot">
                <input id="deal-attachment-input" type="file" class="deal-details-sheet__file-input" multiple @change="handleAttachmentChange" />
                <button type="button" class="deal-details-sheet__attachment-btn" @click="triggerAttachmentPick">
                  Прикрепить файл
                </button>

                <EntityAttachmentList
                  :attachments="selectedDeal.attachments"
                  @remove="removeAttachment"
                />
              </div>
            </div>
          </section>

          <section class="deal-details-sheet__side-block deal-details-sheet__side-block--timeline">
            <h3 class="deal-details-sheet__side-title">Таймлайн</h3>

            <ul v-if="sortedActivities.length > 0" class="deal-details-sheet__timeline">
              <li
                v-for="entry in sortedActivities"
                :key="entry.id"
                class="deal-details-sheet__timeline-entry"
                :class="{
                  'deal-details-sheet__timeline-entry--comment': entry.type === 'comment',
                  'deal-details-sheet__timeline-entry--system': entry.type === 'system',
                }"
              >
                <div class="deal-details-sheet__timeline-entry-body">
                  <p class="deal-details-sheet__timeline-text">{{ entry.text }}</p>
                  <p class="deal-details-sheet__timeline-meta">
                    {{ formatDateTime(entry.createdAt) }}
                  </p>
                </div>
              </li>
            </ul>
            <p v-else class="deal-details-sheet__timeline-empty">Таймлайн пока пуст</p>
          </section>
        </aside>
      </div>

      <AppModal
        v-model:show="isProductionDateModalOpen"
        title="Дата и время производства"
        width="wide"
        body-variant="date"
      >
        <NDatePicker
          v-model:value="currentProductionDueAt"
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
      >
        <NDatePicker
          v-model:value="currentPickupDate"
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
      >
        <NDatePicker
          v-model:value="currentDeliveryDate"
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
        v-model:show="isCreateTaskModalOpen"
        title="Новая задача по сделке"
        width="wide"
        actions-align="end"
        close-label="Закрыть окно создания задачи"
        @close="closeCreateTaskModal"
      >
        <div class="deal-details-sheet__modal-fields">
          <label class="deal-details-sheet__field">
            <span class="deal-details-sheet__label">Заголовок</span>
            <input
              v-model="taskForm.title"
              type="text"
              class="deal-details-sheet__input"
              placeholder="Введите заголовок задачи"
            />
          </label>

          <label class="deal-details-sheet__field">
            <span class="deal-details-sheet__label">Текст задачи</span>
            <textarea
              v-model="taskForm.text"
              class="deal-details-sheet__textarea"
              rows="5"
              placeholder="Введите текст задачи"
            />
          </label>

          <label class="deal-details-sheet__field">
            <span class="deal-details-sheet__label">Крайний срок</span>
            <button
              type="button"
              class="deal-details-sheet__date-trigger"
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
        v-model:show="isStatusValidationModalOpen"
        title="Нельзя изменить статус"
        body-variant="center"
        @close="closeStatusValidationModal"
      >
        <p class="app-modal__message">
          {{ statusValidationMessage }}
        </p>

        <template #actions>
          <AppModalButton block @click="handleStatusValidationGoToSection">
            Перейти в раздел «{{ statusValidationSectionTitle }}»
          </AppModalButton>
        </template>
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
          placeholder="Укажите пожалуйста причину провала сделки"
        />

        <template #actions>
          <AppModalButton :disabled="!canConfirmFailureReason" @click="confirmFailureReason">
            Подтвердить
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
              <NButton
                class="app-modal__date-confirm-btn"
                type="primary"
                size="small"
                @click="handleTaskDateConfirm(onConfirm)"
              >
                {{ text }}
              </NButton>
            </div>
          </template>
        </NDatePicker>
      </AppModal>
      </section>
    </Transition>

    <TaskDetailsSheet
      :task-id="selectedDealTask?.id ?? null"
      :task="selectedDealTask"
      @close="closeDealTaskDetails"
    />
  </Teleport>
</template>

<style scoped>
.deal-details__backdrop {
  position: fixed;
  inset: 0;
  border: 0;
  background: rgba(15, 23, 42, 0.2);
  z-index: 180;
  cursor: default;
}

.deal-details-sheet {
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

.deal-details-sheet__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid #e2e8f0;
}

.deal-details-sheet__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.deal-details-sheet__header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.deal-details-sheet__primary-action {
  padding: 8px 14px;
  border: 1px solid #1f883d;
  border-radius: 8px;
  background: #1f883d;
  color: #ffffff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.deal-details-sheet__primary-action:hover {
  background: #197a35;
}

.deal-details-sheet__icon-action {
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
}

.deal-details-sheet__icon-action:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.deal-details-sheet__icon-action--danger:hover {
  color: #dc2626;
}

.deal-details-sheet__close-btn {
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
}

.deal-details-sheet__close-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1f2937;
}

.deal-details-sheet__statuses {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid #e2e8f0;
}

.deal-details-sheet__status-btn {
  width: 100%;
  min-width: 0;
  padding: 8px 10px;
  border: 1px solid #dbe3ed;
  border-radius: 8px;
  background: #f8fafc;
  color: #4a5568;
  font-size: 13px;
  font-weight: 500;
  text-align: center;
  cursor: pointer;
  box-shadow: inset 0 -2px 0 0 var(--status-color);
}

.deal-details-sheet__status-btn--active {
  background: var(--status-bg);
  border-color: var(--status-border);
  color: var(--status-color);
  box-shadow: inset 0 0 0 1px var(--status-color), inset 0 -2px 0 0 var(--status-color);
}

.deal-details-sheet__status-btn--readonly {
  cursor: default;
  opacity: 0.92;
}

.deal-details-sheet__status-btn--readonly:disabled {
  opacity: 1;
}

.deal-details-sheet__modal-fields {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.deal-details-sheet__body {
  flex: 1 1 auto;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 16px;
  padding: 16px;
}

.deal-details-sheet__left {
  min-width: 0;
  min-height: 0;
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
}

.deal-details-sheet__sections {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.deal-details-sheet__section-btn {
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: #334155;
  font-size: 14px;
  text-align: left;
  padding: 9px 10px;
  cursor: pointer;
}

.deal-details-sheet__section-btn--active {
  background: #ffffff;
  border-color: #d1d9e2;
  color: #0f172a;
  font-weight: 600;
}

.deal-details-sheet__content {
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 16px;
}

.deal-details-sheet__panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.deal-details-sheet__panel-title {
  margin: 0;
  font-size: 18px;
  color: #0f172a;
}

.deal-details-sheet__info-list {
  margin: 0;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px 16px;
}

.deal-details-sheet__info-row {
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.deal-details-sheet__info-row dt {
  font-size: 12px;
  color: #64748b;
}

.deal-details-sheet__info-row dd {
  margin: 0;
  font-size: 14px;
  color: #0f172a;
  font-weight: 500;
}

.deal-details-sheet__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.deal-details-sheet__label {
  font-size: 13px;
  color: #475569;
}

.deal-details-sheet__input,
.deal-details-sheet__textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #0f172a;
  font: inherit;
  padding: 9px 12px;
}

.deal-details-sheet__textarea {
  resize: vertical;
  min-height: 96px;
}

.deal-details-sheet__select {
  width: 100%;
}

.deal-details-sheet__empty {
  margin: 0;
  color: #64748b;
  font-size: 13px;
}

.deal-details-sheet__lock-note {
  margin: 0;
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
  color: #475569;
  font-size: 13px;
  line-height: 1.45;
}

.deal-details-sheet__date-trigger {
  width: 100%;
  text-align: left;
  padding: 9px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #0f172a;
  font-size: 14px;
  cursor: pointer;
}

.deal-details-sheet__date-trigger:disabled {
  cursor: not-allowed;
  background: #f8fafc;
  color: #94a3b8;
}

.deal-details-sheet__communication {
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

.deal-details-sheet__side-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.deal-details-sheet__side-block--timeline {
  flex: 1 1 auto;
  min-height: 0;
}

.deal-details-sheet__side-title {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  color: #1a202c;
}

.deal-details-sheet__comment-box {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  overflow: hidden;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.deal-details-sheet__comment-box:focus-within {
  border-color: #cbd5e1;
  box-shadow: 0 0 0 3px rgba(31, 136, 61, 0.08);
}

.deal-details-sheet__comment-area {
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

.deal-details-sheet__comment-area:focus {
  outline: none;
}

.deal-details-sheet__comment-box-foot {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 12px;
  border-top: 1px solid #e2e8f0;
  background: #f8fafc;
}

.deal-details-sheet__file-input {
  display: none;
}

.deal-details-sheet__attachment-btn {
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

.deal-details-sheet__attachment-btn:hover {
  border-color: #cbd5e1;
  background: #f8fafc;
  color: #1a202c;
}

.deal-details-sheet__timeline {
  margin: 0;
  padding: 4px 0 0 14px;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0;
  border-left: 1px solid #e2e8f0;
}

.deal-details-sheet__timeline-entry {
  position: relative;
  padding: 0 0 14px 12px;
}

.deal-details-sheet__timeline-entry::before {
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

.deal-details-sheet__timeline-entry--comment::before {
  background: #1f883d;
  box-shadow: 0 0 0 3px rgba(31, 136, 61, 0.14);
}

.deal-details-sheet__timeline-entry--system::before {
  background: #cbd5e1;
}

.deal-details-sheet__timeline-entry-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.deal-details-sheet__timeline-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.45;
  color: #1a202c;
  white-space: pre-wrap;
  word-break: break-word;
}

.deal-details-sheet__timeline-entry--system .deal-details-sheet__timeline-text {
  color: #4a5568;
}

.deal-details-sheet__timeline-meta {
  margin: 0;
  font-size: 12px;
  color: #718096;
}

.deal-details-sheet__timeline-empty {
  margin: 0;
  font-size: 13px;
  color: #718096;
}

@media (max-width: 1200px) {
  .deal-details-sheet__body {
    grid-template-columns: minmax(0, 1fr);
  }

  .deal-details-sheet__left {
    grid-template-columns: minmax(0, 1fr);
  }
}

.deal-details-sheet-enter-active,
.deal-details-sheet-leave-active {
  transition:
    transform 0.28s ease,
    opacity 0.28s ease;
}

.deal-details-sheet-enter-from,
.deal-details-sheet-leave-to {
  transform: translateY(100%);
  opacity: 0.98;
}

.deal-details-backdrop-enter-active,
.deal-details-backdrop-leave-active {
  transition: opacity 0.2s ease;
}

.deal-details-backdrop-enter-from,
.deal-details-backdrop-leave-to {
  opacity: 0;
}

.deal-details-sheet__task-create-btn {
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

.deal-details-sheet__sub-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
}

.deal-details-sheet__tasks-list-wrapper {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.deal-details-sheet__tasks-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.deal-details-sheet__task-item {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #ffffff;
  padding: 8px 9px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.deal-details-sheet__task-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.deal-details-sheet__task-item-title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #1a202c;
}

.deal-details-sheet__task-item-status {
  border: 1px solid #d1d9e2;
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 12px;
  font-weight: 600;
  color: #1f883d;
  background: #f0fff4;
  white-space: nowrap;
}

.deal-details-sheet__task-item-status--completed {
  color: #4a5568;
  background: #f8fafc;
}

.deal-details-sheet__task-item-text {
  margin: 0;
  font-size: 12px;
  color: #334155;
  white-space: pre-wrap;
  word-break: break-word;
}

.deal-details-sheet__task-item-meta {
  margin: 0;
  font-size: 11px;
  color: #64748b;
}

.deal-details-sheet__task-item--clickable {
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.deal-details-sheet__task-item--clickable:hover {
  border-color: #cbd5e1;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.06);
}

.deal-details-sheet__task-empty {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}
</style>
