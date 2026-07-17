<script setup lang="ts">
import SalaryNav from '@/components/salary/SalaryNav.vue'
import SalarySectionHeader from '@/components/salary/SalarySectionHeader.vue'
</script>

<template>
  <div class="salary-layout">
    <SalarySectionHeader />

    <div class="salary-layout__main">
      <SalaryNav />

      <div class="salary-layout__content">
        <RouterView />
      </div>
    </div>
  </div>
</template>

<style scoped>
.salary-layout {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 64px);
  max-height: calc(100dvh - 64px);
  overflow: hidden;
  background: #ffffff;
}

.salary-layout__main {
  display: flex;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.salary-layout__content {
  flex: 1;
  min-width: 0;
  overflow: auto;
  padding: 24px;
  scrollbar-gutter: stable;
}
</style>
