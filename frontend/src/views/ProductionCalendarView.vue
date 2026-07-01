<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import ProductionCalendarSectionHeader from '@/components/production-calendar/ProductionCalendarSectionHeader.vue'
import ProductionCalendar from '@/components/production-calendar/ProductionCalendar.vue'
import DealDetailsSheet from '@/components/deals/DealDetailsSheet.vue'
import { useDeals } from '@/composables/useDeals'
import type { ProductionCalendarViewMode } from '@/constants/productionCalendar'

const viewMode = ref<ProductionCalendarViewMode>('month')
const employeeFilter = ref<string | null>(null)
const selectedDealId = ref<string | null>(null)
const { deals, loadDeals } = useDeals()

const EMPLOYEE_FILTER_STORAGE_KEY = 'proclients.productionCalendar.employeeFilter'

onMounted(() => {
  const saved = sessionStorage.getItem(EMPLOYEE_FILTER_STORAGE_KEY)
  employeeFilter.value = saved && saved.trim() ? saved : null
})

watch(employeeFilter, (value) => {
  if (!value) {
    sessionStorage.removeItem(EMPLOYEE_FILTER_STORAGE_KEY)
    return
  }
  sessionStorage.setItem(EMPLOYEE_FILTER_STORAGE_KEY, value)
})

async function handleOpenDeal(dealId: string) {
  if (!deals.value.some((deal) => deal.id === dealId)) {
    await loadDeals(true)
  }

  if (deals.value.some((deal) => deal.id === dealId)) {
    selectedDealId.value = dealId
  }
}

function handleCloseDealSheet() {
  selectedDealId.value = null
}
</script>

<template>
  <div class="production-calendar-page">
    <ProductionCalendarSectionHeader
      v-model:view-mode="viewMode"
      v-model:employee-filter="employeeFilter"
    />

    <div class="production-calendar-page__body">
      <ProductionCalendar
        :view-mode="viewMode"
        :employee-filter="employeeFilter"
        @open-deal="handleOpenDeal"
      />
    </div>

    <DealDetailsSheet :deal-id="selectedDealId" @close="handleCloseDealSheet" />
  </div>
</template>

<style scoped>
.production-calendar-page {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 64px);
  max-height: calc(100dvh - 64px);
  overflow: hidden;
  background: #ffffff;
}

.production-calendar-page__body {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;
  padding: 12px 24px 16px;
  box-sizing: border-box;
  overflow: hidden;
}
</style>
