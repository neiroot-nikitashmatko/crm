<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ChatbubblesOutline, LogOutOutline, PeopleOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { useAuth } from '@/composables/useAuth'
import { useNotificationBadges } from '@/composables/useNotificationBadges'
import logoUrl from '@/assets/logo.png'

const emit = defineEmits<{
  toggleSidebar: []
}>()

const router = useRouter()
const { logout } = useAuth()
const {
  newLeadsBadge,
  unreadChatsBadge,
  showLeadsBadge,
  showChatsBadge,
  startNotificationBadges,
  stopNotificationBadges,
} = useNotificationBadges()

onMounted(() => {
  startNotificationBadges()
})

onUnmounted(() => {
  stopNotificationBadges()
})

function handleLogout() {
  stopNotificationBadges()
  logout()
  router.replace({ name: 'login' })
}

function goToChats() {
  void router.push({ name: 'chats' })
}

function goToLeads() {
  void router.push({ name: 'leads' })
}
</script>

<template>
  <header class="app-header">
    <div class="app-header__brand">
      <button
        type="button"
        class="app-header__menu-btn"
        aria-label="Открыть главное меню"
        @click="emit('toggleSidebar')"
      >
        <svg
          class="app-header__menu-btn-icon"
          viewBox="0 0 24 24"
          fill="currentColor"
          aria-hidden="true"
        >
          <rect x="4" y="4" width="16" height="2" rx="1" />
          <rect x="4" y="11" width="16" height="2" rx="1" />
          <rect x="4" y="18" width="16" height="2" rx="1" />
        </svg>
      </button>

      <img class="app-header__logo" :src="logoUrl" alt="NEIROOT" />
    </div>

    <div class="app-header__actions">
      <button
        v-if="showChatsBadge"
        type="button"
        class="app-header__menu-btn app-header__indicator"
        aria-label="Чаты"
        title="Чаты"
        @click="goToChats"
      >
        <NIcon class="app-header__menu-btn-icon" :component="ChatbubblesOutline" :size="20" />
        <span v-if="unreadChatsBadge" class="app-header__badge">{{ unreadChatsBadge }}</span>
      </button>

      <button
        v-if="showLeadsBadge"
        type="button"
        class="app-header__menu-btn app-header__indicator"
        aria-label="Лиды"
        title="Лиды"
        @click="goToLeads"
      >
        <NIcon class="app-header__menu-btn-icon" :component="PeopleOutline" :size="20" />
        <span v-if="newLeadsBadge" class="app-header__badge">{{ newLeadsBadge }}</span>
      </button>

      <span
        v-if="showChatsBadge || showLeadsBadge"
        class="app-header__actions-divider"
        aria-hidden="true"
      />

      <button
        type="button"
        class="app-header__menu-btn"
        aria-label="Выйти из системы"
        @click="handleLogout"
      >
        <NIcon class="app-header__menu-btn-icon" :component="LogOutOutline" :size="20" />
      </button>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 16px;
  background: #f6f8fa;
  border-bottom: 1px solid #e2e8f0;
  flex-shrink: 0;
  z-index: 100;
}

.app-header__brand {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.app-header__logo {
  display: block;
  height: 30px;
  width: auto;
  max-width: min(220px, 45vw);
  object-fit: contain;
  user-select: none;
  pointer-events: none;
}

.app-header__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.app-header__actions-divider {
  flex-shrink: 0;
  align-self: center;
  width: 1px;
  height: 18px;
  margin: 0 2px;
  background: #d1d9e2;
}

.app-header__menu-btn {
  position: relative;
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
    border-color 0.15s ease,
    opacity 0.15s ease;
  -webkit-tap-highlight-color: transparent;
}

.app-header__menu-btn:hover {
  background: #eef1f4;
  border-color: #d0d0d0;
}

.app-header__menu-btn:active {
  background: #e8ebef;
  border-color: #c0c0c0;
}

.app-header__menu-btn-icon {
  width: 20px;
  height: 20px;
}

.app-header__badge {
  position: absolute;
  top: -6px;
  right: -6px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 999px;
  background: #dc2626;
  color: #ffffff;
  font-size: 10px;
  font-weight: 700;
  line-height: 16px;
  text-align: center;
  box-sizing: border-box;
  pointer-events: none;
}
</style>
