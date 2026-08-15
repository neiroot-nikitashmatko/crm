<script setup lang="ts">
import { computed } from 'vue'
import AnalyticsKpiCard from '@/components/analytics/AnalyticsKpiCard.vue'
import AnalyticsLeadsTrafficCard from '@/components/analytics/AnalyticsLeadsTrafficCard.vue'
import { useFailedDealShare } from '@/composables/useFailedDealShare'
import { useFailedLeadShare } from '@/composables/useFailedLeadShare'
import { useLeadToDealConversion } from '@/composables/useLeadToDealConversion'
import {
  useDealTrafficAnalytics,
  useLeadTrafficAnalytics,
} from '@/composables/useLeadTrafficAnalytics'
import { useEmployeeShareAnalytics } from '@/composables/useEmployeeShareAnalytics'
import { useProductionShareAnalytics } from '@/composables/useProductionShareAnalytics'

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
  metrics: productionMetrics,
  isLoading: productionLoading,
  errorMessage: productionError,
} = useProductionShareAnalytics()

const {
  metrics: employeeMetrics,
  isLoading: employeeLoading,
  errorMessage: employeeError,
} = useEmployeeShareAnalytics()

const errorMessage = computed(
  () =>
    leadsError.value ||
    dealsError.value ||
    conversionError.value ||
    failedError.value ||
    failedDealError.value ||
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
      />
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
      />
      <AnalyticsKpiCard
        title="Доля проваленных сделок"
        :percent="failedDealPercent"
        :hint="failedDealHint"
        :loading="failedDealLoading"
        color="#e11d48"
      />
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
      />
      <AnalyticsLeadsTrafficCard
        title="Соотношение участия сотрудника в закрытых сделках"
        :metrics="employeeMetrics"
        :loading="employeeLoading"
        noun-one="сделка"
        noun-few="сделки"
        noun-many="сделок"
        legend-aria-label="Сотрудники производства"
      />
    </div>
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

@media (max-width: 960px) {
  .analytics-view__charts,
  .analytics-view__kpis {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
