<script setup lang="ts">
import { ref } from 'vue'
import DealsSectionHeader from '@/components/deals/DealsSectionHeader.vue'
import DealsKanban from '@/components/deals/DealsKanban.vue'

const searchQuery = ref('')
</script>

<template>
  <div class="deals-page">
    <DealsSectionHeader v-model:search-query="searchQuery" />
    <DealsKanban :search-query="searchQuery" />
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
