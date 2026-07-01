<script setup lang="ts">
import TasksSectionHeader from '@/components/tasks/TasksSectionHeader.vue'
import TasksKanban from '@/components/tasks/TasksKanban.vue'
</script>

<template>
  <div class="tasks-page">
    <TasksSectionHeader />
    <TasksKanban />
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
</style>
