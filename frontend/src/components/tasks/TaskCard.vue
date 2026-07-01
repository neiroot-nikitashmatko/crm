<script setup lang="ts">
import { computed } from 'vue'
import { TASK_KANBAN_COLUMNS } from '@/constants/tasks'
import type { Task, TaskKanbanColumnId } from '@/types/task'

const props = defineProps<{
  task: Task
  columnId: TaskKanbanColumnId
}>()

const emit = defineEmits<{
  open: []
}>()

const displayDueAt = computed(() =>
  props.task.dueAt ? new Date(props.task.dueAt).toLocaleString('ru-RU') : 'Без срока',
)

const accentMap = new Map(TASK_KANBAN_COLUMNS.map((column) => [column.id, column.style.countColor]))
const cardStyle = computed(() => ({
  '--task-accent-color': accentMap.get(props.columnId) ?? '#4a5568',
}))
</script>

<template>
  <article class="task-card" :style="cardStyle" @click="emit('open')">
    <p class="task-card__title">{{ task.title || 'Без названия' }}</p>
    <p class="task-card__meta">{{ displayDueAt }}</p>
  </article>
</template>

<style scoped>
.task-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-height: 74px;
  padding: 10px 11px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #ffffff;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease,
    transform 0.15s ease;
}

.task-card:hover {
  border-color: #cbd5e0;
  transform: translateY(-1px);
  box-shadow:
    inset 3px 0 0 var(--task-accent-color),
    0 4px 12px rgba(15, 23, 42, 0.08);
}

.task-card__title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #1a202c;
  line-height: 1.3;
}

.task-card__meta {
  margin: 0;
  font-size: 12px;
  color: #718096;
  line-height: 1.2;
}
</style>
