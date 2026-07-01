<script setup lang="ts">
import { computed, ref } from 'vue'
import { DownloadOutline, TrashOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { MAX_TASK_ATTACHMENT_SIZE_BYTES, TASK_ATTACHMENT_ACCEPT } from '@/constants/attachments'
import { downloadAttachmentFile } from '@/api/attachments'
import { formatFileSize, getFileExtension } from '@/utils/file'
import type { TaskAttachment } from '@/types/task'

const props = defineProps<{
  attachments: TaskAttachment[]
}>()

const emit = defineEmits<{
  addFiles: [files: File[]]
  remove: [attachmentId: string]
}>()

const fileInputRef = ref<HTMLInputElement | null>(null)
const uploadError = ref('')
const isDragging = ref(false)

const hasAttachments = computed(() => props.attachments.length > 0)

function openPicker() {
  uploadError.value = ''
  fileInputRef.value?.click()
}

function validate(file: File): string | null {
  if (file.size > MAX_TASK_ATTACHMENT_SIZE_BYTES) {
    return `Файл ${file.name} больше 10 МБ`
  }
  return null
}

function processFiles(fileList: FileList | null) {
  if (!fileList || fileList.length === 0) return
  const files = Array.from(fileList)
  const errors = files.map(validate).filter((value): value is string => Boolean(value))
  if (errors.length > 0) {
    uploadError.value = errors[0]
    return
  }
  uploadError.value = ''
  emit('addFiles', files)
}

function handleInputChange(event: Event) {
  const target = event.target as HTMLInputElement | null
  processFiles(target?.files ?? null)
  if (target) target.value = ''
}

function handleDrop(event: DragEvent) {
  event.preventDefault()
  isDragging.value = false
  processFiles(event.dataTransfer?.files ?? null)
}

function handleDragOver(event: DragEvent) {
  event.preventDefault()
  isDragging.value = true
}

function handleDragLeave() {
  isDragging.value = false
}

function downloadAttachment(attachment: TaskAttachment) {
  void downloadAttachmentFile(attachment.id, attachment.name).catch((error) => {
    console.error('Не удалось скачать файл', error)
  })
}
</script>

<template>
  <section class="task-attachments">
    <input
      ref="fileInputRef"
      type="file"
      class="task-attachments__input"
      :accept="TASK_ATTACHMENT_ACCEPT"
      multiple
      @change="handleInputChange"
    />

    <button
      type="button"
      class="task-attachments__dropzone"
      :class="{ 'task-attachments__dropzone--dragging': isDragging }"
      @click="openPicker"
      @drop="handleDrop"
      @dragover="handleDragOver"
      @dragleave="handleDragLeave"
    >
      <span>Перетащите файлы сюда или нажмите для выбора</span>
    </button>

    <p v-if="uploadError" class="task-attachments__error">{{ uploadError }}</p>

    <ul v-if="hasAttachments" class="task-attachments__list">
      <li
        v-for="attachment in attachments"
        :key="attachment.id"
        class="task-attachments__item"
      >
        <div class="task-attachments__meta">
          <p class="task-attachments__name">{{ attachment.name }}</p>
          <p class="task-attachments__info">
            {{ formatFileSize(attachment.size) }}
            <span v-if="getFileExtension(attachment.name)">
              · {{ getFileExtension(attachment.name).toUpperCase() }}
            </span>
          </p>
        </div>

        <div class="task-attachments__actions">
          <button
            type="button"
            class="task-attachments__action-btn"
            @click="downloadAttachment(attachment)"
          >
            <NIcon :size="14">
              <DownloadOutline />
            </NIcon>
            Скачать
          </button>
          <button
            type="button"
            class="task-attachments__action-btn task-attachments__action-btn--danger"
            @click="emit('remove', attachment.id)"
          >
            <NIcon :size="14">
              <TrashOutline />
            </NIcon>
            Удалить
          </button>
        </div>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.task-attachments {
  min-width: 0;
  display: grid;
  gap: 10px;
}

.task-attachments__input {
  display: none;
}

.task-attachments__dropzone {
  min-height: 64px;
  border: 1px dashed #cbd5e1;
  border-radius: 8px;
  background: #f8fafc;
  color: #4a5568;
  font-size: 13px;
  cursor: pointer;
  padding: 10px;
}

.task-attachments__dropzone--dragging {
  border-color: #1f883d;
  background: #edf7ef;
}

.task-attachments__error {
  margin: 0;
  font-size: 12px;
  color: #dc2626;
}

.task-attachments__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 8px;
}

.task-attachments__item {
  min-width: 0;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 8px;
  display: grid;
  gap: 8px;
}

.task-attachments__meta {
  min-width: 0;
}

.task-attachments__name {
  margin: 0;
  font-size: 13px;
  color: #1a202c;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.task-attachments__info {
  margin: 2px 0 0;
  font-size: 12px;
  color: #718096;
}

.task-attachments__actions {
  display: flex;
  gap: 6px;
}

.task-attachments__action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-width: 74px;
  height: 28px;
  padding: 4px 8px;
  border: 1px solid #d1d9e2;
  border-radius: 6px;
  background: #ffffff;
  color: #475569;
  font-size: 11px;
  font-weight: 600;
  line-height: 1;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.task-attachments__action-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1f2937;
}

.task-attachments__action-btn--danger:hover {
  color: #dc2626;
}
</style>
