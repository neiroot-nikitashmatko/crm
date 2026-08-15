<script setup lang="ts">
import { computed, watch } from 'vue'
import EmployeeAvatar from '@/components/employees/EmployeeAvatar.vue'
import { getAuthorInitials } from '@/constants/employees'
import { useEmployees } from '@/composables/useEmployees'

const props = withDefaults(
  defineProps<{
    author: string
    authorId?: string
    size?: number
  }>(),
  {
    authorId: '',
    size: 20,
  },
)

const { avatarUrls, ensureAvatar } = useEmployees()

watch(
  () => props.authorId,
  (authorId) => {
    if (authorId) void ensureAvatar(authorId)
  },
  { immediate: true },
)

const src = computed(() => (props.authorId ? avatarUrls.value[props.authorId] ?? '' : ''))
const initials = computed(() => getAuthorInitials(props.author))
</script>

<template>
  <EmployeeAvatar :src="src" :initials="initials" :size="size" />
</template>
