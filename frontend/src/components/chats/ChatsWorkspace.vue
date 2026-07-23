<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { SearchOutline, ArrowForwardOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { fetchAvitoChats, prefetchAvitoLeadChat, subscribeAvitoMessages, AvitoChatApiError } from '@/api/avitoChat'
import type { AvitoChatListItem } from '@/types/avitoChat'
import LeadAvitoChatPanel from '@/components/leads/LeadAvitoChatPanel.vue'

const route = useRoute()
const router = useRouter()

const chats = ref<AvitoChatListItem[]>([])
const loading = ref(false)
const error = ref('')
const selectedLeadId = ref<string | null>(null)
const searchQuery = ref('')

let messagesAbort: AbortController | null = null

const selectedChat = computed(
  () => chats.value.find((item) => item.leadId === selectedLeadId.value) ?? null,
)

const filteredChats = computed(() => {
  const normalizedQuery = searchQuery.value.trim().toLowerCase()
  if (!normalizedQuery) return chats.value
  return chats.value.filter((chat) =>
    chat.peerNickname.trim().toLowerCase().includes(normalizedQuery),
  )
})

function avatarLetter(chat: AvitoChatListItem): string {
  const name = chat.peerNickname.trim()
  return name ? name.charAt(0).toUpperCase() : '?'
}

function selectedAvatarLetter(): string {
  if (!selectedChat.value) return '?'
  return avatarLetter(selectedChat.value)
}

function clearUnreadLocally(leadId: string) {
  chats.value = chats.value.map((item) =>
    item.leadId === leadId ? { ...item, unreadCount: 0 } : item,
  )
}

function handleSearchKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && searchQuery.value) {
    event.preventDefault()
    searchQuery.value = ''
  }
}

function goToLead() {
  if (!selectedLeadId.value) return
  void router.push({ name: 'leads', query: { leadId: selectedLeadId.value } })
}

function formatLastMessageDate(timestamp: number): string {
  if (!Number.isFinite(timestamp) || timestamp <= 0) return ''
  const date = new Date(timestamp)
  const now = new Date()

  const isToday =
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate()

  if (isToday) {
    return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
  }

  return date.toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: '2-digit',
  })
}

async function loadChats() {
  loading.value = true
  error.value = ''
  try {
    chats.value = await fetchAvitoChats()
    const queryLeadId =
      typeof route.query.leadId === 'string' ? route.query.leadId.trim() : ''
    if (queryLeadId && chats.value.some((item) => item.leadId === queryLeadId)) {
      selectedLeadId.value = queryLeadId
      clearUnreadLocally(queryLeadId)
    } else if (
      selectedLeadId.value &&
      !chats.value.some((item) => item.leadId === selectedLeadId.value)
    ) {
      selectedLeadId.value = null
    }
  } catch (err) {
    error.value =
      err instanceof AvitoChatApiError || err instanceof Error
        ? err.message
        : 'Не удалось загрузить чаты'
  } finally {
    loading.value = false
  }
}

function selectChat(chat: AvitoChatListItem) {
  selectedLeadId.value = chat.leadId
  clearUnreadLocally(chat.leadId)
  void router.replace({ name: 'chats', query: { leadId: chat.leadId } })
}

function prefetchChat(chat: AvitoChatListItem) {
  prefetchAvitoLeadChat(chat.leadId)
}

function startMessageStream() {
  if (messagesAbort) {
    messagesAbort.abort()
    messagesAbort = null
  }
  const controller = new AbortController()
  messagesAbort = controller
  subscribeAvitoMessages(
    ({ leadId, message }) => {
      if (message.direction !== 'incoming') return
      if (leadId === selectedLeadId.value) {
        clearUnreadLocally(leadId)
        return
      }
      chats.value = chats.value.map((item) => {
        if (item.leadId !== leadId) return item
        return {
          ...item,
          unreadCount: item.unreadCount + 1,
          updatedAt: message.createdAt || Date.now(),
        }
      })
    },
    { signal: controller.signal },
  )
}

watch(
  () => route.query.leadId,
  (raw) => {
    if (typeof raw !== 'string' || raw.trim() === '') return
    const leadId = raw.trim()
    if (chats.value.some((item) => item.leadId === leadId)) {
      selectedLeadId.value = leadId
      clearUnreadLocally(leadId)
    }
  },
)

onMounted(() => {
  void loadChats()
  startMessageStream()
})

onBeforeUnmount(() => {
  if (messagesAbort) {
    messagesAbort.abort()
    messagesAbort = null
  }
})
</script>

<template>
  <div class="chats-workspace">
    <aside class="chats-workspace__sidebar">
      <header class="chats-workspace__pane-header" aria-label="Меню списка чатов">
        <div class="chats-workspace__search-wrap">
          <input
            class="chats-workspace__search"
            type="search"
            v-model="searchQuery"
            placeholder="Поиск"
            aria-label="Поиск чата по никнейму клиента"
            @keydown="handleSearchKeydown"
          />
          <span class="chats-workspace__search-icon" aria-hidden="true">
            <NIcon :size="16" :component="SearchOutline" />
          </span>
        </div>
      </header>

      <div class="chats-workspace__list-body">
        <p v-if="loading" class="chats-workspace__hint">Загрузка…</p>
        <p v-else-if="error" class="chats-workspace__hint chats-workspace__hint--error">
          {{ error }}
        </p>
        <p v-else-if="chats.length === 0" class="chats-workspace__hint">Чатов пока нет</p>
        <p v-else-if="filteredChats.length === 0" class="chats-workspace__hint">
          Ничего не найдено
        </p>

        <ul v-else class="chats-workspace__list" role="listbox" aria-label="Список чатов Авито">
          <li v-for="chat in filteredChats" :key="chat.id">
            <button
              type="button"
              class="chats-workspace__item"
              :class="{
                'chats-workspace__item--active': selectedLeadId === chat.leadId,
                'chats-workspace__item--unread': chat.unreadCount > 0,
              }"
              role="option"
              :aria-selected="selectedLeadId === chat.leadId"
              @click="selectChat(chat)"
              @pointerenter="prefetchChat(chat)"
            >
              <span class="chats-workspace__avatar" aria-hidden="true">
                <img
                  v-if="chat.peerAvatarUrl"
                  class="chats-workspace__avatar-img"
                  :src="chat.peerAvatarUrl"
                  alt=""
                />
                <span v-else class="chats-workspace__avatar-fallback">{{ avatarLetter(chat) }}</span>
              </span>
              <span class="chats-workspace__item-text">
                <span class="chats-workspace__item-name-row">
                  <span class="chats-workspace__item-name">{{ chat.peerNickname }}</span>
                  <span
                    v-if="chat.unreadCount > 0"
                    class="chats-workspace__item-dot"
                    aria-label="Есть непрочитанные сообщения"
                  />
                </span>
                <span v-if="chat.itemTitle" class="chats-workspace__item-meta">{{
                  chat.itemTitle
                }}</span>
              </span>
              <time
                v-if="chat.updatedAt"
                class="chats-workspace__item-date"
                :datetime="new Date(chat.updatedAt).toISOString()"
              >
                {{ formatLastMessageDate(chat.updatedAt) }}
              </time>
            </button>
          </li>
        </ul>
      </div>
    </aside>

    <section class="chats-workspace__conversation">
      <header class="chats-workspace__pane-header" aria-label="Меню переписки">
        <div v-if="selectedChat" class="chats-workspace__peer">
          <span class="chats-workspace__peer-avatar" aria-hidden="true">
            <img
              v-if="selectedChat.peerAvatarUrl"
              class="chats-workspace__peer-avatar-img"
              :src="selectedChat.peerAvatarUrl"
              alt=""
            />
            <span v-else class="chats-workspace__peer-avatar-fallback">{{
              selectedAvatarLetter()
            }}</span>
          </span>
          <div class="chats-workspace__peer-text">
            <p class="chats-workspace__peer-name">{{ selectedChat.peerNickname }}</p>
            <p class="chats-workspace__peer-subtitle">Чат Авито</p>
          </div>
        </div>
        <div class="chats-workspace__pane-actions">
          <button
            v-if="selectedLeadId"
            type="button"
            class="chats-workspace__icon-btn"
            title="Перейти в лид"
            aria-label="Перейти в лид"
            @click="goToLead"
          >
            <NIcon :size="20" :component="ArrowForwardOutline" />
          </button>
        </div>
      </header>

      <div class="chats-workspace__conversation-body">
        <LeadAvitoChatPanel
          v-if="selectedLeadId"
          :key="selectedLeadId"
          :lead-id="selectedLeadId"
          hide-header
        />
        <p v-else class="chats-workspace__hint">Выберите чат слева</p>
      </div>
    </section>
  </div>
</template>

<style scoped>
.chats-workspace {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #ffffff;
  overflow: hidden;
}

.chats-workspace__sidebar {
  display: flex;
  flex-direction: column;
  flex: 0 0 320px;
  width: 320px;
  max-width: 320px;
  border-right: 1px solid #e2e8f0;
  background: #f8fafc;
  min-height: 0;
  min-width: 0;
  overflow: hidden;
}

.chats-workspace__conversation {
  display: flex;
  flex-direction: column;
  flex: 1 1 0;
  min-width: 0;
  min-height: 0;
  background: #ffffff;
  overflow: hidden;
}

.chats-workspace__conversation > .chats-workspace__pane-header {
  padding: 0 14px;
}

.chats-workspace__pane-header {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
  min-height: 52px;
  padding: 0 10px;
  border-bottom: 1px solid #e2e8f0;
  background: #ffffff;
  box-sizing: border-box;
}

.chats-workspace__search-wrap {
  position: relative;
  width: 100%;
  min-width: 0;
}

.chats-workspace__search {
  width: 100%;
  height: 32px;
  padding: 0 32px 0 10px;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  box-sizing: border-box;
  font-size: 13px;
  color: #1a202c;
  outline: none;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.chats-workspace__search::placeholder {
  color: #94a3b8;
}

.chats-workspace__search:focus {
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(147, 197, 253, 0.25);
}

.chats-workspace__search::-webkit-search-cancel-button {
  -webkit-appearance: none;
}

.chats-workspace__search-icon {
  position: absolute;
  top: 50%;
  right: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  transform: translateY(-50%);
  pointer-events: none;
}

.chats-workspace__peer {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1 1 auto;
}

.chats-workspace__peer-avatar {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  border-radius: 50%;
  overflow: hidden;
  background: #e2e8f0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.chats-workspace__peer-avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.chats-workspace__peer-avatar-fallback {
  font-size: 13px;
  font-weight: 700;
  color: #475569;
}

.chats-workspace__peer-text {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.chats-workspace__peer-name {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #0f172a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.25;
}

.chats-workspace__peer-subtitle {
  margin: 0;
  font-size: 11px;
  color: #64748b;
  line-height: 1.25;
}

.chats-workspace__pane-actions {
  flex-shrink: 0;
  margin-left: auto;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.chats-workspace__icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  background: #f6f8fa;
  color: #4a5568;
  cursor: pointer;
  flex-shrink: 0;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease;
  -webkit-tap-highlight-color: transparent;
}

.chats-workspace__icon-btn:hover {
  background: #eef1f4;
  border-color: #d0d0d0;
}

.chats-workspace__icon-btn:active {
  background: #e8ebef;
  border-color: #c0c0c0;
}

.chats-workspace__list-body,
.chats-workspace__conversation-body {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.chats-workspace__list-body {
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.chats-workspace__conversation-body {
  padding: 10px;
}

.chats-workspace__conversation-body :deep(.lead-avito-chat) {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
}

.chats-workspace__list {
  list-style: none;
  margin: 0;
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.chats-workspace__item {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  width: 100%;
  padding: 10px 10px 10px 12px;
  border: none;
  border-radius: 10px;
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    box-shadow 0.15s ease;
}

.chats-workspace__item:hover {
  background: #eef2f6;
}

.chats-workspace__item--active {
  background: #eef7f0;
  box-shadow: inset 0 0 0 1px rgba(31, 136, 61, 0.18);
}

.chats-workspace__item--active::before {
  content: '';
  position: absolute;
  top: 8px;
  bottom: 8px;
  left: 0;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: #1f883d;
}

.chats-workspace__item--active .chats-workspace__item-name {
  color: #14532d;
}

.chats-workspace__item--active .chats-workspace__item-date {
  color: #6b9f7a;
}

.chats-workspace__item--unread .chats-workspace__item-name {
  font-weight: 700;
  color: #0f172a;
}

.chats-workspace__item--unread .chats-workspace__item-meta {
  font-weight: 600;
  color: #475569;
}

.chats-workspace__item--unread:not(.chats-workspace__item--active) {
  background: #e8f0fa;
}

.chats-workspace__item--unread:not(.chats-workspace__item--active):hover {
  background: #dde8f5;
}

.chats-workspace__avatar {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  border-radius: 50%;
  overflow: hidden;
  background: #e2e8f0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.chats-workspace__avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.chats-workspace__avatar-fallback {
  font-size: 14px;
  font-weight: 600;
  color: #475569;
}

.chats-workspace__item-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1 1 auto;
}

.chats-workspace__item-name-row {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.chats-workspace__item-name {
  font-size: 14px;
  font-weight: 600;
  color: #1a202c;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}

.chats-workspace__item-dot {
  width: 8px;
  height: 8px;
  flex-shrink: 0;
  border-radius: 50%;
  background: #1f883d;
}

.chats-workspace__item-meta {
  font-size: 12px;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.chats-workspace__item-date {
  flex-shrink: 0;
  margin-top: 2px;
  font-size: 11px;
  line-height: 1.2;
  color: #94a3b8;
  white-space: nowrap;
}

.chats-workspace__hint {
  margin: 0;
  padding: 16px;
  font-size: 13px;
  color: #64748b;
}

.chats-workspace__hint--error {
  color: #b91c1c;
}
</style>
