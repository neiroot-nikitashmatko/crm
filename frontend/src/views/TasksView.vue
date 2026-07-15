<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { NButton, NDatePicker } from 'naive-ui'
import AppModal from '@/components/common/AppModal.vue'
import AppModalButton from '@/components/common/AppModalButton.vue'
import TasksKanban from '@/components/tasks/TasksKanban.vue'
import TasksSectionHeader from '@/components/tasks/TasksSectionHeader.vue'
import { useTasks } from '@/composables/useTasks'

const { addTask } = useTasks()

const isCreateTaskModalOpen = ref(false)
const isTaskDateModalOpen = ref(false)
const taskForm = reactive({
  title: '',
  text: '',
  dueAt: null as number | null,
})

const WORKING_HOURS = [9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19]
const TIME_MINUTE_STEP = 5
const TIME_PICKER_PROPS = {
  format: 'HH:mm',
  hours: WORKING_HOURS,
  minutes: TIME_MINUTE_STEP,
  actions: ['confirm'] as Array<'confirm'>,
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

function resetTaskForm() {
  taskForm.title = ''
  taskForm.text = ''
  taskForm.dueAt = null
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

function openTaskDateModal() {
  taskForm.dueAt = normalizeDateTimeValue(taskForm.dueAt)
  isTaskDateModalOpen.value = true
}

function handleTaskDateConfirm(onConfirm: () => void) {
  onConfirm()
  isTaskDateModalOpen.value = false
}

async function handleCreateTaskClick() {
  if (!canCreateTask.value) return

  try {
    await addTask({
      title: taskForm.title.trim(),
      text: taskForm.text.trim(),
      dueAt: taskForm.dueAt,
    })
    closeCreateTaskModal()
  } catch (error) {
    console.error('Не удалось создать задачу', error)
  }
}
</script>

<template>
  <div class="tasks-page">
    <TasksSectionHeader @create-task="openCreateTaskModal" />
    <TasksKanban />

    <AppModal
      v-model:show="isCreateTaskModalOpen"
      title="Новая задача"
      width="wide"
      actions-align="end"
      close-label="Закрыть окно создания задачи"
      @close="closeCreateTaskModal"
    >
      <div class="tasks-create-modal__fields">
        <label class="tasks-create-modal__field">
          <span class="tasks-create-modal__label">Заголовок</span>
          <input
            v-model="taskForm.title"
            type="text"
            class="tasks-create-modal__input"
            placeholder="Введите заголовок задачи"
          />
        </label>

        <label class="tasks-create-modal__field">
          <span class="tasks-create-modal__label">Текст задачи</span>
          <textarea
            v-model="taskForm.text"
            class="tasks-create-modal__textarea"
            rows="5"
            placeholder="Введите текст задачи"
          />
        </label>

        <label class="tasks-create-modal__field">
          <span class="tasks-create-modal__label">Крайний срок</span>
          <button type="button" class="tasks-create-modal__date-trigger" @click="openTaskDateModal">
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
  </div>
</template>

<style scoped>
.tasks-page {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 64px);
  max-height: calc(100dvh - 64px);
  overflow: hidden;
  background: #ffffff;
}

.tasks-create-modal__fields {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tasks-create-modal__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.tasks-create-modal__label {
  font-size: 13px;
  color: #475569;
}

.tasks-create-modal__input,
.tasks-create-modal__textarea {
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

.tasks-create-modal__textarea {
  resize: vertical;
  min-height: 110px;
}

.tasks-create-modal__input:focus,
.tasks-create-modal__textarea:focus {
  outline: none;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.tasks-create-modal__date-trigger {
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
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: pointer;
}

.tasks-create-modal__date-trigger:hover {
  border-color: #93c5fd;
}
</style>
