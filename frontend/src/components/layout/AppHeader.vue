<script setup lang="ts">
import { useRouter } from 'vue-router'
import { LogOutOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { useAuth } from '@/composables/useAuth'
import logoUrl from '@/assets/logo.png'

const emit = defineEmits<{
  toggleSidebar: []
}>()

const router = useRouter()
const { logout } = useAuth()

function handleLogout() {
  logout()
  router.replace({ name: 'login' })
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

    <button
      type="button"
      class="app-header__menu-btn"
      aria-label="Выйти из системы"
      @click="handleLogout"
    >
      <NIcon class="app-header__menu-btn-icon" :component="LogOutOutline" :size="20" />
    </button>
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

.app-header__menu-btn {
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
</style>
