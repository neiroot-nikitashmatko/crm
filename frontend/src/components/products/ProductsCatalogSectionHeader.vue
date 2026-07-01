<script setup lang="ts">
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
        class="products-catalog-section-header__action"
        aria-label="Добавить товар"
        @click="handleAddProductClick"
      >
        Добавить товар
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
  padding: 0 24px;
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

.products-catalog-section-header__action {
  padding: 8px 14px;
  border: 1px solid #1f883d;
  border-radius: 8px;
  background: #1f883d;
  color: #ffffff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.products-catalog-section-header__action:hover {
  background: #197a35;
}
</style>
