<script setup lang="ts">
defineProps<{
  disabled?: boolean
  block?: boolean
}>()

defineEmits<{
  click: []
}>()
</script>

<template>
  <button
    type="button"
    class="app-modal__button"
    :class="{ 'app-modal__button--block': block }"
    :disabled="disabled"
    @click="$emit('click')"
  >
    <slot />
  </button>
</template>

<style scoped>
.app-modal__button {
  min-width: min(100%, 220px);
  padding: 10px 20px;
  border: 1px solid #1f883d;
  border-radius: 10px;
  background: #1f883d;
  color: #ffffff;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.35;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.app-modal__button:hover:not(:disabled) {
  background: #1a7534;
  border-color: #1a7534;
  box-shadow: 0 4px 12px rgba(31, 136, 61, 0.24);
}

.app-modal__button:disabled {
  border-color: #2aa04c;
  background: #2aa04c;
  color: #f8fff9;
  cursor: not-allowed;
  box-shadow: none;
}

.app-modal__button--block {
  min-width: min(100%, 320px);
}
</style>
