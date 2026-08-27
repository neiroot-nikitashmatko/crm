<script setup lang="ts">
import { computed } from 'vue'
import AppModal from '@/components/common/AppModal.vue'
import type { FailedLeadListItem } from '@/types/analytics'

const DATE_FORMATTER = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
})

const show = defineModel<boolean>('show', { required: true })

defineProps<{
  leads: readonly FailedLeadListItem[]
  loading?: boolean
  errorMessage?: string
}>()

const emit = defineEmits<{
  select: [leadId: string]
}>()

function formatLeadDate(timestamp: number): string {
  if (!timestamp) return '—'
  return DATE_FORMATTER.format(new Date(timestamp))
}

function formatLeadFailureReason(lead: FailedLeadListItem): string {
  return lead.failureReason.trim() || 'Без причины'
}

function formatLeadName(lead: FailedLeadListItem): string {
  const parts = [lead.firstName.trim(), lead.patronymic.trim()].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : 'Без имени'
}

const emptyMessage = computed(() => 'Нет проваленных лидов за выбранный период')
</script>

<template>
  <AppModal v-model:show="show" title="Проваленные лиды за период" width="wide">
    <p v-if="errorMessage" class="analytics-failed-leads-modal__error">{{ errorMessage }}</p>
    <p v-else-if="loading" class="analytics-failed-leads-modal__empty">Загрузка…</p>
    <p v-else-if="leads.length === 0" class="analytics-failed-leads-modal__empty">
      {{ emptyMessage }}
    </p>
    <ul v-else class="analytics-failed-leads-modal__list">
      <li v-for="lead in leads" :key="lead.id">
        <button
          type="button"
          class="analytics-failed-leads-modal__item"
          @click="emit('select', lead.id)"
        >
          <span class="analytics-failed-leads-modal__item-row">
            <span class="analytics-failed-leads-modal__item-number">#{{ lead.leadNumber }}</span>
            <span class="analytics-failed-leads-modal__item-date">
              {{ formatLeadDate(lead.createdAt) }}
            </span>
          </span>
          <span class="analytics-failed-leads-modal__item-row">
            <span class="analytics-failed-leads-modal__item-title">
              {{ formatLeadFailureReason(lead) }}
            </span>
            <span class="analytics-failed-leads-modal__item-name">
              {{ formatLeadName(lead) }}
            </span>
          </span>
        </button>
      </li>
    </ul>
  </AppModal>
</template>

<style scoped>
.analytics-failed-leads-modal__error,
.analytics-failed-leads-modal__empty {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #64748b;
}

.analytics-failed-leads-modal__error {
  color: #cf222e;
}

.analytics-failed-leads-modal__list {
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

.analytics-failed-leads-modal__item {
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

.analytics-failed-leads-modal__item:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.analytics-failed-leads-modal__item-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.analytics-failed-leads-modal__item-number {
  font-size: 14px;
  font-weight: 700;
  color: #1a202c;
  font-variant-numeric: tabular-nums;
}

.analytics-failed-leads-modal__item-date {
  flex-shrink: 0;
  font-size: 13px;
  color: #64748b;
  font-variant-numeric: tabular-nums;
}

.analytics-failed-leads-modal__item-title {
  min-width: 0;
  flex: 1;
  font-size: 14px;
  line-height: 1.4;
  color: #334155;
  white-space: normal;
  overflow-wrap: anywhere;
}

.analytics-failed-leads-modal__item-name {
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
