<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import {
  AttachOutline,
  FlashOutline,
  PaperPlaneOutline,
} from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import {
  fetchAvitoLeadChat,
  sendAvitoLeadMessage,
  subscribeAvitoMessages,
} from '@/api/avitoChat'
import type { LeadChatMessage, LeadChatParticipant } from '@/types/leadChat'
import LeadQuickRepliesPanel from './LeadQuickRepliesPanel.vue'

const props = defineProps<{
  leadId: string
}>()

const messages = ref<LeadChatMessage[]>([])
const participant = ref<LeadChatParticipant>({
  nickname: 'Пользователь Авито',
  avatarUrl: null,
})
const linked = ref(false)
const loading = ref(false)
const sending = ref(false)
const loadError = ref('')
const draft = ref('')
const messagesViewportRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLTextAreaElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const pendingFiles = ref<File[]>([])
const showQuickReplies = ref(false)

let sseAbort: AbortController | null = null

const avatarLetter = computed(() => {
  const name = participant.value.nickname.trim()
  return name ? name.charAt(0).toUpperCase() : '?'
})

const canSend = computed(
  () =>
    linked.value &&
    !sending.value &&
    (draft.value.trim().length > 0 || pendingFiles.value.length > 0),
)

const emptyText = computed(() => {
  if (loading.value) return 'Загрузка переписки…'
  if (loadError.value) return loadError.value
  if (!linked.value) {
    return 'Чат Авито ещё не привязан к этому лиду. Он появится после входящего сообщения.'
  }
  return 'Сообщений пока нет. Когда клиент напишет в Авито, переписка появится здесь.'
})

async function scrollToBottom() {
  await nextTick()
  const el = messagesViewportRef.value
  if (!el) return
  el.scrollTop = el.scrollHeight
}

watch(
  messages,
  () => {
    void scrollToBottom()
  },
  { deep: true },
)

function formatTime(timestamp: number): string {
  return new Date(timestamp).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

watch(draft, async () => {
  await nextTick()
  const el = inputRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = `${Math.min(el.scrollHeight, 168)}px`
})

function upsertMessage(message: LeadChatMessage) {
  const index = messages.value.findIndex((item) => item.id === message.id)
  if (index === -1) {
    messages.value = [...messages.value, message].sort((a, b) => a.createdAt - b.createdAt)
    return
  }
  const next = [...messages.value]
  next[index] = message
  messages.value = next
}

async function loadChat() {
  const leadId = props.leadId.trim()
  if (!leadId) return

  loading.value = true
  loadError.value = ''
  try {
    const bundle = await fetchAvitoLeadChat(leadId)
    linked.value = bundle.linked
    participant.value = bundle.participant
    messages.value = bundle.messages
  } catch (error) {
    linked.value = false
    messages.value = []
    loadError.value = error instanceof Error ? error.message : 'Не удалось загрузить чат'
  } finally {
    loading.value = false
    void scrollToBottom()
  }
}

function startMessageStream() {
  if (sseAbort) {
    sseAbort.abort()
    sseAbort = null
  }
  const controller = new AbortController()
  sseAbort = controller
  subscribeAvitoMessages(
    ({ leadId, message }) => {
      if (leadId !== props.leadId) return
      linked.value = true
      upsertMessage(message)
    },
    { signal: controller.signal },
  )
}

async function handleSend() {
  const text = draft.value.trim()
  const files = [...pendingFiles.value]
  if (!canSend.value) return

  const tempIds: string[] = []
  const optimistic: LeadChatMessage[] = []
  if (text) {
    const tempId = `local-text-${Date.now()}`
    tempIds.push(tempId)
    optimistic.push({
      id: tempId,
      direction: 'outgoing',
      text,
      kind: 'text',
      createdAt: Date.now(),
      status: 'sending',
    })
  }
  files.forEach((file, index) => {
    const tempId = `local-file-${Date.now()}-${index}`
    tempIds.push(tempId)
    optimistic.push({
      id: tempId,
      direction: 'outgoing',
      text: file.name,
      kind: 'image',
      imageUrl: URL.createObjectURL(file),
      createdAt: Date.now() + index + 1,
      status: 'sending',
    })
  })
  messages.value = [...messages.value, ...optimistic]
  draft.value = ''
  pendingFiles.value = []
  sending.value = true
  loadError.value = ''

  try {
    const saved = await sendAvitoLeadMessage(props.leadId, text, files)
    messages.value = messages.value.filter((item) => !tempIds.includes(item.id))
    for (const message of saved) {
      upsertMessage(message)
    }
  } catch (error) {
    messages.value = messages.value.map((item) =>
      tempIds.includes(item.id) ? { ...item, status: 'failed' as const } : item,
    )
    loadError.value = error instanceof Error ? error.message : 'Не удалось отправить сообщение'
  } finally {
    sending.value = false
  }
}

function handleQuickReplies() {
  showQuickReplies.value = !showQuickReplies.value
}

function applyQuickReply(text: string) {
  const next = text.trim()
  if (!next) return
  draft.value = draft.value.trim() ? `${draft.value.trim()}\n\n${next}` : next
  showQuickReplies.value = false
  void nextTick(() => {
    inputRef.value?.focus()
  })
}

function triggerAttach() {
  fileInputRef.value?.click()
}

function handleFilesSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files ?? [])
  if (files.length === 0) return
  const images = files.filter((file) => file.type.startsWith('image/') || /\.(jpe?g|png|gif|webp|bmp|heic|heif)$/i.test(file.name))
  if (images.length === 0) {
    loadError.value = 'Авито принимает только изображения (JPEG, PNG, GIF, WEBP…)'
    input.value = ''
    return
  }
  if (images.length < files.length) {
    loadError.value = 'Часть файлов пропущена: Авито принимает только изображения'
  } else {
    loadError.value = ''
  }
  pendingFiles.value = [...pendingFiles.value, ...images]
  input.value = ''
}

function removePendingFile(index: number) {
  pendingFiles.value = pendingFiles.value.filter((_, i) => i !== index)
}

function handleComposerKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    void handleSend()
  }
}

watch(
  () => props.leadId,
  () => {
    showQuickReplies.value = false
    void loadChat()
    startMessageStream()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  if (sseAbort) {
    sseAbort.abort()
    sseAbort = null
  }
})
</script>

<template>
  <div class="lead-avito-chat">
    <header class="lead-avito-chat__header">
      <div class="lead-avito-chat__avatar" aria-hidden="true">
        <img
          v-if="participant.avatarUrl"
          class="lead-avito-chat__avatar-img"
          :src="participant.avatarUrl"
          alt=""
        />
        <span v-else class="lead-avito-chat__avatar-fallback">{{ avatarLetter }}</span>
      </div>
      <div class="lead-avito-chat__peer">
        <p class="lead-avito-chat__nickname">{{ participant.nickname }}</p>
        <p class="lead-avito-chat__subtitle">Чат Авито</p>
      </div>
    </header>

    <div ref="messagesViewportRef" class="lead-avito-chat__messages" role="log" aria-live="polite">
      <p v-if="messages.length === 0" class="lead-avito-chat__empty">
        {{ emptyText }}
      </p>

      <div
        v-for="message in messages"
        :key="message.id"
        class="lead-avito-chat__bubble-row"
        :class="{
          'lead-avito-chat__bubble-row--outgoing': message.direction === 'outgoing',
        }"
      >
        <div
          class="lead-avito-chat__bubble"
          :class="{
            'lead-avito-chat__bubble--outgoing': message.direction === 'outgoing',
            'lead-avito-chat__bubble--incoming': message.direction === 'incoming',
            'lead-avito-chat__bubble--failed': message.status === 'failed',
          }"
        >
          <img
            v-if="message.kind === 'image' && message.imageUrl"
            class="lead-avito-chat__bubble-image"
            :src="message.imageUrl"
            alt="Изображение"
          />
          <p v-else-if="message.kind === 'image'" class="lead-avito-chat__bubble-text">[Изображение]</p>
          <p v-else class="lead-avito-chat__bubble-text">{{ message.text }}</p>
          <time class="lead-avito-chat__bubble-time" :datetime="new Date(message.createdAt).toISOString()">
            {{ formatTime(message.createdAt) }}
            <span v-if="message.status === 'sending'"> · отправка…</span>
            <span v-else-if="message.status === 'failed'"> · ошибка</span>
          </time>
        </div>
      </div>
    </div>

    <div class="lead-avito-chat__composer" :class="{ 'lead-avito-chat__composer--disabled': !linked }">
      <LeadQuickRepliesPanel
        v-if="showQuickReplies"
        class="lead-avito-chat__quick-replies"
        @select="applyQuickReply"
        @close="showQuickReplies = false"
      />

      <ul v-if="pendingFiles.length > 0" class="lead-avito-chat__files">
        <li v-for="(file, index) in pendingFiles" :key="`${file.name}-${index}`" class="lead-avito-chat__file">
          <span class="lead-avito-chat__file-name">{{ file.name }}</span>
          <button
            type="button"
            class="lead-avito-chat__file-remove"
            title="Убрать файл"
            aria-label="Убрать файл"
            @click="removePendingFile(index)"
          >
            ×
          </button>
        </li>
      </ul>

      <div class="lead-avito-chat__bar">
        <input
          ref="fileInputRef"
          type="file"
          class="lead-avito-chat__file-input"
          accept="image/jpeg,image/png,image/gif,image/webp,image/bmp,image/heic,image/heif,.jpg,.jpeg,.png,.gif,.webp,.bmp,.heic,.heif"
          multiple
          @change="handleFilesSelected"
        />

        <textarea
          ref="inputRef"
          v-model="draft"
          class="lead-avito-chat__input"
          rows="1"
          placeholder="Напишите сообщение или прикрепите фото..."
          :disabled="!linked || sending"
          @keydown="handleComposerKeydown"
        />

        <div class="lead-avito-chat__actions">
          <button
            type="button"
            class="lead-avito-chat__icon-btn"
            title="Прикрепить изображение"
            aria-label="Прикрепить изображение"
            :disabled="!linked"
            @click="triggerAttach"
          >
            <NIcon :size="18" :component="AttachOutline" />
          </button>

          <button
            type="button"
            class="lead-avito-chat__icon-btn"
            :class="{ 'lead-avito-chat__icon-btn--active': showQuickReplies }"
            title="Быстрые ответы"
            aria-label="Быстрые ответы"
            :disabled="!linked"
            @click="handleQuickReplies"
          >
            <NIcon :size="18" :component="FlashOutline" />
          </button>

          <button
            type="button"
            class="lead-avito-chat__icon-btn"
            :class="{ 'lead-avito-chat__icon-btn--primary': canSend }"
            title="Отправить сообщение"
            aria-label="Отправить сообщение"
            :disabled="!canSend"
            @click="void handleSend()"
          >
            <NIcon :size="18" :component="PaperPlaneOutline" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.lead-avito-chat {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
  height: 100%;
  overflow: hidden;
}

.lead-avito-chat__header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid #e2e8f0;
  flex-shrink: 0;
}

.lead-avito-chat__avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  background: #e2e8f0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.lead-avito-chat__avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.lead-avito-chat__avatar-fallback {
  font-size: 13px;
  font-weight: 700;
  color: #475569;
}

.lead-avito-chat__peer {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.lead-avito-chat__nickname {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #0f172a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.25;
}

.lead-avito-chat__subtitle {
  margin: 0;
  font-size: 11px;
  color: #64748b;
  line-height: 1.25;
}

.lead-avito-chat__messages {
  flex: 1 1 auto;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-gutter: stable;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
}

.lead-avito-chat__empty {
  margin: auto;
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: #64748b;
  max-width: 280px;
}

.lead-avito-chat__bubble-row {
  display: flex;
  justify-content: flex-start;
}

.lead-avito-chat__bubble-row--outgoing {
  justify-content: flex-end;
}

.lead-avito-chat__bubble {
  max-width: min(80%, 420px);
  padding: 8px 10px;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.lead-avito-chat__bubble--incoming {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-bottom-left-radius: 4px;
}

.lead-avito-chat__bubble--outgoing {
  background: #e8f5ec;
  border: 1px solid #c6e6d0;
  border-bottom-right-radius: 4px;
}

.lead-avito-chat__bubble--failed {
  background: #fef2f2;
  border-color: #fecaca;
}

.lead-avito-chat__bubble-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.4;
  color: #1a202c;
  white-space: pre-wrap;
  word-break: break-word;
}

.lead-avito-chat__bubble-image {
  display: block;
  max-width: min(260px, 70vw);
  max-height: 220px;
  width: auto;
  height: auto;
  border-radius: 8px;
  object-fit: contain;
  background: #ffffff;
}

.lead-avito-chat__bubble-time {
  align-self: flex-end;
  font-size: 11px;
  color: #64748b;
}

.lead-avito-chat__composer {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex-shrink: 0;
  margin-bottom: 0;
}

.lead-avito-chat__quick-replies {
  margin-bottom: 2px;
}

.lead-avito-chat__icon-btn--active {
  background: #e8f5e9;
  color: #1f883d;
}

.lead-avito-chat__icon-btn--active:hover:not(:disabled) {
  background: #dcefde;
  color: #197a35;
}

.lead-avito-chat__composer--disabled {
  opacity: 0.55;
  pointer-events: none;
}

.lead-avito-chat__bar {
  display: flex;
  align-items: flex-end;
  gap: 6px;
  padding: 4px 4px 4px 10px;
  border: 1px solid #cbd5e1;
  border-radius: 12px;
  background: #ffffff;
  box-sizing: border-box;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.lead-avito-chat__bar:focus-within {
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.lead-avito-chat__actions {
  display: flex;
  align-items: flex-end;
  align-self: stretch;
  gap: 2px;
  flex-shrink: 0;
  padding: 0 0 0 8px;
  margin-left: 4px;
  border-left: 1px solid #e2e8f0;
}

.lead-avito-chat__input {
  flex: 1 1 auto;
  min-width: 0;
  box-sizing: border-box;
  resize: none;
  min-height: 32px;
  max-height: 168px;
  padding: 6px 8px;
  border: none;
  border-radius: 8px;
  background: transparent;
  font: inherit;
  font-size: 13px;
  line-height: 1.4;
  color: #1a202c;
  outline: none;
}

.lead-avito-chat__files {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.lead-avito-chat__file {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  padding: 4px 8px;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
  background: #f8fafc;
  font-size: 12px;
  color: #334155;
}

.lead-avito-chat__file-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 180px;
}

.lead-avito-chat__file-remove {
  border: none;
  background: transparent;
  color: #64748b;
  cursor: pointer;
  padding: 0;
  line-height: 1;
  font-size: 14px;
}

.lead-avito-chat__file-input {
  display: none;
}

.lead-avito-chat__icon-btn {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #475569;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    color 0.15s ease;
}

.lead-avito-chat__icon-btn:hover {
  background: #f1f5f9;
  color: #1f2937;
}

.lead-avito-chat__icon-btn--primary {
  background: #1f883d;
  color: #ffffff;
}

.lead-avito-chat__icon-btn--primary:hover:not(:disabled) {
  background: #197a35;
  color: #ffffff;
}

.lead-avito-chat__icon-btn:disabled {
  opacity: 1;
  cursor: default;
  color: #475569;
}

.lead-avito-chat__icon-btn:disabled:hover {
  background: transparent;
  color: #475569;
}
</style>
