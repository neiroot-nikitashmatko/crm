<script setup lang="ts">
import TradeNav from '@/components/trade/TradeNav.vue'
import TradeSectionHeader from '@/components/trade/TradeSectionHeader.vue'
</script>

<template>
  <div class="trade-layout">
    <TradeSectionHeader />

    <div class="trade-layout__main">
      <TradeNav />

      <div class="trade-layout__content">
        <RouterView />
      </div>
    </div>
  </div>
</template>

<style scoped>
.trade-layout {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 64px);
  max-height: calc(100dvh - 64px);
  overflow: hidden;
  background: #ffffff;
}

.trade-layout__main {
  display: flex;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.trade-layout__content {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  padding: 0;
}

.trade-layout__content > * {
  flex: 1;
  min-height: 0;
}
</style>
