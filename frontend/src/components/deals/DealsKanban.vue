<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DEAL_KANBAN_COLUMNS } from '@/constants/deals'
import { useDeals } from '@/composables/useDeals'
import { useTasks } from '@/composables/useTasks'
import { resolveDealKanbanColumnId } from '@/utils/dealKanban'
import type { Deal, DealKanbanColumnId } from '@/types/deal'
import { getLeadsKanbanHeight } from '@/constants/layout'
import DealDetailsSheet from './DealDetailsSheet.vue'
import DealsKanbanColumn from './DealsKanbanColumn.vue'

const route = useRoute()
const router = useRouter()
const { deals, loadDeals } = useDeals()
const { loadTasks } = useTasks()
const selectedDealId = ref<string | null>(null)
const kanbanHeightPx = ref(getLeadsKanbanHeight())
const resizeHandler = () => {
  kanbanHeightPx.value = getLeadsKanbanHeight()
}

async function openDealFromRouteQuery() {
  const rawDealId = route.query.dealId
  if (typeof rawDealId !== 'string' || rawDealId.trim() === '') {
    return
  }

  const dealId = rawDealId.trim()
  if (!deals.value.some((deal) => deal.id === dealId)) {
    await loadDeals(true)
  }

  if (deals.value.some((deal) => deal.id === dealId)) {
    selectedDealId.value = dealId
  }
}

function handleCloseDealSheet() {
  selectedDealId.value = null

  if (route.query.dealId) {
    void router.replace({ name: 'deals' })
  }
}

onMounted(() => {
  window.addEventListener('resize', resizeHandler)
  resizeHandler()
  void loadDeals().then(() => openDealFromRouteQuery())
  void loadTasks()
})

onUnmounted(() => {
  window.removeEventListener('resize', resizeHandler)
})

watch(
  () => route.query.dealId,
  () => {
    void openDealFromRouteQuery()
  },
)

const groupedDeals = computed(() => {
  const map: Record<DealKanbanColumnId, Deal[]> = {
    production: [],
    pickup: [],
    delivery: [],
    closed: [],
    failed: [],
  }
  for (const deal of deals.value) {
    map[resolveDealKanbanColumnId(deal)].push(deal)
  }
  return map
})
</script>

<template>
  <section
    class="deals-kanban"
    :style="{ height: `${kanbanHeightPx}px`, maxHeight: `${kanbanHeightPx}px` }"
  >
    <div class="deals-kanban__viewport">
      <div class="deals-kanban__track">
        <DealsKanbanColumn
          v-for="column in DEAL_KANBAN_COLUMNS"
          :key="column.id"
          :column="column"
          :deals="groupedDeals[column.id]"
          @open-deal="selectedDealId = $event.id"
        />
      </div>
    </div>

    <DealDetailsSheet :deal-id="selectedDealId" @close="handleCloseDealSheet" />
  </section>
</template>

<style scoped>
.deals-kanban {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #ffffff;
}

.deals-kanban__viewport {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  overflow-x: auto;
  overflow-y: hidden;
  padding-top: 16px;
  box-sizing: border-box;
}

.deals-kanban__track {
  display: flex;
  gap: 10px;
  flex: 1;
  min-height: 0;
  width: max-content;
  margin: 0 auto;
  padding: 0 24px;
  box-sizing: border-box;
  align-items: stretch;
}
</style>
