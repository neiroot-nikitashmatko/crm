<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  PeopleOutline,
  BriefcaseOutline,
  CheckboxOutline,
  CubeOutline,
  CalendarOutline,
  CashOutline,
  IdCardOutline,
} from '@vicons/ionicons5'
import { useAuth } from '@/composables/useAuth'

defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const route = useRoute()
const router = useRouter()
const { isAdmin, canAccessSection } = useAuth()

const menuItems = [
  { label: 'Лиды', name: 'leads', sectionName: 'leads', icon: PeopleOutline, adminOnly: false },
  { label: 'Сделки', name: 'deals', sectionName: 'deals', icon: BriefcaseOutline, adminOnly: false },
  { label: 'Задачи', name: 'tasks', sectionName: 'tasks', icon: CheckboxOutline, adminOnly: false },
  { label: 'Каталог товаров', name: 'products-catalog', sectionName: 'products-catalog', icon: CubeOutline, adminOnly: false },
  {
    label: 'Календарь производства',
    name: 'production-calendar',
    sectionName: 'production-calendar',
    icon: CalendarOutline,
    adminOnly: false,
  },
  { label: 'Зарплата', name: 'salary-add-deal', sectionName: 'salary', icon: CashOutline, adminOnly: false },
  { label: 'Сотрудники', name: 'employees-list', sectionName: 'employees', icon: IdCardOutline, adminOnly: true },
]

const visibleMenuItems = computed(() =>
  menuItems.filter((item) => {
    if (item.adminOnly) return isAdmin.value
    return canAccessSection(item.sectionName)
  }),
)

const activeName = computed(() => {
  const name = route.name as string
  if (name === 'employees-list' || name === 'employees-new') {
    return 'employees-list'
  }
  if (name === 'salary-add-deal' || name === 'salary-report') {
    return 'salary-add-deal'
  }

  return name
})

function navigate(name: string) {
  router.push({ name })
  emit('close')
}
</script>

<template>
  <aside class="app-sidebar" :class="{ 'app-sidebar--open': open }">
    <nav class="app-sidebar__nav">
      <button
        v-for="item in visibleMenuItems"
        :key="item.name"
        type="button"
        class="app-sidebar__item"
        :class="{ 'app-sidebar__item--active': activeName === item.name }"
        @click="navigate(item.name)"
      >
        <component :is="item.icon" class="app-sidebar__icon" />
        <span>{{ item.label }}</span>
      </button>
    </nav>
  </aside>
</template>

<style scoped>
.app-sidebar {
  position: fixed;
  top: 64px;
  left: 0;
  bottom: 0;
  width: 280px;
  background: #ffffff;
  border-right: 1px solid #e2e8f0;
  border-top-right-radius: 12px;
  border-bottom-right-radius: 12px;
  transform: translateX(-100%);
  transition: transform 0.25s ease;
  z-index: 95;
  overflow-x: hidden;
  overflow-y: auto;
}

.app-sidebar--open {
  transform: translateX(0);
}

.app-sidebar__nav {
  display: flex;
  flex-direction: column;
  padding: 12px 8px;
  gap: 2px;
}

.app-sidebar__item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 10px 12px 10px 14px;
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

.app-sidebar__item:hover {
  background: #edf2f7;
  color: #1a202c;
}

.app-sidebar__item--active {
  background: #eef7f0;
  color: #14532d;
  font-weight: 600;
}

.app-sidebar__item--active::before {
  content: '';
  position: absolute;
  top: 8px;
  bottom: 8px;
  left: 0;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: #1f883d;
}

.app-sidebar__icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.app-sidebar__item--active .app-sidebar__icon {
  color: #1f883d;
}
</style>
