<script setup lang="ts">
import { ref } from 'vue'
import { DownloadOutline, TrashOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { downloadAttachmentFile, openAttachmentInNewTab } from '@/api/attachments'
import type { StoredAttachment } from '@/types/attachment'
import { canPreviewAttachment } from '@/utils/attachmentPreview'
import { formatFileSize, getFileExtension } from '@/utils/file'

const props = defineProps<{
  attachments: StoredAttachment[]
}>()

const emit = defineEmits<{
  remove: [attachmentId: string]
}>()

const openingAttachmentId = ref<string | null>(null)

async function openPreview(attachment: StoredAttachment) {
  if (!canPreviewAttachment(attachment) || openingAttachmentId.value) return

  openingAttachmentId.value = attachment.id
  try {
    await openAttachmentInNewTab(attachment.id, attachment.mimeType)
  } catch (error) {
    const message =
      error instanceof Error ? error.message : 'Не удалось открыть файл в новой вкладке'
    window.alert(message)
    console.error('Не удалось открыть файл', error)
  } finally {
    openingAttachmentId.value = null
  }
}

async function downloadAttachment(attachment: StoredAttachment) {
  try {
    await downloadAttachmentFile(attachment.id, attachment.name)
  } catch (error) {
    console.error('Не удалось скачать файл', error)
  }
}

function handleNameClick(attachment: StoredAttachment) {
  if (canPreviewAttachment(attachment)) {
    void openPreview(attachment)
  }
}
</script>

<template>
  <ul v-if="props.attachments.length > 0" class="entity-attachment-list">
    <li
      v-for="attachment in props.attachments"
      :key="attachment.id"
      class="entity-attachment-list__item"
    >
      <div class="entity-attachment-list__info">
        <p
          class="entity-attachment-list__name"
          :class="{
            'entity-attachment-list__name--previewable': canPreviewAttachment(attachment),
            'entity-attachment-list__name--loading': openingAttachmentId === attachment.id,
          }"
          :title="attachment.name"
          @click="handleNameClick(attachment)"
        >
          {{ attachment.name }}
        </p>
        <p class="entity-attachment-list__meta">
          {{ formatFileSize(attachment.size) }}
          <span v-if="getFileExtension(attachment.name)">
            · {{ getFileExtension(attachment.name).toUpperCase() }}
          </span>
        </p>
      </div>

      <div class="entity-attachment-list__actions">
        <button
          type="button"
          class="entity-attachment-list__icon-action"
          aria-label="Скачать файл"
          title="Скачать"
          @click="downloadAttachment(attachment)"
        >
          <NIcon :size="16">
            <DownloadOutline />
          </NIcon>
        </button>
        <button
          type="button"
          class="entity-attachment-list__icon-action entity-attachment-list__icon-action--danger"
          aria-label="Удалить файл"
          title="Удалить"
          @click="emit('remove', attachment.id)"
        >
          <NIcon :size="16">
            <TrashOutline />
          </NIcon>
        </button>
      </div>
    </li>
  </ul>
</template>

<style scoped>
.entity-attachment-list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.entity-attachment-list__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #ffffff;
  padding: 8px;
}

.entity-attachment-list__info {
  flex: 1 1 auto;
  min-width: 0;
}

.entity-attachment-list__name {
  margin: 0;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.35;
  color: #1a202c;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  word-break: break-word;
}

.entity-attachment-list__name--previewable {
  cursor: pointer;
  transition: color 0.15s ease;
}

.entity-attachment-list__name--previewable:hover {
  color: #1f883d;
}

.entity-attachment-list__name--loading {
  cursor: wait;
  color: #718096;
}

.entity-attachment-list__meta {
  margin: 4px 0 0;
  font-size: 12px;
  color: #718096;
}

.entity-attachment-list__actions {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  gap: 6px;
}

.entity-attachment-list__icon-action {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #64748b;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.entity-attachment-list__icon-action:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #334155;
}

.entity-attachment-list__icon-action--danger:hover:not(:disabled) {
  color: #dc2626;
}

.entity-attachment-list__icon-action:disabled {
  opacity: 0.55;
  cursor: wait;
}
</style>
