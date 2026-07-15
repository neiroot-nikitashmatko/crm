<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { AttachOutline, PaperPlaneOutline } from '@vicons/ionicons5'
import { NButton, NDatePicker, NIcon, NInput } from 'naive-ui'
import EntityAttachmentList from '@/components/attachments/EntityAttachmentList.vue'
import { useDeals } from '@/composables/useDeals'
import { useLeads } from '@/composables/useLeads'
import { useTasks } from '@/composables/useTasks'
import type { StoredAttachment } from '@/types/attachment'
import type { Task, TaskActivityEntry } from '@/types/task'
import AppModal from '@/components/common/AppModal.vue'
import {
  clearTaskCommentDraft,
  loadTaskCommentDraft,
  saveTaskCommentDraft,
} from '@/utils/taskCommentDraft'

const props = defineProps<{
  taskId: string | null
  task?: Task | null
}>()

const emit = defineEmits<{
  close: []
}>()

const { tasks, completeTask, updateTask, addTaskAttachments, removeTaskAttachment, addTaskComment } = useTasks()
const { deals, getLeadDeal, loadDeals } = useDeals()
const { leads, loadLeads } = useLeads()
const router = useRouter()
const isDateModalOpen = ref(false)
const commentDraft = ref('')
const titleDraft = ref('')
const textDraft = ref('')
const dueAtDraft = ref<number | null>(null)
const WORKING_HOURS = [9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]
const TIME_MINUTE_STEP = 5
const TIME_PICKER_PROPS = {
  format: 'HH:mm',
  hours: WORKING_HOURS,
  minutes: TIME_MINUTE_STEP,
  actions: ['confirm'] as Array<'confirm'>,
}

const selectedTask = computed(() =>
  props.taskId
    ? tasks.value.find((task) => task.id === props.taskId) ?? props.task ?? null
    : props.task ?? null,
)

const sortedActivities = computed<TaskActivityEntry[]>(() => {
  if (!selectedTask.value) return []
  return [...selectedTask.value.activities].sort(
    (a, b) => b.createdAt.getTime() - a.createdAt.getTime(),
  )
})

const attachmentItems = computed<StoredAttachment[]>(() => {
  if (!selectedTask.value) return []
  return selectedTask.value.attachments.map((attachment) => ({
    id: attachment.id,
    name: attachment.name,
    size: attachment.size,
    mimeType: attachment.mimeType,
    uploadedBy: attachment.uploadedBy,
    uploadedAt: attachment.uploadedAt.getTime(),
  }))
})

const formattedDueAt = computed(() =>
  selectedTask.value?.dueAt ? new Date(selectedTask.value.dueAt).toLocaleString('ru-RU') : 'Без срока',
)

const linkedLead = computed(() => {
  const leadId = selectedTask.value?.leadId
  if (!leadId) return null
  return leads.value.find((lead) => lead.id === leadId) ?? null
})

const linkedDeal = computed(() => {
  const task = selectedTask.value
  if (!task) return null

  if (task.dealId) {
    const byId = deals.value.find((deal) => deal.id === task.dealId)
    if (byId) return byId
  }

  if (task.leadId) {
    return getLeadDeal(task.leadId) ?? null
  }

  return null
})

const clientSource = computed(() => linkedLead.value ?? linkedDeal.value)

const displayClientFirstName = computed(
  () => clientSource.value?.firstName?.trim() || selectedTask.value?.clientFirstName?.trim() || '—',
)
const displayClientPatronymic = computed(
  () => clientSource.value?.patronymic?.trim() || selectedTask.value?.clientPatronymic?.trim() || '—',
)
const displayClientPhone = computed(
  () => clientSource.value?.phone?.trim() || selectedTask.value?.clientPhone?.trim() || '—',
)
const displayTrafficSource = computed(
  () => clientSource.value?.trafficSource?.trim() || selectedTask.value?.trafficSource?.trim() || '—',
)
const displayResponsible = computed(() => selectedTask.value?.createdBy?.trim() || '—')

const linkedDealId = computed(() => {
  const task = selectedTask.value
  if (!task) return null

  if (task.dealId && deals.value.some((deal) => deal.id === task.dealId)) {
    return task.dealId
  }

  if (task.leadId) {
    return getLeadDeal(task.leadId)?.id ?? null
  }

  return null
})

const linkedLeadId = computed(() => selectedTask.value?.leadId ?? null)

const navigationTarget = computed<'deal' | 'lead' | null>(() => {
  if (linkedDealId.value) return 'deal'
  if (linkedLeadId.value) return 'lead'
  return null
})

const navigationButtonLabel = computed(() =>
  navigationTarget.value === 'deal' ? 'Перейти в сделку' : 'Перейти в лид',
)

onMounted(() => {
  void Promise.all([loadDeals(), loadLeads()])
})

function syncDrafts() {
  if (!selectedTask.value) return
  titleDraft.value = selectedTask.value.title
  textDraft.value = selectedTask.value.text
  dueAtDraft.value = selectedTask.value.dueAt
}

watch(
  () => selectedTask.value?.id,
  (taskId) => {
    syncDrafts()
    commentDraft.value = taskId ? loadTaskCommentDraft(taskId) : ''
  },
  { immediate: true },
)

watch(commentDraft, (text) => {
  const taskId = selectedTask.value?.id
  if (!taskId) return
  saveTaskCommentDraft(taskId, text)
})

function formatDateTime(value: Date | number | null) {
  if (value === null) return '—'
  const timestamp = value instanceof Date ? value.getTime() : value
  if (Number.isNaN(timestamp)) return '—'
  return new Date(timestamp).toLocaleString('ru-RU')
}

function openDateModal() {
  syncDrafts()
  const initialValue = dueAtDraft.value ?? Date.now()
  dueAtDraft.value = normalizeDateTimeValue(initialValue)
  isDateModalOpen.value = true
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

async function persistTaskFields() {
  if (!selectedTask.value) return
  await updateTask(selectedTask.value.id, {
    title: titleDraft.value,
    text: textDraft.value,
    dueAt: dueAtDraft.value,
  })
}

function handleDueDateConfirm(onConfirm: () => void) {
  onConfirm()
  isDateModalOpen.value = false
  void persistTaskFields()
}

async function handleCompleteTask() {
  if (!selectedTask.value || selectedTask.value.status === 'completed') return
  try {
    await completeTask(selectedTask.value.id)
    emit('close')
  } catch (error) {
    console.error('Не удалось завершить задачу', error)
  }
}

function triggerAttachmentPick() {
  ;(document.getElementById('task-attachment-input') as HTMLInputElement | null)?.click()
}

function handleAttachmentChange(event: Event) {
  if (!selectedTask.value) return
  const target = event.target as HTMLInputElement | null
  const files = target?.files ? Array.from(target.files) : []
  if (target) target.value = ''
  if (files.length === 0) return

  void uploadAttachments(files)
}

async function uploadAttachments(files: File[]) {
  if (!selectedTask.value) return

  try {
    await addTaskAttachments(selectedTask.value.id, files)
  } catch (error) {
    console.error('Не удалось загрузить вложения', error)
  }
}

async function handleRemoveAttachment(attachmentId: string) {
  if (!selectedTask.value) return

  try {
    await removeTaskAttachment(selectedTask.value.id, attachmentId)
  } catch (error) {
    console.error('Не удалось удалить файл', error)
  }
}

async function sendTaskComment() {
  if (!selectedTask.value) return

  const taskId = selectedTask.value.id
  const text = commentDraft.value.trim()
  if (!text) return

  try {
    await addTaskComment(taskId, text)
    commentDraft.value = ''
    clearTaskCommentDraft(taskId)
  } catch (error) {
    console.error('Не удалось добавить комментарий', error)
  }
}

async function handleNavigationClick() {
  if (!selectedTask.value || !navigationTarget.value) return

  emit('close')

  if (navigationTarget.value === 'deal' && linkedDealId.value) {
    await router.push({ name: 'deals', query: { dealId: linkedDealId.value } })
    return
  }

  if (navigationTarget.value === 'lead' && linkedLeadId.value) {
    await router.push({ name: 'leads', query: { leadId: linkedLeadId.value } })
  }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="selectedTask" class="task-details-sheet__backdrop" @click.self="emit('close')">
      <section class="task-details-sheet">
        <header class="task-details-sheet__header">
          <h2 class="task-details-sheet__title">Задача</h2>
          <div class="task-details-sheet__header-actions">
            <button
              v-if="navigationTarget"
              type="button"
              class="task-details-sheet__action"
              @click="handleNavigationClick"
            >
              {{ navigationButtonLabel }}
            </button>
            <button
              type="button"
              class="task-details-sheet__action"
              :disabled="selectedTask.status === 'completed'"
              @click="handleCompleteTask"
            >
              Завершить задачу
            </button>
            <button type="button" class="task-details-sheet__close-btn" @click="emit('close')">
              <span aria-hidden="true">×</span>
            </button>
          </div>
        </header>

        <div class="task-details-sheet__body">
          <div class="task-details-sheet__main">
            <div class="task-details-sheet__info-grid">
              <div class="task-details-sheet__info-column">
                <div class="task-details-sheet__info-item">
                  <p class="task-details-sheet__label">Имя</p>
                  <p class="task-details-sheet__value">{{ displayClientFirstName }}</p>
                </div>
                <div class="task-details-sheet__info-item">
                  <p class="task-details-sheet__label">Отчество</p>
                  <p class="task-details-sheet__value">{{ displayClientPatronymic }}</p>
                </div>
                <div class="task-details-sheet__info-item">
                  <p class="task-details-sheet__label">Телефон</p>
                  <p class="task-details-sheet__value">{{ displayClientPhone }}</p>
                </div>
              </div>

              <div class="task-details-sheet__info-column">
                <div class="task-details-sheet__info-item">
                  <p class="task-details-sheet__label">Ответственный</p>
                  <p class="task-details-sheet__value">{{ displayResponsible }}</p>
                </div>
                <div class="task-details-sheet__info-item">
                  <p class="task-details-sheet__label">Источник</p>
                  <p class="task-details-sheet__value">{{ displayTrafficSource }}</p>
                </div>
                <div class="task-details-sheet__info-item">
                  <div class="task-details-sheet__row">
                    <p class="task-details-sheet__label">Крайний срок</p>
                    <button type="button" class="task-details-sheet__link" @click="openDateModal">
                      Изменить
                    </button>
                  </div>
                  <p class="task-details-sheet__value">{{ formattedDueAt }}</p>
                </div>
              </div>
            </div>

            <div class="task-details-sheet__field">
              <p class="task-details-sheet__label">Заголовок</p>
              <NInput v-model:value="titleDraft" placeholder="Заголовок задачи" @blur="persistTaskFields" />
            </div>

            <div class="task-details-sheet__field">
              <p class="task-details-sheet__label">Текст задачи</p>
              <NInput
                v-model:value="textDraft"
                type="textarea"
                :autosize="{ minRows: 4, maxRows: 8 }"
                placeholder="Опишите задачу"
                @blur="persistTaskFields"
              />
            </div>
          </div>

          <aside class="task-details-sheet__communication">
            <section class="task-details-sheet__side-block">
              <h3 class="task-details-sheet__side-title">Комментарий</h3>
              <div class="task-details-sheet__comment-box">
                <textarea
                  v-model="commentDraft"
                  class="task-details-sheet__comment-area"
                  rows="4"
                  placeholder="Напишите комментарий..."
                />

                <div class="task-details-sheet__comment-box-foot">
                  <input
                    id="task-attachment-input"
                    type="file"
                    class="task-details-sheet__file-input"
                    multiple
                    @change="handleAttachmentChange"
                  />
                  <div class="task-details-sheet__comment-box-actions">
                    <button
                      type="button"
                      class="task-details-sheet__close-btn"
                      title="Прикрепить файл"
                      aria-label="Прикрепить файл"
                      @click="triggerAttachmentPick"
                    >
                      <NIcon :size="16" :component="AttachOutline" />
                    </button>
                    <button
                      type="button"
                      class="task-details-sheet__close-btn"
                      title="Отправить комментарий"
                      aria-label="Отправить комментарий"
                      :disabled="!commentDraft.trim()"
                      @click="sendTaskComment"
                    >
                      <NIcon :size="16" :component="PaperPlaneOutline" />
                    </button>
                  </div>

                  <EntityAttachmentList
                    :attachments="attachmentItems"
                    @remove="handleRemoveAttachment"
                  />
                </div>
              </div>
            </section>

            <section class="task-details-sheet__side-block task-details-sheet__side-block--timeline">
              <h3 class="task-details-sheet__side-title">Таймлайн</h3>

              <ul v-if="sortedActivities.length > 0" class="task-details-sheet__timeline">
                <li
                  v-for="entry in sortedActivities"
                  :key="entry.id"
                  class="task-details-sheet__timeline-entry"
                  :class="{
                    'task-details-sheet__timeline-entry--comment': entry.type === 'comment',
                    'task-details-sheet__timeline-entry--system': entry.type === 'system',
                  }"
                >
                  <div class="task-details-sheet__timeline-entry-body">
                    <p class="task-details-sheet__timeline-text">{{ entry.text }}</p>
                    <p class="task-details-sheet__timeline-meta">
                      {{ formatDateTime(entry.createdAt) }}
                    </p>
                  </div>
                </li>
              </ul>
              <p v-else class="task-details-sheet__timeline-empty">Таймлайн пока пуст</p>
            </section>
          </aside>
        </div>

        <AppModal
          v-model:show="isDateModalOpen"
          title="Дата и время задачи"
          width="wide"
          body-variant="date"
          close-label="Закрыть выбор даты и времени"
        >
          <NDatePicker
            v-model:value="dueAtDraft"
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
                  @click="handleDueDateConfirm(onConfirm)"
                >
                  {{ text }}
                </NButton>
              </div>
            </template>
          </NDatePicker>
        </AppModal>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.task-details-sheet__backdrop {
  position: fixed;
  inset: 0;
  border: 0;
  background: rgba(15, 23, 42, 0.2);
  z-index: 280;
  cursor: default;
}

.task-details-sheet {
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
  z-index: 290;
  overflow: hidden;
}

.task-details-sheet__header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid #e2e8f0;
}

.task-details-sheet__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.task-details-sheet__header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.task-details-sheet__action {
  padding: 8px 14px;
  border: 1px solid #1f883d;
  border-radius: 8px;
  background: #1f883d;
  color: #ffffff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.task-details-sheet__action:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.task-details-sheet__close-btn {
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

.task-details-sheet__close-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1f2937;
}

.task-details-sheet__body {
  flex: 1 1 auto;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 16px;
  padding: 16px;
}

.task-details-sheet__main {
  min-width: 0;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  padding: 4px 8px 4px 4px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-details-sheet__main::-webkit-scrollbar,
.task-details-sheet__communication::-webkit-scrollbar {
  width: 8px;
}

.task-details-sheet__main::-webkit-scrollbar-track,
.task-details-sheet__communication::-webkit-scrollbar-track {
  background: transparent;
}

.task-details-sheet__main::-webkit-scrollbar-thumb,
.task-details-sheet__communication::-webkit-scrollbar-thumb {
  background: #cbd5e0;
  border-radius: 4px;
}

.task-details-sheet__field {
  display: grid;
  gap: 6px;
}

.task-details-sheet__info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px 20px;
  padding: 12px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
}

.task-details-sheet__info-column {
  min-width: 0;
  display: grid;
  gap: 10px;
  align-content: start;
}

.task-details-sheet__info-item {
  min-width: 0;
  display: grid;
  gap: 4px;
}

.task-details-sheet__row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.task-details-sheet__label {
  margin: 0;
  font-size: 12px;
  color: #718096;
}

.task-details-sheet__value {
  margin: 0;
  color: #1a202c;
  font-size: 14px;
  font-weight: 500;
  line-height: 1.3;
  overflow-wrap: anywhere;
}

.task-details-sheet__link {
  border: none;
  background: transparent;
  color: #1f883d;
  cursor: pointer;
  font-size: 12px;
}

.task-details-sheet__communication {
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

.task-details-sheet__side-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.task-details-sheet__side-block--timeline {
  flex: 1 1 auto;
  min-height: 0;
}

.task-details-sheet__side-title {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  color: #1a202c;
}

.task-details-sheet__comment-box {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  overflow: hidden;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.task-details-sheet__comment-box:focus-within {
  border-color: #cbd5e1;
  box-shadow: 0 0 0 3px rgba(31, 136, 61, 0.08);
}

.task-details-sheet__comment-area {
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

.task-details-sheet__comment-area:focus {
  outline: none;
}

.task-details-sheet__comment-box-foot {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 12px;
  border-top: 1px solid #e2e8f0;
  background: #f8fafc;
}

.task-details-sheet__file-input {
  display: none;
}

.task-details-sheet__comment-box-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.task-details-sheet__comment-box-actions .task-details-sheet__close-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.task-details-sheet__timeline {
  margin: 0;
  padding: 4px 0 0 14px;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0;
  border-left: 1px solid #e2e8f0;
}

.task-details-sheet__timeline-entry {
  position: relative;
  padding: 0 0 14px 12px;
}

.task-details-sheet__timeline-entry::before {
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

.task-details-sheet__timeline-entry--comment::before {
  background: #1f883d;
  box-shadow: 0 0 0 3px rgba(31, 136, 61, 0.14);
}

.task-details-sheet__timeline-entry--system::before {
  background: #cbd5e1;
}

.task-details-sheet__timeline-entry-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.task-details-sheet__timeline-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.45;
  color: #1a202c;
  white-space: pre-wrap;
  word-break: break-word;
}

.task-details-sheet__timeline-entry--system .task-details-sheet__timeline-text {
  color: #4a5568;
}

.task-details-sheet__timeline-meta {
  margin: 0;
  font-size: 12px;
  color: #718096;
}

.task-details-sheet__timeline-empty {
  margin: 0;
  font-size: 13px;
  color: #718096;
}

@media (max-width: 960px) {
  .task-details-sheet {
    top: 8px;
    left: 8px;
    right: 8px;
  }

  .task-details-sheet__body {
    grid-template-columns: minmax(0, 1fr);
  }

  .task-details-sheet__communication {
    padding: 16px 0 0;
    border-left: none;
    border-top: 1px solid #e2e8f0;
  }
}

@supports not (scrollbar-gutter: stable) {
  .task-details-sheet__main,
  .task-details-sheet__communication {
    overflow-y: scroll;
  }
}
</style>
