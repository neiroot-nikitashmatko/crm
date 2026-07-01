<script setup lang="ts">
import DealsSectionHeader from '@/components/deals/DealsSectionHeader.vue'
import DealsKanban from '@/components/deals/DealsKanban.vue'
</script>

<template>
  <div class="deals-page">
    <DealsSectionHeader />
    <DealsKanban />
  </div>
</template>

<style scoped>
.deals-page {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 64px);
  max-height: calc(100dvh - 64px);
  overflow: hidden;
  background: #ffffff;
}
</style>
