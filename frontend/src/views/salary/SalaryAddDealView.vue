<script setup lang="ts">
import { onBeforeRouteLeave } from 'vue-router'
import { ref } from 'vue'
import SalaryAddDealForm from '@/components/salary/SalaryAddDealForm.vue'

const formRef = ref<InstanceType<typeof SalaryAddDealForm> | null>(null)

onBeforeRouteLeave(() => {
  formRef.value?.resetForm()
})
</script>

<template>
  <div class="salary-add-deal-view">
    <header class="salary-add-deal-view__header">
      <h2 class="salary-add-deal-view__title">Добавить сделку</h2>
      <p class="salary-add-deal-view__description">
        Заполните данные новой сделки. Поля со звёздочкой обязательны.
      </p>
    </header>

    <SalaryAddDealForm ref="formRef" />
  </div>
</template>

<style scoped>
.salary-add-deal-view {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 24px;
  width: 100%;
  max-width: 880px;
  min-height: calc(100dvh - 64px - 56px - 48px);
  margin: 0 auto;
  box-sizing: border-box;
}

.salary-add-deal-view__header {
  display: flex;
  flex-direction: column;
  gap: 6px;
  text-align: center;
}

.salary-add-deal-view__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.salary-add-deal-view__description {
  margin: 0;
  font-size: 14px;
  line-height: 1.45;
  color: #64748b;
}
</style>
