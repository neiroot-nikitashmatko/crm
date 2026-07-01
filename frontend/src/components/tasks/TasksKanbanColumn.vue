<script setup lang="ts">
import { computed } from 'vue'
import type { TaskKanbanColumn } from '@/constants/tasks'
import type { Task } from '@/types/task'
import TaskCard from './TaskCard.vue'

const props = defineProps<{
  column: TaskKanbanColumn
  tasks: Task[]
}>()

const emit = defineEmits<{
  openTask: [task: Task]
}>()

const headerStyle = computed(() => ({
  backgroundColor: props.column.style.headerBg,
  borderBottomColor: props.column.style.headerBorder,
}))

const countStyle = computed(() => ({
  backgroundColor: props.column.style.countBg,
  color: props.column.style.countColor,
}))
</script>

<template>
  <section class="tasks-kanban-column">
    <header class="tasks-kanban-column__header" :style="headerStyle">
      <h2 class="tasks-kanban-column__title">{{ column.title }}</h2>
      <span class="tasks-kanban-column__count" :style="countStyle">{{ tasks.length }}</span>
    </header>

    <div class="tasks-kanban-column__scroll">
      <div class="tasks-kanban-column__cards">
        <TaskCard
          v-for="task in tasks"
          :key="task.id"
          :task="task"
          :column-id="column.id"
          @open="emit('openTask', task)"
        />
      </div>
    </div>
  </section>
</template>

<style scoped>
.tasks-kanban-column {
  display: flex;
  flex-direction: column;
  flex: 0 0 260px;
  min-width: 260px;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  box-sizing: border-box;
  background: #f6f8fa;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.tasks-kanban-column__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 12px 14px;
  border-bottom: 1px solid;
}

.tasks-kanban-column__title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #1a202c;
  line-height: 1.3;
}

.tasks-kanban-column__count {
  flex-shrink: 0;
  min-width: 22px;
  padding: 2px 8px;
  border-radius: 10px;
  text-align: center;
  font-size: 12px;
  font-weight: 600;
}

.tasks-kanban-column__scroll {
  flex: 1 1 auto;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
}

.tasks-kanban-column__cards {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px;
}
</style>
