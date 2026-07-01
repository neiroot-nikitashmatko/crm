<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { TASK_KANBAN_COLUMNS } from '@/constants/tasks'
import { getLeadsKanbanHeight } from '@/constants/layout'
import { useTasks } from '@/composables/useTasks'
import type { Task, TaskKanbanColumnId } from '@/types/task'
import TaskDetailsSheet from './TaskDetailsSheet.vue'
import TasksKanbanColumn from './TasksKanbanColumn.vue'

const { tasks, loadTasks } = useTasks()
const selectedTaskId = ref<string | null>(null)
const kanbanHeightPx = ref(getLeadsKanbanHeight())
const resizeHandler = () => {
  kanbanHeightPx.value = getLeadsKanbanHeight()
}

onMounted(() => {
  window.addEventListener('resize', resizeHandler)
  resizeHandler()
  void loadTasks()
})

onUnmounted(() => {
  window.removeEventListener('resize', resizeHandler)
})

function resolveColumnId(task: Task): TaskKanbanColumnId {
  if (!task.dueAt) return 'no-deadline'
  const due = new Date(task.dueAt)
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const tomorrow = new Date(today)
  tomorrow.setDate(today.getDate() + 1)
  const afterTomorrow = new Date(today)
  afterTomorrow.setDate(today.getDate() + 2)
  const dueDay = new Date(due.getFullYear(), due.getMonth(), due.getDate())

  if (dueDay.getTime() < today.getTime()) return 'overdue'
  if (dueDay.getTime() === today.getTime()) return 'today'
  if (dueDay.getTime() === tomorrow.getTime()) return 'tomorrow'
  return 'later'
}

const groupedTasks = computed(() => {
  const map: Record<TaskKanbanColumnId, Task[]> = {
    overdue: [],
    today: [],
    tomorrow: [],
    later: [],
    'no-deadline': [],
  }
  for (const task of tasks.value) {
    if (task.status === 'completed') continue
    map[resolveColumnId(task)].push(task)
  }
  return map
})
</script>

<template>
  <section
    class="tasks-kanban"
    :style="{ height: `${kanbanHeightPx}px`, maxHeight: `${kanbanHeightPx}px` }"
  >
    <div class="tasks-kanban__viewport">
      <div class="tasks-kanban__track">
        <TasksKanbanColumn
          v-for="column in TASK_KANBAN_COLUMNS"
          :key="column.id"
          :column="column"
          :tasks="groupedTasks[column.id]"
          @open-task="selectedTaskId = $event.id"
        />
      </div>
    </div>

    <TaskDetailsSheet :task-id="selectedTaskId" @close="selectedTaskId = null" />
  </section>
</template>

<style scoped>
.tasks-kanban {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #ffffff;
}

.tasks-kanban__viewport {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  overflow-x: auto;
  overflow-y: hidden;
  padding-top: 16px;
  box-sizing: border-box;
}

.tasks-kanban__track {
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
</style>
