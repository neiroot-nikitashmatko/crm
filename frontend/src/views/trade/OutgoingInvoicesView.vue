<script setup lang="ts">
import { AddOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import SectionSubviewHeader from '@/components/common/SectionSubviewHeader.vue'

function handleCreateInvoice() {}
</script>

<template>
  <div class="trade-subview">
    <SectionSubviewHeader title="Расходные накладные">
      <template #actions>
        <button
          type="button"
          class="trade-subview__create-btn"
          title="Добавить расходную накладную"
          aria-label="Добавить расходную накладную"
          @click="handleCreateInvoice"
        >
          <NIcon :size="18" :component="AddOutline" />
        </button>
      </template>
    </SectionSubviewHeader>

    <div class="trade-subview__body" />
  </div>
</template>

<style scoped>
.trade-subview {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.trade-subview__body {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow: auto;
  padding: 24px;
  box-sizing: border-box;
  scrollbar-gutter: stable;
}

.trade-subview__create-btn {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #475569;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.trade-subview__create-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1f2937;
}
</style>
