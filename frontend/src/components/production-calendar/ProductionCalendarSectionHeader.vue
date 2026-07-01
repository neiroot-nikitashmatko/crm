<script setup lang="ts">
import { computed, ref } from 'vue'
import { NIcon, NPopover, NSelect } from 'naive-ui'
import { FunnelOutline } from '@vicons/ionicons5'
import {
  PRODUCTION_CALENDAR_VIEW_OPTIONS,
  type ProductionCalendarViewMode,
} from '@/constants/productionCalendar'

const viewMode = defineModel<ProductionCalendarViewMode>('viewMode', { required: true })
const employeeFilter = defineModel<string | null>('employeeFilter', { required: true })

const isFilterPopoverOpen = ref(false)

const ALL_EMPLOYEES = '__all__'

const employeeOptions = computed(() => [
  { label: 'Все сотрудники', value: ALL_EMPLOYEES },
  { label: 'Никита Хачересов', value: 'Никита Хачересов' },
  { label: 'Сергей Геворкян', value: 'Сергей Геворкян' },
])

const filterButtonTitle = computed(() =>
  employeeFilter.value ? `Фильтр: ${employeeFilter.value}` : 'Фильтр: все сотрудники',
)

const employeeSelectValue = computed<string>({
  get: () => employeeFilter.value ?? ALL_EMPLOYEES,
  set: (value) => {
    employeeFilter.value = value === ALL_EMPLOYEES ? null : value
  },
})
</script>

<template>
  <header class="production-calendar-section-header">
    <h1 class="production-calendar-section-header__title">Календарь производства</h1>

    <div class="production-calendar-section-header__right">
      <NPopover
        v-model:show="isFilterPopoverOpen"
        trigger="click"
        placement="bottom-end"
        :show-arrow="true"
      >
        <template #trigger>
          <button
            type="button"
            class="production-calendar-section-header__filter-btn"
            :title="filterButtonTitle"
          >
            <NIcon :size="18">
              <FunnelOutline />
            </NIcon>
          </button>
        </template>

        <div class="production-calendar-section-header__filter-popover">
          <div class="production-calendar-section-header__filter-title">Фильтр по сотруднику</div>
          <NSelect
            v-model:value="employeeSelectValue"
            :options="employeeOptions"
            placeholder="Все сотрудники"
            @update:value="isFilterPopoverOpen = false"
          />
        </div>
      </NPopover>

      <div
        class="production-calendar-section-header__view-switch"
        role="group"
        aria-label="Вид календаря"
      >
        <button
          v-for="option in PRODUCTION_CALENDAR_VIEW_OPTIONS"
          :key="option.value"
          type="button"
          class="production-calendar-section-header__view-btn"
          :class="{ 'production-calendar-section-header__view-btn--active': viewMode === option.value }"
          :aria-pressed="viewMode === option.value"
          @click="viewMode = option.value"
        >
          {{ option.label }}
        </button>
      </div>
    </div>
  </header>
</template>

<style scoped>
.production-calendar-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  height: 56px;
  padding: 0 24px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
}

.production-calendar-section-header__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1a202c;
  letter-spacing: -0.02em;
}

.production-calendar-section-header__right {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.production-calendar-section-header__filter-btn {
  width: 34px;
  height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #ffffff;
  color: #334155;
  cursor: pointer;
  transition: background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease;
}

.production-calendar-section-header__filter-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1a202c;
}

.production-calendar-section-header__filter-popover {
  width: 240px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.production-calendar-section-header__filter-title {
  font-size: 12px;
  font-weight: 700;
  color: #334155;
}

.production-calendar-section-header__view-switch {
  display: inline-flex;
  align-items: center;
  padding: 3px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
}

.production-calendar-section-header__view-btn {
  padding: 6px 12px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: #64748b;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease, box-shadow 0.15s ease;
}

.production-calendar-section-header__view-btn:hover {
  color: #334155;
}

.production-calendar-section-header__view-btn--active {
  background: #ffffff;
  color: #1a202c;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.08);
}
</style>
