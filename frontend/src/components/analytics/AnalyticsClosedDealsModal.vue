<script setup lang="ts">
import AppModal from '@/components/common/AppModal.vue'
import { PRODUCTION_CATEGORY_COLORS } from '@/constants/analytics'
import { productionCategoryForNomenclature } from '@/constants/production'
import type { ClosedDealListItem } from '@/types/analytics'

const DATE_FORMATTER = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
})

const show = defineModel<boolean>('show', { required: true })

defineProps<{
  deals: readonly ClosedDealListItem[]
  loading?: boolean
  errorMessage?: string
  detail?: 'nomenclature' | 'employee'
}>()

const emit = defineEmits<{
  select: [dealId: string]
}>()

function formatDealDate(timestamp: number): string {
  if (!timestamp) return '—'
  return DATE_FORMATTER.format(new Date(timestamp))
}

function formatDealNomenclature(deal: ClosedDealListItem): string {
  return deal.nomenclature.trim() || 'Без номенклатуры'
}

function formatDealCategory(deal: ClosedDealListItem): string {
  return deal.category.trim() || productionCategoryForNomenclature(deal.nomenclature)
}

function formatDealEmployee(deal: ClosedDealListItem): string {
  return deal.employee.trim() || 'Без сотрудника'
}

function categoryColor(category: string): string {
  return PRODUCTION_CATEGORY_COLORS[category] ?? '#94a3b8'
}
</script>

<template>
  <AppModal v-model:show="show" title="Закрытые сделки за период" width="wide">
    <p v-if="errorMessage" class="analytics-closed-deals-modal__error">{{ errorMessage }}</p>
    <p v-else-if="loading" class="analytics-closed-deals-modal__empty">Загрузка…</p>
    <p v-else-if="deals.length === 0" class="analytics-closed-deals-modal__empty">
      Нет закрытых сделок за выбранный период
    </p>
    <ul v-else class="analytics-closed-deals-modal__list">
      <li v-for="deal in deals" :key="deal.id">
        <button
          type="button"
          class="analytics-closed-deals-modal__item"
          @click="emit('select', deal.id)"
        >
          <span class="analytics-closed-deals-modal__item-row">
            <span class="analytics-closed-deals-modal__item-number">#{{ deal.dealNumber }}</span>
            <span class="analytics-closed-deals-modal__item-date">
              {{ formatDealDate(deal.closedAt || deal.createdAt) }}
            </span>
          </span>
          <span class="analytics-closed-deals-modal__item-row">
            <span class="analytics-closed-deals-modal__item-title">
              {{ detail === 'employee' ? formatDealEmployee(deal) : formatDealNomenclature(deal) }}
            </span>
            <span
              v-if="detail !== 'employee'"
              class="analytics-closed-deals-modal__category"
              :style="{ '--category-color': categoryColor(formatDealCategory(deal)) }"
            >
              {{ formatDealCategory(deal) }}
            </span>
          </span>
        </button>
      </li>
    </ul>
  </AppModal>
</template>

<style scoped>
.analytics-closed-deals-modal__error,
.analytics-closed-deals-modal__empty {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #64748b;
}

.analytics-closed-deals-modal__error {
  color: #cf222e;
}

.analytics-closed-deals-modal__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: min(60vh, 480px);
  margin: 0 -4px;
  padding: 0 4px;
  overflow: auto;
  list-style: none;
  scrollbar-gutter: stable;
}

.analytics-closed-deals-modal__item {
  display: flex;
  width: 100%;
  flex-direction: column;
  gap: 8px;
  padding: 12px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #ffffff;
  text-align: left;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease;
}

.analytics-closed-deals-modal__item:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.analytics-closed-deals-modal__item-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-width: 0;
}

.analytics-closed-deals-modal__item-number {
  font-size: 12px;
  font-weight: 600;
  line-height: 1.2;
  color: #4a5568;
  font-variant-numeric: tabular-nums;
}

.analytics-closed-deals-modal__item-date {
  flex-shrink: 0;
  font-size: 11px;
  line-height: 1.2;
  color: #718096;
  font-variant-numeric: tabular-nums;
}

.analytics-closed-deals-modal__item-title {
  min-width: 0;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.35;
  color: #1a202c;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.analytics-closed-deals-modal__category {
  flex-shrink: 0;
  padding: 2px 8px;
  border: 1px solid color-mix(in srgb, var(--category-color) 22%, #ffffff);
  border-radius: 6px;
  background: color-mix(in srgb, var(--category-color) 12%, #ffffff);
  color: var(--category-color);
  font-size: 11px;
  font-weight: 600;
  line-height: 1.4;
}
</style>
