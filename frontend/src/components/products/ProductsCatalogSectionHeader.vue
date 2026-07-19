<script setup lang="ts">
import { AddOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'

const emit = defineEmits<{
  addProduct: []
}>()

function handleAddProductClick() {
  emit('addProduct')
}
</script>

<template>
  <header class="products-catalog-section-header">
    <h1 class="products-catalog-section-header__title">Каталог товаров</h1>
    <div class="products-catalog-section-header__actions">
      <button
        type="button"
        class="products-catalog-section-header__create-btn"
        title="Добавить товар"
        aria-label="Добавить товар"
        @click="handleAddProductClick"
      >
        <NIcon :size="18" :component="AddOutline" />
      </button>
      <slot name="actions" />
    </div>
  </header>
</template>

<style scoped>
.products-catalog-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  height: 56px;
  padding: 0 16px 0 24px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}

.products-catalog-section-header__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1a202c;
  letter-spacing: -0.02em;
}

.products-catalog-section-header__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 32px;
}

.products-catalog-section-header__create-btn {
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

.products-catalog-section-header__create-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1f2937;
}
</style>
