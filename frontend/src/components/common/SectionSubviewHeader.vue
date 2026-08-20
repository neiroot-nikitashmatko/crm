<script setup lang="ts">
defineProps<{
  title: string
}>()
</script>

<template>
  <header class="section-subview-header">
    <h2 class="section-subview-header__title">{{ title }}</h2>
    <div class="section-subview-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<style scoped>
.section-subview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  gap: 16px;
  min-height: 56px;
  padding: 8px 16px 8px 24px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  box-sizing: border-box;
}

.section-subview-header__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.section-subview-header__actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
  min-height: 32px;
}
</style>
