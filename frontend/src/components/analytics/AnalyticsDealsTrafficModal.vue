<script setup lang="ts">
import AppModal from '@/components/common/AppModal.vue'
import type { DealTrafficListItem } from '@/types/analytics'

const DATE_FORMATTER = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
})

const show = defineModel<boolean>('show', { required: true })

defineProps<{
  deals: readonly DealTrafficListItem[]
  loading?: boolean
  errorMessage?: string
}>()

const emit = defineEmits<{
  select: [dealId: string]
}>()

function formatDealDate(timestamp: number): string {
  if (!timestamp) return '—'
  return DATE_FORMATTER.format(new Date(timestamp))
}

function formatDealTrafficSource(deal: DealTrafficListItem): string {
  return deal.trafficSource.trim() || 'Без источника'
}

function formatDealName(deal: DealTrafficListItem): string {
  const parts = [deal.firstName.trim(), deal.patronymic.trim()].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : 'Без имени'
}
</script>

<template>
  <AppModal v-model:show="show" title="Сделки за период" width="wide">
    <p v-if="errorMessage" class="analytics-deals-traffic-modal__error">{{ errorMessage }}</p>
    <p v-else-if="loading" class="analytics-deals-traffic-modal__empty">Загрузка…</p>
    <p v-else-if="deals.length === 0" class="analytics-deals-traffic-modal__empty">
      За выбранный период нет сделок
    </p>
    <ul v-else class="analytics-deals-traffic-modal__list">
      <li v-for="deal in deals" :key="deal.id">
        <button
          type="button"
          class="analytics-deals-traffic-modal__item"
          @click="emit('select', deal.id)"
        >
          <span class="analytics-deals-traffic-modal__item-row">
            <span class="analytics-deals-traffic-modal__item-number">#{{ deal.dealNumber }}</span>
            <span class="analytics-deals-traffic-modal__item-date">
              {{ formatDealDate(deal.createdAt) }}
            </span>
          </span>
          <span class="analytics-deals-traffic-modal__item-row">
            <span class="analytics-deals-traffic-modal__item-title">
              {{ formatDealTrafficSource(deal) }}
            </span>
            <span class="analytics-deals-traffic-modal__item-name">
              {{ formatDealName(deal) }}
            </span>
          </span>
        </button>
      </li>
    </ul>
  </AppModal>
</template>

<style scoped>
.analytics-deals-traffic-modal__error,
.analytics-deals-traffic-modal__empty {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #64748b;
}

.analytics-deals-traffic-modal__error {
  color: #cf222e;
}

.analytics-deals-traffic-modal__list {
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

.analytics-deals-traffic-modal__item {
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

.analytics-deals-traffic-modal__item:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.analytics-deals-traffic-modal__item-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.analytics-deals-traffic-modal__item-number {
  font-size: 14px;
  font-weight: 700;
  color: #1a202c;
  font-variant-numeric: tabular-nums;
}

.analytics-deals-traffic-modal__item-date {
  flex-shrink: 0;
  font-size: 13px;
  color: #64748b;
  font-variant-numeric: tabular-nums;
}

.analytics-deals-traffic-modal__item-title {
  min-width: 0;
  flex: 1;
  font-size: 14px;
  line-height: 1.4;
  color: #334155;
  white-space: normal;
  overflow-wrap: anywhere;
}

.analytics-deals-traffic-modal__item-name {
  flex-shrink: 0;
  max-width: 40%;
  font-size: 13px;
  color: #64748b;
  text-align: right;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
