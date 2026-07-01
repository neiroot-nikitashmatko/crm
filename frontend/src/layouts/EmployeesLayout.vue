<script setup lang="ts">
import EmployeesNav from '@/components/employees/EmployeesNav.vue'
import EmployeesSectionHeader from '@/components/employees/EmployeesSectionHeader.vue'
</script>

<template>
  <div class="employees-layout">
    <EmployeesSectionHeader />

    <div class="employees-layout__main">
      <EmployeesNav />

      <div class="employees-layout__content">
        <RouterView />
      </div>
    </div>
  </div>
</template>

<style scoped>
.employees-layout {
  display: flex;
  flex-direction: column;
  height: calc(100dvh - 64px);
  max-height: calc(100dvh - 64px);
  overflow: hidden;
  background: #ffffff;
}

.employees-layout__main {
  display: flex;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.employees-layout__content {
  flex: 1;
  min-width: 0;
  overflow: auto;
  padding: 24px;
  scrollbar-gutter: stable;
}
</style>
