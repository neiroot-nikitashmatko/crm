<script setup lang="ts">
import { SearchOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'

const searchQuery = defineModel<string>('searchQuery', { default: '' })

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && searchQuery.value) {
    event.preventDefault()
    searchQuery.value = ''
  }
}
</script>

<template>
  <header class="leads-section-header">
    <h1 class="leads-section-header__title">Лиды</h1>

    <div class="leads-section-header__actions">
      <div class="leads-section-header__search-wrap">
        <input
          class="leads-section-header__search"
          type="search"
          v-model="searchQuery"
          placeholder="Поиск"
          aria-label="Поиск лидов по имени или телефону"
          @keydown="handleKeydown"
        />
        <span class="leads-section-header__search-icon" aria-hidden="true">
          <NIcon :size="16" :component="SearchOutline" />
        </span>
      </div>

      <slot name="actions" />
    </div>
  </header>
</template>

<style scoped>
.leads-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  height: 56px;
  padding: 0 16px 0 24px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}

.leads-section-header__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1a202c;
  letter-spacing: -0.02em;
}

.leads-section-header__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 32px;
}

.leads-section-header__search-wrap {
  position: relative;
  width: min(220px, 48vw);
}

.leads-section-header__search {
  width: 100%;
  height: 32px;
  padding: 0 32px 0 10px;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  box-sizing: border-box;
  font-size: 13px;
  color: #1a202c;
  outline: none;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.leads-section-header__search::placeholder {
  color: #94a3b8;
}

.leads-section-header__search:focus {
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.leads-section-header__search::-webkit-search-cancel-button {
  -webkit-appearance: none;
}

.leads-section-header__search-icon {
  position: absolute;
  top: 50%;
  right: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  transform: translateY(-50%);
  pointer-events: none;
}
</style>
