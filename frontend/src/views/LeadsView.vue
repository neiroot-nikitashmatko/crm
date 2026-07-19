<script setup lang="ts">
import { ref } from 'vue'
import LeadsSectionHeader from '@/components/leads/LeadsSectionHeader.vue'
import LeadsKanban from '@/components/leads/LeadsKanban.vue'

const searchQuery = ref('')
</script>

<template>
  <div class="leads-page">
    <LeadsSectionHeader v-model:search-query="searchQuery" />
    <LeadsKanban :search-query="searchQuery" />
  </div>
</template>

<style scoped>
.leads-page {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 64px);
  max-height: calc(100dvh - 64px);
  overflow: hidden;
  background: #ffffff;
}
</style>
