<script setup lang="ts">
withDefaults(
  defineProps<{
    src?: string
    initials?: string
    size?: number
  }>(),
  {
    src: '',
    initials: '',
    size: 32,
  },
)
</script>

<template>
  <span
    class="employee-avatar"
    :style="{ width: `${size}px`, height: `${size}px`, fontSize: `${Math.max(9, Math.round(size * 0.36))}px` }"
    aria-hidden="true"
  >
    <img v-if="src" class="employee-avatar__image" :src="src" alt="" />
    <span v-else-if="initials" class="employee-avatar__initials">{{ initials }}</span>
  </span>
</template>

<style scoped>
.employee-avatar {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  border-radius: 50%;
  background: #f6f8fa;
  color: #4a5568;
  font-weight: 700;
  letter-spacing: 0.02em;
  line-height: 1;
}

.employee-avatar__image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.employee-avatar__initials {
  user-select: none;
}
</style>
