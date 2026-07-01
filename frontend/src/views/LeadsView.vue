<script setup lang="ts">
import LeadsSectionHeader from '@/components/leads/LeadsSectionHeader.vue'
import LeadsKanban from '@/components/leads/LeadsKanban.vue'
</script>

<template>
  <div class="leads-page">
    <LeadsSectionHeader />
    <LeadsKanban />
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
