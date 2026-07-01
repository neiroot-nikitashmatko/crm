<script setup lang="ts">
import { computed } from 'vue'
import { DEAL_KANBAN_COLUMNS } from '@/constants/deals'
import { resolveDealKanbanColumnId } from '@/utils/dealKanban'
import type { Deal } from '@/types/deal'

const props = defineProps<{
  deal: Deal
}>()

const emit = defineEmits<{
  open: []
}>()

const displayName = computed(() => {
  const parts = [props.deal.firstName?.trim() ?? '', props.deal.patronymic?.trim() ?? ''].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : 'Без имени'
})

const displayPhone = computed(() => props.deal.phone?.trim() || '—')
const displayAmount = computed(() => `${Number(props.deal.totalAmount || 0)} ₽`)
const displayCreatedAt = computed(() =>
  new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(new Date(props.deal.createdAt)),
)

const accentMap = new Map(DEAL_KANBAN_COLUMNS.map((column) => [column.id, column.style.countColor]))
const cardStyle = computed(() => ({
  '--deal-accent-color': accentMap.get(resolveDealKanbanColumnId(props.deal)) ?? '#4a5568',
}))
</script>

<template>
  <article class="deal-card" :style="cardStyle" @click="emit('open')">
    <div class="deal-card__meta">
      <p class="deal-card__number">Сделка #{{ deal.dealNumber }}</p>
      <p class="deal-card__created-at">{{ displayCreatedAt }}</p>
    </div>

    <p class="deal-card__name">{{ displayName }}</p>
    <p class="deal-card__phone">{{ displayPhone }}</p>
    <p class="deal-card__amount">{{ displayAmount }}</p>
  </article>
</template>

<style scoped>
.deal-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 112px;
  padding: 12px 12px 10px;
  border: 1px solid #d9e3f0;
  border-radius: 10px;
  background: #ffffff;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease,
    transform 0.15s ease;
}

.deal-card:hover {
  border-color: #cbd5e0;
  transform: translateY(-1px);
  box-shadow:
    inset 3px 0 0 var(--deal-accent-color),
    0 4px 12px rgba(15, 23, 42, 0.08);
}

.deal-card__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.deal-card__number {
  margin: 0;
  font-size: 12px;
  font-weight: 600;
  color: #4a5568;
}

.deal-card__created-at {
  margin: 0;
  font-size: 11px;
  color: #718096;
}

.deal-card__name {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #1a202c;
  line-height: 1.3;
}

.deal-card__phone,
.deal-card__amount {
  margin: 0;
  font-size: 13px;
  font-weight: 500;
  color: #4a5568;
  line-height: 1.3;
}
</style>
