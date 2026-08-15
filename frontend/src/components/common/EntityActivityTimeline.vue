<script setup lang="ts">
import ActivityAuthorAvatar from '@/components/employees/ActivityAuthorAvatar.vue'
import { getAuthorShortName } from '@/constants/employees'

defineProps<{
  activities: Array<{
    id: string
    type: 'system' | 'comment'
    author: string
    authorId?: string
    text: string
    createdAt: number | Date
  }>
}>()

function formatCreatedAt(value: number | Date) {
  const timestamp = value instanceof Date ? value.getTime() : value
  if (!timestamp || Number.isNaN(timestamp)) return '—'
  return new Date(timestamp).toLocaleString('ru-RU')
}

function authorLabel(author: string) {
  return getAuthorShortName(author)
}
</script>

<template>
  <ul v-if="activities.length > 0" class="entity-timeline">
    <li
      v-for="entry in activities"
      :key="entry.id"
      class="entity-timeline__entry"
      :class="{
        'entity-timeline__entry--comment': entry.type === 'comment',
        'entity-timeline__entry--system': entry.type === 'system',
      }"
    >
      <span
        v-if="entry.type === 'comment'"
        class="entity-timeline__avatar"
        :title="authorLabel(entry.author)"
      >
        <ActivityAuthorAvatar :author="entry.author" :author-id="entry.authorId" :size="24" />
      </span>
      <div class="entity-timeline__body">
        <p class="entity-timeline__text">{{ entry.text }}</p>
        <p class="entity-timeline__meta">{{ formatCreatedAt(entry.createdAt) }}</p>
      </div>
    </li>
  </ul>
  <p v-else class="entity-timeline__empty">Таймлайн пока пуст</p>
</template>

<style scoped>
.entity-timeline {
  margin: 0;
  padding: 4px 0 0 20px;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0;
  border-left: 1px solid #e2e8f0;
}

.entity-timeline__entry {
  position: relative;
  padding: 0 0 14px 12px;
}

.entity-timeline__entry::before {
  content: '';
  position: absolute;
  left: -24px;
  top: 8px;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #cbd5e1;
  box-shadow: 0 0 0 3px #ffffff;
}

.entity-timeline__entry--comment::before {
  display: none;
}

.entity-timeline__avatar {
  position: absolute;
  left: -32px;
  top: 0;
  z-index: 1;
  border-radius: 50%;
  box-shadow: 0 0 0 2px #ffffff;
  line-height: 0;
}

.entity-timeline__body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.entity-timeline__text {
  margin: 0;
  font-size: 13px;
  line-height: 1.45;
  color: #1a202c;
  white-space: pre-wrap;
  word-break: break-word;
}

.entity-timeline__entry--system .entity-timeline__text {
  color: #4a5568;
}

.entity-timeline__meta {
  margin: 0;
  font-size: 12px;
  line-height: 1.35;
  color: #718096;
}

.entity-timeline__empty {
  margin: 0;
  font-size: 13px;
  color: #718096;
}
</style>
