<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NIcon } from 'naive-ui'
import { EyeOutline } from '@vicons/ionicons5'
import AnalyticsAmountCard from '@/components/analytics/AnalyticsAmountCard.vue'
import AnalyticsClosedDealsModal from '@/components/analytics/AnalyticsClosedDealsModal.vue'
import AnalyticsDealsTrafficModal from '@/components/analytics/AnalyticsDealsTrafficModal.vue'
import AnalyticsFailedLeadsModal from '@/components/analytics/AnalyticsFailedLeadsModal.vue'
import AnalyticsKpiCard from '@/components/analytics/AnalyticsKpiCard.vue'
import AnalyticsLeadsTrafficCard from '@/components/analytics/AnalyticsLeadsTrafficCard.vue'
import AnalyticsTradeProfitModal from '@/components/analytics/AnalyticsTradeProfitModal.vue'
import DealDetailsSheet from '@/components/deals/DealDetailsSheet.vue'
import { useClosedDealsList } from '@/composables/useClosedDealsList'
import { useDeals } from '@/composables/useDeals'
import { useDealsTrafficList } from '@/composables/useDealsTrafficList'
import { useEmployeeShareAnalytics } from '@/composables/useEmployeeShareAnalytics'
import { useFailedDealShare } from '@/composables/useFailedDealShare'
import { useFailedLeadShare } from '@/composables/useFailedLeadShare'
import { useFailedLeadsList } from '@/composables/useFailedLeadsList'
import { useLeadToDealConversion } from '@/composables/useLeadToDealConversion'
import {
  useDealTrafficAnalytics,
  useLeadTrafficAnalytics,
} from '@/composables/useLeadTrafficAnalytics'
import { useProductionShareAnalytics } from '@/composables/useProductionShareAnalytics'
import { useTradeProfit } from '@/composables/useTradeProfit'
import { useTradeProfitItems } from '@/composables/useTradeProfitItems'

const router = useRouter()

const {
  metrics: leadMetrics,
  isLoading: leadsLoading,
  errorMessage: leadsError,
} = useLeadTrafficAnalytics()

const {
  metrics: dealMetrics,
  isLoading: dealsLoading,
  errorMessage: dealsError,
} = useDealTrafficAnalytics()

const {
  conversion,
  isLoading: conversionLoading,
  errorMessage: conversionError,
} = useLeadToDealConversion()

const {
  share: failedShare,
  isLoading: failedLoading,
  errorMessage: failedError,
} = useFailedLeadShare()

const {
  share: failedDealShare,
  isLoading: failedDealLoading,
  errorMessage: failedDealError,
} = useFailedDealShare()

const {
  profitAmount: tradeProfitAmount,
  revenueAmount: tradeRevenueAmount,
  marginPercent: tradeMarginPercent,
  markupPercent: tradeMarkupPercent,
  emptyHint: tradeEmptyHint,
  markupHint: tradeMarkupHint,
  isLoading: tradeProfitLoading,
  errorMessage: tradeProfitError,
} = useTradeProfit()

const {
  items: tradeProfitItems,
  isLoading: tradeProfitItemsLoading,
  errorMessage: tradeProfitItemsError,
  loadItems: loadTradeProfitItems,
} = useTradeProfitItems()

const {
  metrics: productionMetrics,
  isLoading: productionLoading,
  errorMessage: productionError,
} = useProductionShareAnalytics()

const {
  metrics: employeeMetrics,
  isLoading: employeeLoading,
  errorMessage: employeeError,
} = useEmployeeShareAnalytics()

const {
  deals: closedDeals,
  isLoading: closedDealsLoading,
  errorMessage: closedDealsError,
  loadDeals: loadClosedDeals,
  loadFailedDeals,
} = useClosedDealsList()

const {
  leads: failedLeads,
  isLoading: failedLeadsLoading,
  errorMessage: failedLeadsError,
  loadFailedLeads,
} = useFailedLeadsList()

const {
  deals: trafficDeals,
  isLoading: trafficDealsLoading,
  errorMessage: trafficDealsError,
  loadDeals: loadTrafficDeals,
  showEmpty: showEmptyTrafficDeals,
} = useDealsTrafficList()

const { deals, loadDeals } = useDeals()

const isClosedDealsListOpen = ref(false)
const isFailedLeadsListOpen = ref(false)
const isDealsTrafficListOpen = ref(false)
const isTradeProfitListOpen = ref(false)
const selectedDealId = ref<string | null>(null)
const shouldReturnToClosedDealsList = ref(false)
const shouldReturnToDealsTrafficList = ref(false)
const closedDealsListDetail = ref<'nomenclature' | 'employee' | 'failed'>('nomenclature')

const errorMessage = computed(
  () =>
    leadsError.value ||
    dealsError.value ||
    conversionError.value ||
    failedError.value ||
    failedDealError.value ||
    tradeProfitError.value ||
    productionError.value ||
    employeeError.value,
)

const conversionPercent = computed(() => {
  if (conversion.value.leadsCount === 0) return null
  return conversion.value.percent
})

const conversionHint = computed(() => {
  const { leadsCount, convertedCount } = conversion.value
  if (leadsCount === 0) return 'Нет лидов за выбранный период'
  return `${convertedCount} из ${leadsCount} ${getLeadWord(leadsCount)}`
})

const failedPercent = computed(() => {
  if (failedShare.value.leadsCount === 0) return null
  return failedShare.value.percent
})

const failedHint = computed(() => {
  const { leadsCount, failedCount } = failedShare.value
  if (leadsCount === 0) return 'Нет лидов за выбранный период'
  return `${failedCount} из ${leadsCount} ${getLeadWord(leadsCount)}`
})

const failedDealPercent = computed(() => {
  if (failedDealShare.value.dealsCount === 0) return null
  return failedDealShare.value.percent
})

const failedDealHint = computed(() => {
  const { dealsCount, failedCount } = failedDealShare.value
  if (dealsCount === 0) return 'Нет сделок за выбранный период'
  return `${failedCount} из ${dealsCount} ${getDealWord(dealsCount)}`
})

const dealsTrafficTotal = computed(() =>
  dealMetrics.value.reduce((total, metric) => total + metric.count, 0),
)

async function openClosedDealsList(detail: 'nomenclature' | 'employee' = 'nomenclature') {
  closedDealsListDetail.value = detail
  isClosedDealsListOpen.value = true
  await loadClosedDeals({
    requireEmployee: detail === 'employee',
    requireProduction: detail === 'nomenclature',
  })
}

async function openFailedDealsList() {
  closedDealsListDetail.value = 'failed'
  isClosedDealsListOpen.value = true
  await loadFailedDeals()
}

async function openFailedLeadsList() {
  isFailedLeadsListOpen.value = true
  await loadFailedLeads()
}

async function openDealsTrafficList() {
  isDealsTrafficListOpen.value = true
  if (dealsTrafficTotal.value === 0) {
    showEmptyTrafficDeals()
    return
  }
  await loadTrafficDeals()
}

async function openTradeProfitList() {
  isTradeProfitListOpen.value = true
  await loadTradeProfitItems()
}

async function openClosedDeal(dealId: string) {
  shouldReturnToClosedDealsList.value = true
  isClosedDealsListOpen.value = false

  if (!deals.value.some((deal) => deal.id === dealId)) {
    await loadDeals(true)
  }

  if (deals.value.some((deal) => deal.id === dealId)) {
    selectedDealId.value = dealId
    return
  }

  shouldReturnToClosedDealsList.value = false
  isClosedDealsListOpen.value = true
}

async function openDealFromTrafficList(dealId: string) {
  shouldReturnToDealsTrafficList.value = true
  isDealsTrafficListOpen.value = false

  if (!deals.value.some((deal) => deal.id === dealId)) {
    await loadDeals(true)
  }

  if (deals.value.some((deal) => deal.id === dealId)) {
    selectedDealId.value = dealId
    return
  }

  shouldReturnToDealsTrafficList.value = false
  isDealsTrafficListOpen.value = true
}

async function openFailedLead(leadId: string) {
  if (!leadId) return
  isFailedLeadsListOpen.value = false
  await router.push({ name: 'leads', query: { leadId } })
}

function handleCloseDealSheet() {
  selectedDealId.value = null
  if (shouldReturnToClosedDealsList.value) {
    shouldReturnToClosedDealsList.value = false
    isClosedDealsListOpen.value = true
  }
  if (shouldReturnToDealsTrafficList.value) {
    shouldReturnToDealsTrafficList.value = false
    isDealsTrafficListOpen.value = true
  }
}

function getLeadWord(count: number): string {
  const lastTwoDigits = count % 100
  const lastDigit = count % 10

  if (lastTwoDigits >= 11 && lastTwoDigits <= 14) return 'лидов'
  if (lastDigit === 1) return 'лида'
  return 'лидов'
}

function getDealWord(count: number): string {
  const lastTwoDigits = count % 100
  const lastDigit = count % 10

  if (lastTwoDigits >= 11 && lastTwoDigits <= 14) return 'сделок'
  if (lastDigit === 1) return 'сделки'
  return 'сделок'
}
</script>

<template>
  <div class="analytics-view">
    <p v-if="errorMessage" class="analytics-view__error">{{ errorMessage }}</p>

    <div class="analytics-view__charts">
      <AnalyticsLeadsTrafficCard
        title="Количество лидов и источники трафика"
        :metrics="leadMetrics"
        :loading="leadsLoading"
        noun-one="лид"
        noun-few="лида"
        noun-many="лидов"
      />
      <AnalyticsLeadsTrafficCard
        title="Количество сделок и источники трафика"
        :metrics="dealMetrics"
        :loading="dealsLoading"
        noun-one="сделка"
        noun-few="сделки"
        noun-many="сделок"
      >
        <template #header-actions>
          <button
            type="button"
            class="analytics-view__icon-action"
            title="Список сделок"
            aria-label="Открыть список сделок за период"
            @click="openDealsTrafficList"
          >
            <NIcon :size="16">
              <EyeOutline />
            </NIcon>
          </button>
        </template>
      </AnalyticsLeadsTrafficCard>
    </div>

    <div class="analytics-view__kpis">
      <AnalyticsKpiCard
        title="Конверсия из лидов в сделку"
        :percent="conversionPercent"
        :hint="conversionHint"
        :loading="conversionLoading"
      />
      <AnalyticsKpiCard
        title="Доля проваленных лидов"
        :percent="failedPercent"
        :hint="failedHint"
        :loading="failedLoading"
        color="#e11d48"
      >
        <template #header-actions>
          <button
            type="button"
            class="analytics-view__icon-action"
            title="Список лидов"
            aria-label="Открыть список проваленных лидов"
            @click="openFailedLeadsList"
          >
            <NIcon :size="16">
              <EyeOutline />
            </NIcon>
          </button>
        </template>
      </AnalyticsKpiCard>
      <AnalyticsKpiCard
        title="Доля проваленных сделок"
        :percent="failedDealPercent"
        :hint="failedDealHint"
        :loading="failedDealLoading"
        color="#e11d48"
      >
        <template #header-actions>
          <button
            type="button"
            class="analytics-view__icon-action"
            title="Список сделок"
            aria-label="Открыть список проваленных сделок"
            @click="openFailedDealsList"
          >
            <NIcon :size="16">
              <EyeOutline />
            </NIcon>
          </button>
        </template>
      </AnalyticsKpiCard>
    </div>

    <div class="analytics-view__charts">
      <AnalyticsLeadsTrafficCard
        title="Соотношение производственных работ в закрытых сделках"
        :metrics="productionMetrics"
        :loading="productionLoading"
        noun-one="сделка"
        noun-few="сделки"
        noun-many="сделок"
        legend-aria-label="Категории производства"
      >
        <template #header-actions>
          <button
            type="button"
            class="analytics-view__icon-action"
            title="Список сделок"
            aria-label="Открыть список закрытых сделок"
            @click="openClosedDealsList('nomenclature')"
          >
            <NIcon :size="16">
              <EyeOutline />
            </NIcon>
          </button>
        </template>
      </AnalyticsLeadsTrafficCard>
      <AnalyticsLeadsTrafficCard
        title="Соотношение участия сотрудника в закрытых сделках"
        :metrics="employeeMetrics"
        :loading="employeeLoading"
        noun-one="сделка"
        noun-few="сделки"
        noun-many="сделок"
        legend-aria-label="Сотрудники производства"
      >
        <template #header-actions>
          <button
            type="button"
            class="analytics-view__icon-action"
            title="Список сделок"
            aria-label="Открыть список закрытых сделок"
            @click="openClosedDealsList('employee')"
          >
            <NIcon :size="16">
              <EyeOutline />
            </NIcon>
          </button>
        </template>
      </AnalyticsLeadsTrafficCard>
    </div>

    <div class="analytics-view__kpis analytics-view__kpis--trade">
      <AnalyticsAmountCard
        title="Прибыль"
        :amount="tradeProfitAmount"
        :hint="tradeEmptyHint"
        :loading="tradeProfitLoading"
      >
        <template #header-actions>
          <button
            type="button"
            class="analytics-view__icon-action"
            title="Из чего сложилась прибыль"
            aria-label="Открыть список товаров, из которых сложилась прибыль"
            @click="openTradeProfitList"
          >
            <NIcon :size="16">
              <EyeOutline />
            </NIcon>
          </button>
        </template>
      </AnalyticsAmountCard>
      <AnalyticsAmountCard
        title="Выручка"
        :amount="tradeRevenueAmount"
        :hint="tradeEmptyHint"
        :loading="tradeProfitLoading"
      />
      <AnalyticsAmountCard
        title="Маржа"
        format="percent"
        :amount="tradeMarginPercent"
        :hint="tradeEmptyHint"
        :loading="tradeProfitLoading"
      />
      <AnalyticsAmountCard
        title="Наценка"
        format="percent"
        :amount="tradeMarkupPercent"
        :hint="tradeMarkupHint"
        :loading="tradeProfitLoading"
      />
    </div>

    <AnalyticsClosedDealsModal
      v-model:show="isClosedDealsListOpen"
      :deals="closedDeals"
      :loading="closedDealsLoading"
      :error-message="closedDealsError"
      :detail="closedDealsListDetail"
      @select="openClosedDeal"
    />

    <AnalyticsFailedLeadsModal
      v-model:show="isFailedLeadsListOpen"
      :leads="failedLeads"
      :loading="failedLeadsLoading"
      :error-message="failedLeadsError"
      @select="openFailedLead"
    />

    <AnalyticsDealsTrafficModal
      v-model:show="isDealsTrafficListOpen"
      :deals="trafficDeals"
      :loading="trafficDealsLoading"
      :error-message="trafficDealsError"
      @select="openDealFromTrafficList"
    />

    <AnalyticsTradeProfitModal
      v-model:show="isTradeProfitListOpen"
      :items="tradeProfitItems"
      :loading="tradeProfitItemsLoading"
      :error-message="tradeProfitItemsError"
    />

    <DealDetailsSheet :deal-id="selectedDealId" @close="handleCloseDealSheet" />
  </div>
</template>

<style scoped>
.analytics-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 100%;
  background: #ffffff;
}

.analytics-view__error {
  margin: 0;
  color: #cf222e;
  font-size: 13px;
  line-height: 1.4;
}

.analytics-view__charts,
.analytics-view__kpis {
  display: grid;
  gap: 16px;
}

.analytics-view__charts {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.analytics-view__kpis {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.analytics-view__kpis--trade {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.analytics-view__icon-action {
  width: 30px;
  height: 30px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #64748b;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.analytics-view__icon-action:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

@media (max-width: 1200px) {
  .analytics-view__kpis,
  .analytics-view__kpis--trade {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 960px) {
  .analytics-view__charts,
  .analytics-view__kpis,
  .analytics-view__kpis--trade {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
