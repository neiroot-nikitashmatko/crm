<script setup lang="ts">
import { provide } from 'vue'
import AnalyticsSectionHeader from '@/components/analytics/AnalyticsSectionHeader.vue'
import { ANALYTICS_PERIOD_KEY, useAnalyticsPeriod } from '@/composables/useAnalyticsPeriod'

const analyticsPeriod = useAnalyticsPeriod()

provide(ANALYTICS_PERIOD_KEY, analyticsPeriod)
</script>

<template>
  <div class="analytics-layout">
    <AnalyticsSectionHeader />

    <div class="analytics-layout__content">
      <RouterView />
    </div>
  </div>
</template>

<style scoped>
.analytics-layout {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 64px);
  max-height: calc(100dvh - 64px);
  overflow: hidden;
  background: #ffffff;
}

.analytics-layout__content {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 24px;
  scrollbar-gutter: stable;
}
</style>
