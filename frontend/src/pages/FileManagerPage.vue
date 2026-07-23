<template>
    <div class="tab-content">

        <div class="fm-header">
            <div class="fm-breadcrumb">
                <span class="crumb" @click="navigateTo('')">root</span>
                <template v-for="(seg, i) in breadcrumbs" :key="i">
                    <span class="crumb-sep">/</span>
                    <span class="crumb" @click="navigateTo(seg.path)">{{ seg.name }}</span>
                </template>
            </div>
            <div class="fm-actions">
                <button class="btn btn-outline" @click="loadFiles()">Refresh</button>
                <button class="btn btn-outline" @click="showNewFile = true">New File</button>
                <button class="btn btn-outline" @click="showNewFolder = true">New Folder</button>
            </div>
        </div>

        <div v-if="loading" class="fm-loading">Loading...</div>

        <div v-else-if="files.length === 0 && !editorFile" class="fm-empty">This directory is empty</div>

        <div v-else class="fm-table">
            <div class="fm-row header">
                <span class="fm-col name">Name</span>
                <span class="fm-col size">Size</span>
                <span class="fm-col date">Modified</span>
                <span class="fm-col actions">Actions</span>
            </div>
            <div v-for="f in files" :key="f.name" class="fm-row"
                @dblclick="f.isDir ? navigateTo(joinPath(currentPath, f.name)) : openFile(f)">
                <span class="fm-col name">
                    <span class="file-icon">{{ f.isDir ? '📁' : '📄' }}</span>
                    {{ f.name }}
                </span>
                <span class="fm-col size">{{ f.isDir ? '-' : formatSize(f.size) }}</span>
                <span class="fm-col date">{{ formatDate(f.modTime) }}</span>
                <span class="fm-col actions">
                    <button v-if="!f.isDir" class="tbl-btn" @click="openFile(f)">Edit</button>
                    <button class="tbl-btn danger" @click="handleDelete(f)">Delete</button>
                </span>
            </div>
        </div>

        <div v-if="editorFile" class="fm-editor">
            <div class="fm-editor-header">
                <span class="fm-editor-title">{{ editorFile }}</span>
                <div class="fm-editor-actions">
                    <button class="btn btn-outline" @click="closeEditor">Close</button>
                    <button class="btn btn-primary" :disabled="saving" @click="handleSave">
                        {{ saving ? 'Saving...' : 'Save' }}
                    </button>
                </div>
            </div>
            <textarea class="fm-editor-area" v-model="editorContent" spellcheck="false"></textarea>
        </div>

        <div v-if="showNewFile" class="modal-overlay" @click.self="showNewFile = false">
            <div class="modal-card">
                <h3 class="modal-title">New File</h3>
                <input class="fm-input" v-model="newFileName" placeholder="filename.txt"
                    @keydown.enter="handleCreateFile" ref="newFileInput" />
                <div class="modal-actions">
                    <button class="btn btn-outline" @click="showNewFile = false">Cancel</button>
                    <button class="btn btn-primary" @click="handleCreateFile" :disabled="!newFileName.trim()">Create</button>
                </div>
            </div>
        </div>

        <div v-if="showNewFolder" class="modal-overlay" @click.self="showNewFolder = false">
            <div class="modal-card">
                <h3 class="modal-title">New Folder</h3>
                <input class="fm-input" v-model="newFolderName" placeholder="folder-name"
                    @keydown.enter="handleCreateFolder" ref="newFolderInput" />
                <div class="modal-actions">
                    <button class="btn btn-outline" @click="showNewFolder = false">Cancel</button>
                    <button class="btn btn-primary" @click="handleCreateFolder" :disabled="!newFolderName.trim()">Create</button>
                </div>
            </div>
        </div>

    </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { api } from '../api'

const emit = defineEmits<{ (e: 'toast', v: { msg: string; type: string }): void }>()

interface FileItem {
    name: string
    isDir: boolean
    size: number
    modTime: string
}

const files = ref<FileItem[]>([])
const currentPath = ref('')
const loading = ref(false)

const editorFile = ref('')
const editorContent = ref('')
const saving = ref(false)

const showNewFile = ref(false)
const newFileName = ref('')
const showNewFolder = ref(false)
const newFolderName = ref('')

const newFileInput = ref<HTMLInputElement | null>(null)
const newFolderInput = ref<HTMLInputElement | null>(null)

const breadcrumbs = computed(() => {
    if (!currentPath.value) return []
    const parts = currentPath.value.split('/')
    return parts.map((name, i) => ({
        name,
        path: parts.slice(0, i + 1).join('/'),
    }))
})

function joinPath(base: string, name: string): string {
    return base ? base + '/' + name : name
}

async function loadFiles() {
    loading.value = true
    try {
        files.value = await api.listFiles(currentPath.value)
    } catch (e: any) {
        emit('toast', { msg: `Failed to load files: ${e.message}`, type: 'danger' })
    } finally {
        loading.value = false
    }
}

function navigateTo(path: string) {
    currentPath.value = path
    editorFile.value = ''
    editorContent.value = ''
    loadFiles()
}

async function openFile(f: FileItem) {
    const fullPath = joinPath(currentPath.value, f.name)
    try {
        const result = await api.readFile(fullPath)
        editorFile.value = result.path
        editorContent.value = result.content
    } catch (e: any) {
        emit('toast', { msg: `Failed to read file: ${e.message}`, type: 'danger' })
    }
}

function closeEditor() {
    editorFile.value = ''
    editorContent.value = ''
}

async function handleSave() {
    if (!editorFile.value) return
    saving.value = true
    try {
        await api.writeFile(editorFile.value, editorContent.value)
        emit('toast', { msg: 'File saved', type: 'success' })
    } catch (e: any) {
        emit('toast', { msg: `Save failed: ${e.message}`, type: 'danger' })
    } finally {
        saving.value = false
    }
}

async function handleDelete(f: FileItem) {
    const label = f.isDir ? `folder "${f.name}"` : `file "${f.name}"`
    if (!confirm(`Delete ${label}? This cannot be undone.`)) return
    const fullPath = joinPath(currentPath.value, f.name)
    try {
        await api.deleteFile(fullPath)
        emit('toast', { msg: `Deleted ${f.name}`, type: 'danger' })
        if (editorFile.value === fullPath) closeEditor()
        await loadFiles()
    } catch (e: any) {
        emit('toast', { msg: `Delete failed: ${e.message}`, type: 'danger' })
    }
}

async function handleCreateFile() {
    const name = newFileName.value.trim()
    if (!name) return
    const fullPath = joinPath(currentPath.value, name)
    try {
        await api.writeFile(fullPath, '')
        showNewFile.value = false
        newFileName.value = ''
        emit('toast', { msg: `Created ${name}`, type: 'success' })
        await loadFiles()
    } catch (e: any) {
        emit('toast', { msg: `Create failed: ${e.message}`, type: 'danger' })
    }
}

async function handleCreateFolder() {
    const name = newFolderName.value.trim()
    if (!name) return
    const fullPath = joinPath(currentPath.value, name)
    try {
        await api.makeDir(fullPath)
        showNewFolder.value = false
        newFolderName.value = ''
        emit('toast', { msg: `Created folder ${name}`, type: 'success' })
        await loadFiles()
    } catch (e: any) {
        emit('toast', { msg: `Create folder failed: ${e.message}`, type: 'danger' })
    }
}

function formatSize(bytes: number): string {
    if (bytes >= 1 << 30) return (bytes / (1 << 30)).toFixed(1) + ' GB'
    if (bytes >= 1 << 20) return (bytes / (1 << 20)).toFixed(1) + ' MB'
    if (bytes >= 1 << 10) return (bytes / (1 << 10)).toFixed(1) + ' KB'
    return bytes + ' B'
}

function formatDate(iso: string): string {
    if (!iso) return ''
    const d = new Date(iso)
    return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

onMounted(() => {
    loadFiles()
})
</script>

<style scoped>
.fm-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;
    flex-wrap: wrap;
    gap: 10px;
}

.fm-breadcrumb {
    display: flex;
    align-items: center;
    gap: 2px;
    font-size: 13px;
    color: var(--text2);
    font-family: 'VT323', monospace;
}

.crumb {
    cursor: pointer;
    padding: 2px 4px;
    border-radius: var(--radius);
    transition: background 0.15s;
}

.crumb:hover {
    background: var(--bg3);
    color: var(--green);
}

.crumb-sep {
    color: var(--muted);
    margin: 0 1px;
}

.fm-actions {
    display: flex;
    gap: 8px;
}

.fm-loading,
.fm-empty {
    text-align: center;
    padding: 40px 20px;
    color: var(--muted);
    font-size: 14px;
}

.fm-table {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
}

.fm-row {
    display: grid;
    grid-template-columns: 2fr 100px 160px 120px;
    align-items: center;
    padding: 10px 16px;
    border-bottom: 1px solid var(--border);
    font-size: 13px;
    cursor: pointer;
    transition: background 0.1s;
}

.fm-row:not(.header):hover {
    background: var(--bg3);
}

.fm-row:last-child {
    border-bottom: none;
}

.fm-row.header {
    font-weight: 600;
    color: var(--muted);
    font-size: 11px;
    letter-spacing: 0.5px;
    text-transform: uppercase;
    background: var(--bg3);
    cursor: default;
}

.fm-col.name {
    display: flex;
    align-items: center;
    gap: 8px;
    word-break: break-all;
}

.fm-col.actions {
    display: flex;
    gap: 6px;
    justify-content: flex-end;
}

.file-icon {
    font-size: 14px;
    flex-shrink: 0;
}

.fm-editor {
    margin-top: 20px;
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
}

.fm-editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 16px;
    background: var(--bg3);
    border-bottom: 1px solid var(--border);
}

.fm-editor-title {
    font-size: 13px;
    font-family: 'VT323', monospace;
    color: var(--green);
}

.fm-editor-actions {
    display: flex;
    gap: 8px;
}

.fm-editor-area {
    width: 100%;
    min-height: 300px;
    max-height: 500px;
    padding: 16px;
    background: #0a0a0f;
    color: var(--text2);
    border: none;
    font-family: 'Fira Code', 'Consolas', monospace;
    font-size: 13px;
    line-height: 1.6;
    resize: vertical;
    outline: none;
    tab-size: 4;
}

.modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 150;
}

.modal-card {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 24px;
    max-width: 420px;
    width: 90%;
}

.modal-title {
    font-family: 'VT323', monospace;
    font-size: 20px;
    letter-spacing: 1px;
    margin-bottom: 16px;
}

.modal-actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
    margin-top: 16px;
}

.fm-input {
    width: 100%;
    padding: 8px 12px;
    background: var(--bg3);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--text2);
    font-size: 13px;
    outline: none;
    font-family: 'Fira Code', 'Consolas', monospace;
    box-sizing: border-box;
}

.fm-input:focus {
    border-color: var(--green);
}
</style>
