<script setup lang="ts">
import { AddOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'

defineEmits<{
  'create-task': []
}>()
</script>

<template>
  <header class="tasks-section-header">
    <h1 class="tasks-section-header__title">Задачи</h1>
    <button
      type="button"
      class="tasks-section-header__create-btn"
      title="Создать задачу"
      aria-label="Создать задачу"
      @click="$emit('create-task')"
    >
      <NIcon :size="18" :component="AddOutline" />
    </button>
  </header>
</template>

<style scoped>
.tasks-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  height: 56px;
  padding: 0 16px 0 24px;
  border-bottom: 1px solid #e2e8f0;
  background: #ffffff;
}

.tasks-section-header__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1a202c;
  letter-spacing: -0.02em;
}

.tasks-section-header__create-btn {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #475569;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.tasks-section-header__create-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1f2937;
}
</style>
