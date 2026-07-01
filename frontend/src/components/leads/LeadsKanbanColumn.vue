<script setup lang="ts">
import { computed, ref } from 'vue'
import type { LeadKanbanColumnStyle } from '@/constants/leads'
import type { Lead, NewLeadForm } from '@/types/lead'
import LeadCard from './LeadCard.vue'
import LeadsAddLeadPanel from './LeadsAddLeadPanel.vue'

const props = defineProps<{
  title: string
  count?: number
  columnStyle: LeadKanbanColumnStyle
  showAddLeadButton?: boolean
  leads: Lead[]
}>()

const emit = defineEmits<{
  addLead: [payload: NewLeadForm]
  layoutChange: []
  openLead: [lead: Lead]
}>()

const addLeadPanelRef = ref<InstanceType<typeof LeadsAddLeadPanel> | null>(null)

const headerStyle = computed(() => ({
  backgroundColor: props.columnStyle.headerBg,
  borderBottomColor: props.columnStyle.headerBorder,
}))

const countStyle = computed(() => ({
  backgroundColor: props.columnStyle.countBg,
  color: props.columnStyle.countColor,
}))

function handlePanelLayoutChange() {
  emit('layoutChange')
}

function handleAddLead(payload: NewLeadForm) {
  emit('addLead', payload)
}

function openAddLeadForm() {
  addLeadPanelRef.value?.openForm()
}

function handleOpenLead(lead: Lead) {
  emit('openLead', lead)
}
</script>

<template>
  <section class="leads-kanban-column">
    <div class="leads-kanban-column__top">
      <header class="leads-kanban-column__header" :style="headerStyle">
        <h2 class="leads-kanban-column__title">{{ title }}</h2>
        <span class="leads-kanban-column__count" :style="countStyle">{{ count ?? 0 }}</span>
      </header>

      <div v-if="showAddLeadButton" class="leads-kanban-column__toolbar">
        <button
          type="button"
          class="leads-kanban-column__add-btn"
          @click="openAddLeadForm"
        >
          Добавить лид
        </button>
      </div>
    </div>

    <div class="leads-kanban-column__scroll">
      <LeadsAddLeadPanel
        v-if="showAddLeadButton"
        ref="addLeadPanelRef"
        :show-trigger="false"
        class="leads-kanban-column__form"
        @save="handleAddLead"
        @layout-change="handlePanelLayoutChange"
      />

      <div class="leads-kanban-column__cards">
        <LeadCard
          v-for="lead in leads"
          :key="lead.id"
          :lead="lead"
          @open="handleOpenLead(lead)"
        />
      </div>
    </div>
  </section>
</template>

<style scoped>
.leads-kanban-column {
  display: flex;
  flex-direction: column;
  flex: 0 0 260px;
  min-width: 260px;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  box-sizing: border-box;
  background: #f6f8fa;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.leads-kanban-column__top {
  flex-shrink: 0;
}

.leads-kanban-column__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 12px 14px;
  border-bottom: 1px solid;
}

.leads-kanban-column__title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.3;
  color: #1a202c;
}

.leads-kanban-column__count {
  flex-shrink: 0;
  min-width: 22px;
  padding: 2px 8px;
  font-size: 12px;
  font-weight: 600;
  text-align: center;
  border-radius: 10px;
}

.leads-kanban-column__toolbar {
  padding: 10px;
}

.leads-kanban-column__add-btn {
  display: block;
  width: 100%;
  padding: 7px 12px;
  border: 1px dashed #c5d9f0;
  border-radius: 6px;
  background: #ffffff;
  font-size: 14px;
  font-weight: 500;
  color: #4a5568;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
  -webkit-tap-highlight-color: transparent;
}

.leads-kanban-column__add-btn:hover {
  background: #f0f6fd;
  border-color: #a8c4e8;
  color: #2c5282;
}

.leads-kanban-column__add-btn:active {
  background: #e8f1fc;
}

.leads-kanban-column__scroll {
  flex: 1 1 auto;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior-y: contain;
  -webkit-overflow-scrolling: touch;
  touch-action: pan-y;
}

.leads-kanban-column__scroll::-webkit-scrollbar {
  width: 8px;
}

.leads-kanban-column__scroll::-webkit-scrollbar-track {
  background: transparent;
}

.leads-kanban-column__scroll::-webkit-scrollbar-thumb {
  background: #cbd5e0;
  border-radius: 4px;
}

.leads-kanban-column__scroll::-webkit-scrollbar-thumb:hover {
  background: #a0aec0;
}

.leads-kanban-column__form {
  padding: 10px 10px 0;
}

.leads-kanban-column__cards {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px;
  box-sizing: border-box;
}
</style>
