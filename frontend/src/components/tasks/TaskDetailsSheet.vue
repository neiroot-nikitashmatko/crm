<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NDatePicker, NInput } from 'naive-ui'
import { useDeals } from '@/composables/useDeals'
import { useTasks } from '@/composables/useTasks'
import type { Task, TaskActivityEntry } from '@/types/task'
import TaskAttachments from './TaskAttachments.vue'
import AppModal from '@/components/common/AppModal.vue'

const props = defineProps<{
  taskId: string | null
  task?: Task | null
}>()

const emit = defineEmits<{
  close: []
}>()

const { tasks, completeTask, updateTask, addTaskAttachments, removeTaskAttachment, addTaskComment } = useTasks()
const { deals, getLeadDeal, loadDeals } = useDeals()
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

const timelineEntries = computed<TaskActivityEntry[]>(() => {
  if (!selectedTask.value) return []
  return [...selectedTask.value.activities].sort(
    (a, b) => b.createdAt.getTime() - a.createdAt.getTime(),
  )
})

const formattedDueAt = computed(() =>
  selectedTask.value?.dueAt ? new Date(selectedTask.value.dueAt).toLocaleString('ru-RU') : 'Без срока',
)

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
  void loadDeals()
})

function syncDrafts() {
  if (!selectedTask.value) return
  titleDraft.value = selectedTask.value.title
  textDraft.value = selectedTask.value.text
  dueAtDraft.value = selectedTask.value.dueAt
}

watch(
  () => selectedTask.value?.id,
  () => {
    syncDrafts()
  },
  { immediate: true },
)

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

function handleAddAttachments(files: File[]) {
  if (!selectedTask.value) return

  void (async () => {
    try {
      await addTaskAttachments(selectedTask.value!.id, files)
    } catch (error) {
      console.error('Не удалось загрузить вложения', error)
    }
  })()
}

function handleRemoveAttachment(attachmentId: string) {
  if (!selectedTask.value) return

  void (async () => {
    try {
      await removeTaskAttachment(selectedTask.value!.id, attachmentId)
    } catch (error) {
      console.error('Не удалось удалить файл', error)
    }
  })()
}

async function addCommentToTimeline() {
  if (!selectedTask.value) return
  const text = commentDraft.value.trim()
  if (!text) return

  try {
    await addTaskComment(selectedTask.value.id, text)
    commentDraft.value = ''
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
          <div class="task-details-sheet__field">
            <p class="task-details-sheet__label">Ответственный</p>
            <p class="task-details-sheet__value">{{ selectedTask.createdBy }}</p>
          </div>

          <div class="task-details-sheet__field">
            <div class="task-details-sheet__row">
              <p class="task-details-sheet__label">Крайний срок</p>
              <button type="button" class="task-details-sheet__link" @click="openDateModal">
                Изменить
              </button>
            </div>
            <p class="task-details-sheet__value">{{ formattedDueAt }}</p>
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

          <div class="task-details-sheet__field">
            <p class="task-details-sheet__label">Вложения</p>
            <TaskAttachments
              :attachments="selectedTask.attachments"
              @add-files="handleAddAttachments"
              @remove="handleRemoveAttachment"
            />
          </div>
        </div>

        <aside class="task-details-sheet__timeline">
          <p class="task-details-sheet__timeline-title">Комментарий и таймлайн</p>
          <div class="task-details-sheet__comment-form">
            <NInput
              v-model:value="commentDraft"
              type="textarea"
              :autosize="{ minRows: 2, maxRows: 4 }"
              placeholder="Добавьте комментарий"
            />
            <button
              type="button"
              class="task-details-sheet__button task-details-sheet__button--primary task-details-sheet__comment-btn"
              @click="addCommentToTimeline"
            >
              Добавить
            </button>
          </div>
          <ul class="task-details-sheet__timeline-list">
            <li
              v-for="entry in timelineEntries"
              :key="entry.id"
              class="task-details-sheet__timeline-item"
            >
              <p class="task-details-sheet__timeline-text">{{ entry.text }}</p>
              <p class="task-details-sheet__timeline-meta">
                {{ entry.author }} · {{ entry.createdAt.toLocaleString('ru-RU') }}
              </p>
            </li>
          </ul>
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
  grid-template-columns: minmax(0, 1fr) 350px;
  gap: 0;
}

.task-details-sheet__main {
  min-width: 0;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-details-sheet__main::-webkit-scrollbar,
.task-details-sheet__timeline::-webkit-scrollbar {
  width: 8px;
}

.task-details-sheet__main::-webkit-scrollbar-track,
.task-details-sheet__timeline::-webkit-scrollbar-track {
  background: transparent;
}

.task-details-sheet__main::-webkit-scrollbar-thumb,
.task-details-sheet__timeline::-webkit-scrollbar-thumb {
  background: #cbd5e0;
  border-radius: 4px;
}

.task-details-sheet__field {
  display: grid;
  gap: 6px;
}

.task-details-sheet__row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-details-sheet__label {
  margin: 0;
  font-size: 12px;
  color: #718096;
}

.task-details-sheet__value,
.task-details-sheet__text {
  margin: 0;
  color: #1a202c;
  font-size: 14px;
}

.task-details-sheet__link {
  border: none;
  background: transparent;
  color: #1f883d;
  cursor: pointer;
  font-size: 12px;
}

.task-details-sheet__timeline {
  min-width: 0;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  padding: 20px 20px 24px;
  border-left: 1px solid #e2e8f0;
  background: #f8fafc;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.task-details-sheet__timeline-title {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  color: #1a202c;
}

.task-details-sheet__comment-form {
  display: grid;
  gap: 8px;
}

.task-details-sheet__comment-btn {
  justify-self: flex-end;
}

.task-details-sheet__timeline-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 8px;
}

.task-details-sheet__timeline-item {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 8px;
  background: #ffffff;
}

.task-details-sheet__timeline-text {
  margin: 0;
  font-size: 13px;
  color: #1a202c;
}

.task-details-sheet__timeline-meta {
  margin: 4px 0 0;
  font-size: 12px;
  color: #718096;
}

.task-details-sheet__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 14px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  line-height: 1;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.task-details-sheet__button--primary {
  border: 1px solid #1f883d;
  background: #1f883d;
  color: #ffffff;
}

.task-details-sheet__button--primary:hover {
  background: #197634;
  border-color: #197634;
}

.task-details-sheet__button--secondary {
  border: 1px solid #d1d9e2;
  background: #ffffff;
  color: #475569;
}

.task-details-sheet__button--secondary:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1f2937;
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

  .task-details-sheet__timeline {
    border-left: none;
    border-top: 1px solid #e2e8f0;
  }
}

@supports not (scrollbar-gutter: stable) {
  .task-details-sheet__main,
  .task-details-sheet__timeline {
    overflow-y: scroll;
  }
}
</style>
