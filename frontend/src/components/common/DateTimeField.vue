<script setup lang="ts">
defineProps<{
  displayValue: string
  hasValue: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  open: []
  clear: []
}>()

function handleClear(event: MouseEvent) {
  event.preventDefault()
  event.stopPropagation()
  emit('clear')
}
</script>

<template>
  <div
    class="date-time-field"
    :class="{
      'date-time-field--disabled': disabled,
      'date-time-field--has-clear': hasValue && !disabled,
    }"
  >
    <button
      type="button"
      class="date-time-field__trigger"
      :disabled="disabled"
      @click="emit('open')"
    >
      {{ displayValue }}
    </button>
    <button
      v-if="hasValue && !disabled"
      type="button"
      class="date-time-field__clear"
      aria-label="Очистить дату и время"
      @click="handleClear"
    >
      ×
    </button>
  </div>
</template>

<style scoped>
.date-time-field {
  position: relative;
  width: 100%;
}

.date-time-field__trigger {
  width: 100%;
  box-sizing: border-box;
  appearance: none;
  text-align: left;
  padding: 9px 12px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #0f172a;
  font-size: 14px;
  font-family: inherit;
  line-height: 1.3;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.date-time-field--has-clear .date-time-field__trigger {
  padding-right: 36px;
}

.date-time-field:not(.date-time-field--disabled):hover .date-time-field__trigger {
  border-color: #93c5fd;
}

.date-time-field--disabled .date-time-field__trigger {
  cursor: not-allowed;
  background: #f8fafc;
  color: #94a3b8;
}

.date-time-field__clear {
  position: absolute;
  top: 50%;
  right: 8px;
  transform: translateY(-50%);
  width: 22px;
  height: 22px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: #94a3b8;
  font-size: 16px;
  line-height: 1;
  cursor: pointer;
  transition:
    color 0.15s ease,
    background-color 0.15s ease;
}

.date-time-field__clear:hover {
  color: #64748b;
  background: #f1f5f9;
}
</style>
