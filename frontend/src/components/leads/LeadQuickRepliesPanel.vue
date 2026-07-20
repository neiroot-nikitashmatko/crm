<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { NIcon } from 'naive-ui'
import {
  AddOutline,
  CheckmarkOutline,
  ChevronBackOutline,
  CloseOutline,
  CreateOutline,
  TrashOutline,
} from '@vicons/ionicons5'
import {
  createQuickReply,
  createQuickReplySection,
  deleteQuickReply,
  deleteQuickReplySection,
  fetchQuickReplySections,
  updateQuickReply,
  updateQuickReplySection,
} from '@/api/quickReplies'
import type { QuickReply, QuickReplySection } from '@/types/quickReply'

const emit = defineEmits<{
  select: [text: string]
  close: []
}>()

const sections = ref<QuickReplySection[]>([])
const loading = ref(false)
const error = ref('')
const selectedSectionId = ref<string | null>(null)
const view = ref<'sections' | 'replies'>('sections')

const creatingSection = ref(false)
const newSectionTitle = ref('')

const creatingReply = ref(false)
const newReplyTitle = ref('')
const newReplyBody = ref('')

const editingSectionId = ref<string | null>(null)
const editSectionTitle = ref('')

const editingReplyId = ref<string | null>(null)
const editReplyTitle = ref('')
const editReplyBody = ref('')

const selectedSection = computed(
  () => sections.value.find((item) => item.id === selectedSectionId.value) ?? null,
)

const editingReply = computed(
  () =>
    selectedSection.value?.replies.find((item) => item.id === editingReplyId.value) ?? null,
)

const isReplyEditorOpen = computed(
  () => creatingReply.value || editingReplyId.value !== null,
)

async function loadSections() {
  loading.value = true
  error.value = ''
  try {
    sections.value = await fetchQuickReplySections()
    if (
      selectedSectionId.value &&
      !sections.value.some((item) => item.id === selectedSectionId.value)
    ) {
      selectedSectionId.value = null
      view.value = 'sections'
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось загрузить быстрые ответы'
  } finally {
    loading.value = false
  }
}

function openSection(section: QuickReplySection) {
  selectedSectionId.value = section.id
  view.value = 'replies'
  creatingReply.value = false
  editingReplyId.value = null
}

function goBack() {
  view.value = 'sections'
  selectedSectionId.value = null
  creatingReply.value = false
  editingReplyId.value = null
}

function pickReply(reply: QuickReply) {
  emit('select', reply.body)
}

async function submitSection() {
  const title = newSectionTitle.value.trim()
  if (!title) return
  error.value = ''
  try {
    const section = await createQuickReplySection(title)
    sections.value = [...sections.value, section]
    newSectionTitle.value = ''
    creatingSection.value = false
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось создать раздел'
  }
}

async function saveSection(section: QuickReplySection) {
  const title = editSectionTitle.value.trim()
  if (!title) return
  error.value = ''
  try {
    const updated = await updateQuickReplySection(section.id, title)
    sections.value = sections.value.map((item) =>
      item.id === section.id ? { ...item, title: updated.title } : item,
    )
    editingSectionId.value = null
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось сохранить раздел'
  }
}

async function removeSection(section: QuickReplySection) {
  if (!window.confirm(`Удалить раздел «${section.title}» и все ответы в нём?`)) return
  error.value = ''
  try {
    await deleteQuickReplySection(section.id)
    sections.value = sections.value.filter((item) => item.id !== section.id)
    if (selectedSectionId.value === section.id) {
      goBack()
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось удалить раздел'
  }
}

async function submitReply() {
  if (!selectedSectionId.value) return
  const title = newReplyTitle.value.trim()
  const body = newReplyBody.value.trim()
  if (!title || !body) return
  error.value = ''
  try {
    const reply = await createQuickReply(selectedSectionId.value, title, body)
    sections.value = sections.value.map((section) =>
      section.id === selectedSectionId.value
        ? { ...section, replies: [...section.replies, reply] }
        : section,
    )
    newReplyTitle.value = ''
    newReplyBody.value = ''
    creatingReply.value = false
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось создать ответ'
  }
}

async function saveReply(reply: QuickReply) {
  const title = editReplyTitle.value.trim()
  const body = editReplyBody.value.trim()
  if (!title || !body) return
  error.value = ''
  try {
    const updated = await updateQuickReply(reply.id, title, body)
    sections.value = sections.value.map((section) =>
      section.id === reply.sectionId
        ? {
            ...section,
            replies: section.replies.map((item) => (item.id === reply.id ? updated : item)),
          }
        : section,
    )
    editingReplyId.value = null
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось сохранить ответ'
  }
}

async function removeReply(reply: QuickReply) {
  if (!window.confirm(`Удалить ответ «${reply.title}»?`)) return
  error.value = ''
  try {
    await deleteQuickReply(reply.id)
    sections.value = sections.value.map((section) =>
      section.id === reply.sectionId
        ? { ...section, replies: section.replies.filter((item) => item.id !== reply.id) }
        : section,
    )
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось удалить ответ'
  }
}

function startEditSection(section: QuickReplySection) {
  editingSectionId.value = section.id
  editSectionTitle.value = section.title
}

function startEditReply(reply: QuickReply) {
  creatingReply.value = false
  editingReplyId.value = reply.id
  editReplyTitle.value = reply.title
  editReplyBody.value = reply.body
}

function cancelReplyEditor() {
  creatingReply.value = false
  editingReplyId.value = null
  newReplyTitle.value = ''
  newReplyBody.value = ''
  editReplyTitle.value = ''
  editReplyBody.value = ''
}

async function submitReplyEditor() {
  if (editingReply.value) {
    await saveReply(editingReply.value)
    return
  }
  await submitReply()
}

watch(view, () => {
  creatingSection.value = false
  creatingReply.value = false
  editingReplyId.value = null
})

onMounted(() => {
  void loadSections()
})
</script>

<template>
  <div class="quick-replies-panel" :class="{ 'quick-replies-panel--editor': isReplyEditorOpen }">
    <header class="quick-replies-panel__header">
      <button
        v-if="view === 'replies'"
        type="button"
        class="quick-replies-panel__toolbar-btn"
        :title="isReplyEditorOpen ? 'К списку ответов' : 'Назад к разделам'"
        :aria-label="isReplyEditorOpen ? 'К списку ответов' : 'Назад к разделам'"
        @click="isReplyEditorOpen ? cancelReplyEditor() : goBack()"
      >
        <NIcon :size="18" :component="ChevronBackOutline" />
      </button>
      <h4 class="quick-replies-panel__title">
        <template v-if="view === 'sections'">Быстрые ответы</template>
        <template v-else-if="creatingReply">Новый ответ</template>
        <template v-else-if="editingReplyId">Редактирование</template>
        <template v-else>{{ selectedSection?.title || 'Ответы' }}</template>
      </h4>
      <div class="quick-replies-panel__header-actions">
        <button
          v-if="view === 'sections' && !creatingSection && !loading"
          type="button"
          class="quick-replies-panel__toolbar-btn"
          title="Добавить раздел"
          aria-label="Добавить раздел"
          @click="creatingSection = true"
        >
          <NIcon :size="18" :component="AddOutline" />
        </button>
        <button
          v-else-if="view === 'replies' && !isReplyEditorOpen && !loading"
          type="button"
          class="quick-replies-panel__toolbar-btn"
          title="Добавить ответ"
          aria-label="Добавить ответ"
          @click="creatingReply = true"
        >
          <NIcon :size="18" :component="AddOutline" />
        </button>
        <button
          type="button"
          class="quick-replies-panel__toolbar-btn"
          title="Закрыть"
          aria-label="Закрыть"
          @click="emit('close')"
        >
          <span aria-hidden="true">×</span>
        </button>
      </div>
    </header>

    <p v-if="error" class="quick-replies-panel__banner quick-replies-panel__banner--error">
      {{ error }}
    </p>
    <p v-else-if="loading" class="quick-replies-panel__banner">Загрузка…</p>

    <div v-else-if="view === 'sections'" class="quick-replies-panel__body">
      <p v-if="sections.length === 0 && !creatingSection" class="quick-replies-panel__empty">
        Разделов пока нет
      </p>

      <ul class="quick-replies-panel__grid">
        <li
          v-for="section in sections"
          :key="section.id"
          class="quick-replies-panel__cell"
        >
          <template v-if="editingSectionId === section.id">
            <input
              v-model="editSectionTitle"
              class="quick-replies-panel__cell-input"
              placeholder="Название раздела"
              autofocus
              @keydown.enter.prevent="saveSection(section)"
              @keydown.esc.prevent="editingSectionId = null"
            />
            <div class="quick-replies-panel__cell-actions">
              <button
                type="button"
                class="quick-replies-panel__icon-btn quick-replies-panel__icon-btn--confirm"
                title="Сохранить"
                aria-label="Сохранить раздел"
                @click="saveSection(section)"
              >
                <NIcon :size="14" :component="CheckmarkOutline" />
              </button>
              <button
                type="button"
                class="quick-replies-panel__icon-btn"
                title="Отмена"
                aria-label="Отменить редактирование"
                @click="editingSectionId = null"
              >
                <NIcon :size="14" :component="CloseOutline" />
              </button>
            </div>
          </template>
          <template v-else>
            <button type="button" class="quick-replies-panel__cell-main" @click="openSection(section)">
              <span class="quick-replies-panel__cell-title">{{ section.title }}</span>
            </button>
            <div class="quick-replies-panel__cell-actions">
              <button
                type="button"
                class="quick-replies-panel__icon-btn"
                title="Изменить"
                aria-label="Изменить раздел"
                @click="startEditSection(section)"
              >
                <NIcon :size="14" :component="CreateOutline" />
              </button>
              <button
                type="button"
                class="quick-replies-panel__icon-btn quick-replies-panel__icon-btn--danger"
                title="Удалить"
                aria-label="Удалить раздел"
                @click="removeSection(section)"
              >
                <NIcon :size="14" :component="TrashOutline" />
              </button>
            </div>
          </template>
        </li>
      </ul>

      <div v-if="creatingSection" class="quick-replies-panel__form">
        <input
          v-model="newSectionTitle"
          class="quick-replies-panel__input"
          placeholder="Название раздела"
          autofocus
          @keydown.enter.prevent="submitSection"
          @keydown.esc.prevent="creatingSection = false"
        />
        <div class="quick-replies-panel__inline-actions">
          <button type="button" class="quick-replies-panel__primary-btn" @click="submitSection">
            Создать
          </button>
          <button type="button" class="quick-replies-panel__link-btn" @click="creatingSection = false">
            Отмена
          </button>
        </div>
      </div>
    </div>

    <div v-else class="quick-replies-panel__body">
      <template v-if="isReplyEditorOpen">
        <div class="quick-replies-panel__editor">
          <template v-if="creatingReply">
            <input
              v-model="newReplyTitle"
              class="quick-replies-panel__input"
              placeholder="Название, например «Стоимость»"
              autofocus
            />
            <textarea
              v-model="newReplyBody"
              class="quick-replies-panel__textarea"
              placeholder="Текст, который вставится в чат"
            />
          </template>
          <template v-else>
            <input
              v-model="editReplyTitle"
              class="quick-replies-panel__input"
              placeholder="Название, например «Стоимость»"
              autofocus
            />
            <textarea
              v-model="editReplyBody"
              class="quick-replies-panel__textarea"
              placeholder="Текст, который вставится в чат"
            />
          </template>
          <div class="quick-replies-panel__editor-footer">
            <button type="button" class="quick-replies-panel__primary-btn" @click="submitReplyEditor">
              {{ creatingReply ? 'Создать' : 'Сохранить' }}
            </button>
            <button type="button" class="quick-replies-panel__link-btn" @click="cancelReplyEditor">
              Отмена
            </button>
          </div>
        </div>
      </template>

      <template v-else>
        <p
          v-if="!selectedSection || selectedSection.replies.length === 0"
          class="quick-replies-panel__empty"
        >
          Ответов пока нет
        </p>

        <ul class="quick-replies-panel__grid">
          <li
            v-for="reply in selectedSection?.replies ?? []"
            :key="reply.id"
            class="quick-replies-panel__cell"
          >
            <button
              type="button"
              class="quick-replies-panel__cell-main"
              :title="reply.body"
              @click="pickReply(reply)"
            >
              <span class="quick-replies-panel__cell-title">{{ reply.title }}</span>
            </button>
            <div class="quick-replies-panel__cell-actions">
              <button
                type="button"
                class="quick-replies-panel__icon-btn"
                title="Изменить"
                aria-label="Изменить ответ"
                @click="startEditReply(reply)"
              >
                <NIcon :size="14" :component="CreateOutline" />
              </button>
              <button
                type="button"
                class="quick-replies-panel__icon-btn quick-replies-panel__icon-btn--danger"
                title="Удалить"
                aria-label="Удалить ответ"
                @click="removeReply(reply)"
              >
                <NIcon :size="14" :component="TrashOutline" />
              </button>
            </div>
          </li>
        </ul>
      </template>
    </div>
  </div>
</template>

<style scoped>
.quick-replies-panel {
  display: flex;
  flex-direction: column;
  max-height: min(220px, 32dvh);
  border: 1px solid #d8e0e8;
  border-radius: 10px;
  background: #ffffff;
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.08);
  overflow: hidden;
}

.quick-replies-panel--editor {
  max-height: min(300px, 42dvh);
}

.quick-replies-panel__header {
  display: flex;
  align-items: center;
  gap: 6px;
  min-height: 40px;
  padding: 4px 6px 4px 8px;
  border-bottom: 1px solid #e2e8f0;
  flex-shrink: 0;
}

.quick-replies-panel__title {
  margin: 0;
  flex: 1 1 auto;
  min-width: 0;
  font-size: 13px;
  font-weight: 600;
  color: #1a202c;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.quick-replies-panel__header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.quick-replies-panel__toolbar-btn {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d1d9e2;
  border-radius: 8px;
  background: #ffffff;
  color: #475569;
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.quick-replies-panel__toolbar-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #1f2937;
}

.quick-replies-panel__body {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.quick-replies-panel__grid {
  list-style: none;
  margin: 0;
  padding: 6px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px;
  overflow-y: auto;
  min-height: 0;
}

.quick-replies-panel__cell {
  display: flex;
  align-items: center;
  gap: 2px;
  min-height: 32px;
  padding: 4px 4px 4px 8px;
  border: 1px solid #e8eef4;
  border-radius: 8px;
  background: #f8fafc;
}

.quick-replies-panel__cell:hover {
  border-color: #d5dee8;
  background: #f1f5f9;
}

.quick-replies-panel__cell-input {
  flex: 1 1 auto;
  min-width: 0;
  box-sizing: border-box;
  border: none;
  border-radius: 4px;
  padding: 2px 4px;
  font: inherit;
  font-size: 12px;
  font-weight: 500;
  color: #0f172a;
  background: transparent;
  outline: none;
}

.quick-replies-panel__cell-input:focus {
  background: #ffffff;
}

.quick-replies-panel__editor {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-height: 0;
  gap: 8px;
  padding: 8px;
}

.quick-replies-panel__editor-footer {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
  padding-top: 8px;
  margin-top: 2px;
  border-top: 1px solid #e2e8f0;
}

.quick-replies-panel__cell-main {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1 1 auto;
  min-width: 0;
  border: none;
  background: transparent;
  padding: 2px 0;
  text-align: left;
  cursor: pointer;
}

.quick-replies-panel__cell-title {
  flex: 1 1 auto;
  min-width: 0;
  font-size: 12px;
  font-weight: 500;
  color: #0f172a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.quick-replies-panel__cell-actions {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  margin-left: 2px;
  padding-left: 4px;
  border-left: 1px solid #d8e0e8;
}

.quick-replies-panel__icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: #64748b;
  cursor: pointer;
}

.quick-replies-panel__icon-btn:hover {
  background: #eef2f6;
  color: #0f172a;
}

.quick-replies-panel__icon-btn--danger {
  color: #b91c1c;
}

.quick-replies-panel__icon-btn--danger:hover {
  background: #fef2f2;
  color: #991b1b;
}

.quick-replies-panel__icon-btn--confirm {
  color: #1f883d;
}

.quick-replies-panel__icon-btn--confirm:hover {
  background: #f3faf5;
  color: #197a35;
}

.quick-replies-panel__form {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 8px;
  border-top: 1px solid #eef2f6;
  background: #fafbfc;
  flex-shrink: 0;
}

.quick-replies-panel__inline-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.quick-replies-panel__link-btn {
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 11px;
  padding: 0;
  cursor: pointer;
}

.quick-replies-panel__link-btn:hover {
  color: #0f172a;
}

.quick-replies-panel__primary-btn {
  border: none;
  border-radius: 6px;
  background: #1f883d;
  color: #ffffff;
  font-size: 11px;
  font-weight: 600;
  padding: 4px 10px;
  cursor: pointer;
}

.quick-replies-panel__primary-btn:hover {
  background: #197a35;
}

.quick-replies-panel__input,
.quick-replies-panel__textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #d8e0e8;
  border-radius: 6px;
  padding: 5px 8px;
  font: inherit;
  font-size: 12px;
  color: #0f172a;
  background: #ffffff;
}

.quick-replies-panel__textarea {
  flex: 1 1 auto;
  min-height: 88px;
  line-height: 1.4;
  resize: none;
  overflow-y: auto;
}

.quick-replies-panel__banner,
.quick-replies-panel__empty {
  margin: 0;
  padding: 8px 10px;
  font-size: 11px;
  color: #94a3b8;
}

.quick-replies-panel__banner--error {
  color: #b91c1c;
}

@media (max-width: 420px) {
  .quick-replies-panel__grid {
    grid-template-columns: 1fr;
  }
}

</style>
