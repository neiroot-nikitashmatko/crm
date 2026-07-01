<script setup lang="ts">
import { onBeforeRouteLeave } from 'vue-router'
import { ref } from 'vue'
import AddEmployeeForm from '@/components/employees/AddEmployeeForm.vue'

const formRef = ref<InstanceType<typeof AddEmployeeForm> | null>(null)

onBeforeRouteLeave(() => {
  formRef.value?.resetForm()
})
</script>

<template>
  <div class="add-employee-view">
    <header class="add-employee-view__header">
      <h2 class="add-employee-view__title">Добавить сотрудника</h2>
      <p class="add-employee-view__description">
        Заполните данные нового сотрудника. Поля со звёздочкой обязательны.
      </p>
    </header>

    <AddEmployeeForm ref="formRef" />
  </div>
</template>

<style scoped>
.add-employee-view {
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

.add-employee-view__header {
  display: flex;
  flex-direction: column;
  gap: 6px;
  text-align: center;
}

.add-employee-view__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a202c;
}

.add-employee-view__description {
  margin: 0;
  font-size: 14px;
  line-height: 1.45;
  color: #64748b;
}
</style>
