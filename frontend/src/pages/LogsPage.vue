<template>
    <div class="tab-content logs-page">

        <div class="logs-toolbar">
            <div class="logs-summary" v-if="logFiles.length > 0">
                <span class="summary-item">{{ logFiles.length }} log file{{ logFiles.length !== 1 ? 's' : '' }}</span>
                <span class="summary-sep">·</span>
                <span class="summary-item">{{ totalSize }}</span>
            </div>
            <div class="logs-actions">
                <button class="btn btn-outline" @click="refresh">Refresh</button>
            </div>
        </div>

        <div v-if="logFiles.length === 0" class="logs-empty">
            <span>No log files found — start the server to generate logs</span>
        </div>

        <div v-else class="logs-layout">
            <!-- ── File list ── -->
            <div class="logs-list card">
                <div class="card-header">
                    <span class="card-title">LOG FILES</span>
                </div>
                <div class="log-file-list">
                    <div v-for="file in logFiles" :key="file.name"
                        :class="['log-file-row', { active: activeFile?.name === file.name }]"
                        @click="openFile(file)">
                        <span class="log-file-icon">{{ file.isGzipped ? '📦' : '📄' }}</span>
                        <div class="log-file-meta">
                            <span class="log-file-name">{{ file.name }}</span>
                            <span class="log-file-sub">{{ file.size }} · {{ formatDate(file.modifiedDate) }}</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- ── Viewer ── -->
            <div class="logs-viewer card">
                <div class="card-header">
                    <span class="card-title">{{ activeFile ? activeFile.name : 'LOG VIEWER' }}</span>
                    <div class="viewer-controls" v-if="activeFile">
                        <input v-model="searchQuery" class="search-input" placeholder="Search within log..." />
                        <button class="icon-btn" @click="downloadActive" title="Download log">⬇</button>
                        <button class="icon-btn danger" @click="deleteActive" title="Delete log">🗑</button>
                    </div>
                </div>

                <div v-if="isLoading" class="logs-loading">
                    <span class="spinner-lg">◌</span><span>Loading...</span>
                </div>

                <div v-else-if="!activeFile" class="logs-placeholder">
                    <span class="placeholder-icon">📜</span>
                    <span>Select a log file to view its contents</span>
                </div>

                <div v-else class="logs-content-wrap">
                    <div v-if="content.truncated" class="truncated-banner">
                        Showing last 2MB — download the file for the full log
                    </div>
                    <pre class="logs-content" ref="contentEl" v-html="highlightedContent"></pre>
                </div>
            </div>
        </div>

    </div>
</template>

<script>
import { api } from '../api.js'

export default {
    name: 'LogsPage',
    emits: ['toast'],
    data() {
        return {
            logFiles: [],
            activeFile: null,
            content: { name: '', content: '', truncated: false },
            searchQuery: '',
            isLoading: false,
        }
    },
    computed: {
        totalSize() {
            if (!this.logFiles.length) return '0 B'
            const bytes = this.logFiles.reduce((sum, f) => sum + (f.sizeBytes || 0), 0)
            if (bytes < 1024) return bytes + ' B'
            if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
            return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
        },
        highlightedContent() {
            const lines = this.content.content.split('\n')
            return lines.map(line => this.highlightLine(line)).join('\n')
        },
    },
    async mounted() {
        await this.refresh()
    },
    methods: {
        formatDate(iso) {
            if (!iso) return ''
            const d = new Date(iso)
            return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        },
        async refresh() {
            try {
                this.logFiles = await api.getLogFiles()
                if (this.activeFile) {
                    const stillThere = this.logFiles.some(f => f.name === this.activeFile.name)
                    if (!stillThere) {
                        this.activeFile = null
                        this.content = { name: '', content: '', truncated: false }
                    }
                }
            } catch (e) {
                this.$emit('toast', { msg: `Failed to list logs: ${e.message}`, type: 'danger' })
            }
        },
        async openFile(file) {
            this.activeFile = file
            this.isLoading = true
            try {
                this.content = await api.readLogFile(file.name)
            } catch (e) {
                this.content = { name: file.name, content: '', truncated: false }
                this.$emit('toast', { msg: `Failed to read log: ${e.message}`, type: 'danger' })
            } finally {
                this.isLoading = false
            }
        },
        async downloadActive() {
            if (!this.activeFile) return
            try {
                const blob = await api.downloadLogFile(this.activeFile.name)
                const { saveBlob } = await import('../api')
                saveBlob(blob, this.activeFile.name)
            } catch (e) {
                this.$emit('toast', { msg: `Download failed: ${e.message}`, type: 'danger' })
            }
        },
        async deleteActive() {
            if (!this.activeFile) return
            if (!confirm(`Delete log file "${this.activeFile.name}"? This cannot be undone.`)) return
            try {
                await api.deleteLogFile(this.activeFile.name)
                this.$emit('toast', { msg: `Deleted log: ${this.activeFile.name}`, type: 'danger' })
                this.activeFile = null
                this.content = { name: '', content: '', truncated: false }
                await this.refresh()
            } catch (e) {
                this.$emit('toast', { msg: `Delete failed: ${e.message}`, type: 'danger' })
            }
        },
        highlightLine(line) {
            if (!this.searchQuery.trim()) return this.escapeHtml(line)
            const q = this.searchQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
            const regex = new RegExp(`(${q})`, 'gi')
            return this.escapeHtml(line).replace(regex, '<mark class="search-hl">$1</mark>')
        },
        escapeHtml(str) {
            return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
        },
    },
}
</script>

<style scoped>
.logs-page {
    height: calc(100vh - 130px);
    display: flex;
    flex-direction: column;
}

.logs-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    flex-shrink: 0;
}

.logs-summary {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: var(--muted);
}

.summary-sep {
    opacity: 0.4;
}

.logs-actions {
    display: flex;
    gap: 10px;
}

.logs-empty {
    text-align: center;
    padding: 40px 20px;
    color: var(--muted);
    font-size: 14px;
}

.logs-layout {
    display: grid;
    grid-template-columns: 280px 1fr;
    gap: 16px;
    flex: 1;
    min-height: 0;
}

.logs-list {
    display: flex;
    flex-direction: column;
    min-height: 0;
}

.log-file-list {
    flex: 1;
    overflow-y: auto;
    padding: 6px;
}

.log-file-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    border-radius: var(--radius);
    cursor: pointer;
    transition: background 0.15s;
}

.log-file-row:hover {
    background: var(--bg3);
}

.log-file-row.active {
    background: rgba(74, 222, 128, 0.08);
    border: 1px solid rgba(74, 222, 128, 0.3);
}

.log-file-icon {
    font-size: 18px;
    flex-shrink: 0;
}

.log-file-meta {
    min-width: 0;
}

.log-file-name {
    display: block;
    font-size: 13px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.log-file-sub {
    display: block;
    font-size: 10px;
    color: var(--muted);
    margin-top: 2px;
}

.logs-viewer {
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
}

.viewer-controls {
    display: flex;
    align-items: center;
    gap: 8px;
}

.search-input {
    width: 220px;
    padding: 6px 10px;
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--text);
    font-family: 'Share Tech Mono', monospace;
    font-size: 12px;
}

.search-input:focus {
    outline: none;
    border-color: var(--green);
}

.logs-loading,
.logs-placeholder {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    color: var(--muted);
    font-size: 13px;
}

.placeholder-icon {
    font-size: 32px;
}

.logs-content-wrap {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
}

.truncated-banner {
    padding: 6px 16px;
    background: rgba(251, 191, 36, 0.1);
    border-bottom: 1px solid rgba(251, 191, 36, 0.3);
    color: var(--yellow);
    font-size: 11px;
    flex-shrink: 0;
}

.logs-content {
    flex: 1;
    overflow-y: auto;
    margin: 0;
    padding: 12px 16px;
    font-family: 'Share Tech Mono', monospace;
    font-size: 12px;
    line-height: 1.7;
    color: var(--text2);
    background: #060810;
    white-space: pre-wrap;
    word-break: break-word;
}

.spinner-lg {
    font-size: 24px;
    animation: spin 1s linear infinite;
    display: inline-block;
}

.icon-btn.danger {
    color: var(--red);
}

:deep(.search-hl) {
    background: rgba(250, 204, 21, 0.3);
    color: #facc15;
    border-radius: 2px;
    padding: 0 1px;
}
</style>
