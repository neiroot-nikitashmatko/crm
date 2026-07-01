<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ListOutline, PersonAddOutline } from '@vicons/ionicons5'

const route = useRoute()
const router = useRouter()

const navItems = [
  { label: 'Список сотрудников', name: 'employees-list', icon: ListOutline },
  { label: 'Добавить сотрудника', name: 'employees-new', icon: PersonAddOutline },
]

const activeName = computed(() => route.name as string)

function navigate(name: string) {
  router.push({ name })
}
</script>

<template>
  <nav class="employees-nav" aria-label="Раздел сотрудников">
    <button
      v-for="item in navItems"
      :key="item.name"
      type="button"
      class="employees-nav__item"
      :class="{ 'employees-nav__item--active': activeName === item.name }"
      @click="navigate(item.name)"
    >
      <component :is="item.icon" class="employees-nav__icon" />
      <span>{{ item.label }}</span>
    </button>
  </nav>
</template>

<style scoped>
.employees-nav {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  width: 240px;
  padding: 12px 8px;
  gap: 2px;
  border-right: 1px solid #e2e8f0;
  background: #ffffff;
  overflow-y: auto;
}

.employees-nav__item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 12px;
  border: none;
  border-radius: 8px;
  background: transparent;
  font-size: 14px;
  color: #4a5568;
  cursor: pointer;
  text-align: left;
  transition:
    background 0.15s ease,
    color 0.15s ease;
}

.employees-nav__item:hover {
  background: #edf2f7;
  color: #1a202c;
}

.employees-nav__item--active {
  background: #edf2f7;
  color: #1a202c;
  font-weight: 500;
}

.employees-nav__icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}
</style>
