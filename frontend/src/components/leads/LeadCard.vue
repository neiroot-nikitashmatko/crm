<script setup lang="ts">
import { computed } from 'vue'
import { LEAD_KANBAN_COLUMNS } from '@/constants/leads'
import type { Lead } from '@/types/lead'

const props = defineProps<{
  lead: Lead
}>()

const emit = defineEmits<{
  open: []
}>()

/**
 * Быстрый переключатель визуальных вариантов карточки:
 * A — спокойный и деловой
 * B — акцентный с бейджами
 * C — структурный CRM-вид
 */
const CARD_PRESET: 'A' | 'B' | 'C' = 'B'

const displayName = computed(() => {
  const parts = [props.lead.firstName.trim(), props.lead.patronymic.trim()].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : 'Без имени'
})

const displayPhone = computed(() => props.lead.phone.trim() || '—')
const displayTrafficSource = computed(() => props.lead.trafficSource.trim() || '—')
const displayCreatedAt = computed(() =>
  new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(props.lead.createdAt),
)
const presetClass = computed(() => `lead-card--preset-${CARD_PRESET.toLowerCase()}`)
const statusColorByColumnId = new Map(
  LEAD_KANBAN_COLUMNS.map((column) => [column.id, column.style.countColor]),
)
const cardStyle = computed(() => ({
  '--lead-accent-color': statusColorByColumnId.get(props.lead.columnId) ?? '#4a6fa5',
}))
</script>

<template>
  <article class="lead-card" :class="presetClass" :style="cardStyle" @click="emit('open')">
    <div class="lead-card__meta">
      <p class="lead-card__number">Лид #{{ lead.leadNumber }}</p>
      <p class="lead-card__created-at">{{ displayCreatedAt }}</p>
    </div>

    <p class="lead-card__name">{{ displayName }}</p>

    <p class="lead-card__phone">{{ displayPhone }}</p>

    <div class="lead-card__row">
      <span class="lead-card__label">Источник</span>
      <p class="lead-card__value lead-card__traffic-source">{{ displayTrafficSource }}</p>
    </div>
  </article>
</template>

<style scoped>
.lead-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;
  min-height: 120px;
  padding: 12px 12px 10px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  box-sizing: border-box;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease,
    transform 0.15s ease;
}

.lead-card:hover {
  border-color: #cbd5e0;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.08);
  transform: translateY(-1px);
}

.lead-card__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.lead-card__name {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #1a202c;
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.lead-card__number {
  margin: 0;
  font-size: 12px;
  font-weight: 600;
  color: #4a5568;
  line-height: 1.2;
}

.lead-card__created-at {
  margin: 0;
  font-size: 11px;
  color: #718096;
  line-height: 1.2;
}

.lead-card__row {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.lead-card__label {
  flex-shrink: 0;
  font-size: 11px;
  font-weight: 600;
  color: #718096;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.lead-card__value {
  margin: 0;
  min-width: 0;
  font-size: 13px;
  font-weight: 500;
  color: #4a5568;
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.lead-card__phone {
  margin: 0;
  font-size: 13px;
  font-weight: 500;
  color: #4a5568;
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.lead-card__traffic-source {
  color: #3d546c;
}

/* Preset A — спокойный и деловой */
.lead-card--preset-a {
  border-color: #e2e8f0;
}

/* Preset B — акцентный с бейджем источника */
.lead-card--preset-b {
  border-color: #d9e3f0;
}

.lead-card--preset-b:hover {
  box-shadow:
    inset 3px 0 0 var(--lead-accent-color),
    0 4px 12px rgba(15, 23, 42, 0.08);
}

.lead-card--preset-b .lead-card__traffic-source {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 999px;
  background: #edf2f7;
  color: #2d3748;
  font-size: 12px;
  font-weight: 600;
}

.lead-card--preset-b .lead-card__row:last-child .lead-card__label {
  display: none;
}

/* Preset C — более структурный CRM */
.lead-card--preset-c {
  border-color: #d7dee8;
  background: linear-gradient(180deg, #ffffff 0%, #fbfdff 100%);
}

.lead-card--preset-c .lead-card__name {
  font-size: 14px;
}

.lead-card--preset-c .lead-card__row {
  display: grid;
  grid-template-columns: 74px minmax(0, 1fr);
  gap: 8px;
  align-items: center;
}

.lead-card--preset-c .lead-card__label {
  text-transform: none;
  letter-spacing: 0;
  font-size: 12px;
  font-weight: 500;
}
</style>
