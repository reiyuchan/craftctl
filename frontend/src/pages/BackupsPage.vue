<template>
    <div class="tab-content">

        <div class="backups-header">
            <div class="backups-summary" v-if="store.backups.length > 0">
                <span class="summary-item">{{ store.backups.length }} backup{{ store.backups.length !== 1 ? 's' : '' }}</span>
                <span class="summary-sep">·</span>
                <span class="summary-item">{{ totalSize }}</span>
            </div>
            <div class="backups-actions">
                <button class="btn btn-outline" @click="store.fetchBackups()">Refresh</button>
                <button class="btn btn-primary" :disabled="creatingFull" @click="handleCreateFull">
                    {{ creatingFull ? 'Creating...' : 'CREATE FULL BACKUP' }}
                </button>
            </div>
        </div>

        <div v-if="store.backups.length === 0" class="backups-empty">
            <span>No backups found</span>
        </div>

        <div class="backups-table" v-else>
            <div class="backup-row header">
                <span class="backup-col name">Name</span>
                <span class="backup-col type">Type</span>
                <span class="backup-col size">Size</span>
                <span class="backup-col date">Date</span>
                <span class="backup-col actions">Actions</span>
            </div>
            <div v-for="backup in store.backups" :key="backup.name" class="backup-row">
                <span class="backup-col name">{{ backup.name }}</span>
                <span class="backup-col type">
                    <span class="type-badge" :class="backup.type">{{ backup.type }}</span>
                </span>
                <span class="backup-col size">{{ backup.size }}</span>
                <span class="backup-col date">{{ formatDate(backup.modifiedDate) }}</span>
                <span class="backup-col actions">
                    <button class="tbl-btn" :disabled="store.serverStatus !== 'stopped'"
                        @click="handleRestore(backup)">
                        Restore
                    </button>
                    <button class="tbl-btn" :disabled="downloadingBackup" @click="handleDownload(backup)">
                        {{ downloadingBackup ? '...' : 'Download' }}
                    </button>
                    <button class="tbl-btn danger" @click="handleDeleteBackup(backup)">Delete</button>
                </span>
            </div>
        </div>

        <div v-if="showRestoreModal" class="modal-overlay" @click.self="showRestoreModal = false">
            <div class="modal-card">
                <h3 class="modal-title">Restore Backup</h3>
                <p class="modal-text">
                    Are you sure you want to restore from <strong>{{ restoreTarget?.name }}</strong>?
                    This will overwrite current server files.
                </p>
                <div class="modal-actions">
                    <button class="btn btn-outline" @click="showRestoreModal = false">Cancel</button>
                    <button class="btn btn-danger" @click="confirmRestore">Restore</button>
                </div>
            </div>
        </div>

        <div class="config-section card settings-card">
            <div class="card-header">
                <span class="card-title">CONFIG EXPORT / IMPORT</span>
            </div>
            <div class="settings-body">
                <div class="config-row">
                    <div class="setting-info">
                        <span class="setting-name">Export</span>
                        <span class="setting-desc">Download server.properties, eula.txt, scheduler.json as zip</span>
                    </div>
                    <button class="btn btn-primary" @click="handleExportConfig" :disabled="exportingConfig">
                        {{ exportingConfig ? 'Exporting...' : 'EXPORT' }}
                    </button>
                </div>
                <div class="config-row">
                    <div class="setting-info">
                        <span class="setting-name">Import</span>
                        <span class="setting-desc">Upload a previously exported zip to restore config files</span>
                    </div>
                    <div class="config-import-row">
                        <label class="file-label">
                            <input type="file" accept=".zip" @change="onConfigFileChange" ref="configFileInput" style="display:none" />
                            <span class="btn btn-outline">{{ configFileName || 'Choose zip...' }}</span>
                        </label>
                        <button class="btn btn-primary" :disabled="!configFile || importingConfig" @click="handleImportConfig">
                            {{ importingConfig ? 'Importing...' : 'IMPORT' }}
                        </button>
                    </div>
                </div>
                <div v-if="importedFiles.length" class="config-result">
                    <span class="result-label">Extracted:</span>
                    <span v-for="f in importedFiles" :key="f" class="result-file">{{ f }}</span>
                </div>
            </div>
        </div>

    </div>
</template>

<script>
import { store } from '../store.js'
import { api } from '../api.js'

export default {
    name: 'BackupsPage',
    emits: ['toast'],
    data() {
        return {
            store,
            creatingFull: false,
            showRestoreModal: false,
            restoreTarget: null,
            exportingConfig: false,
            importingConfig: false,
            downloadingBackup: false,
            configFile: null,
            configFileName: '',
            importedFiles: [],
        }
    },
    computed: {
        totalSize() {
            if (!this.store.backups.length) return '0 B'
            const bytes = this.store.backups.reduce((sum, b) => sum + (b.sizeBytes || 0), 0)
            if (bytes < 1024) return bytes + ' B'
            if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
            return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
        },
    },
    async mounted() {
        await this.store.fetchBackups()
    },
    methods: {
        formatDate(iso) {
            if (!iso) return ''
            const d = new Date(iso)
            return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        },
        async handleCreateFull() {
            this.creatingFull = true
            try {
                await this.store.createFullBackup()
                this.$emit('toast', { msg: 'Full backup created', type: 'success' })
            } catch (e) {
                this.$emit('toast', { msg: `Full backup failed: ${e.message}`, type: 'danger' })
            } finally {
                this.creatingFull = false
            }
        },
        handleRestore(backup) {
            this.restoreTarget = backup
            this.showRestoreModal = true
        },
        async confirmRestore() {
            this.showRestoreModal = false
            try {
                await this.store.restoreBackup(this.restoreTarget.name)
                this.$emit('toast', { msg: `Restored: ${this.restoreTarget.name}`, type: 'success' })
            } catch (e) {
                this.$emit('toast', { msg: `Restore failed: ${e.message}`, type: 'danger' })
            }
        },
        async handleDeleteBackup(backup) {
            if (!confirm(`Delete backup "${backup.name}"? This cannot be undone.`)) return
            try {
                await this.store.deleteBackup(backup.name)
                this.$emit('toast', { msg: `Deleted backup: ${backup.name}`, type: 'danger' })
            } catch (e) {
                this.$emit('toast', { msg: `Delete failed: ${e.message}`, type: 'danger' })
            }
        },
        async handleDownload(backup) {
            this.downloadingBackup = true
            try {
                const blob = await api.downloadBackup(backup.name)
                const { saveBlob } = await import('../api')
                saveBlob(blob, backup.name)
                this.$emit('toast', { msg: `Downloaded backup: ${backup.name}`, type: 'success' })
            } catch (e) {
                this.$emit('toast', { msg: `Download failed: ${e.message}`, type: 'danger' })
            } finally {
                this.downloadingBackup = false
            }
        },
        async handleExportConfig() {
            this.exportingConfig = true
            try {
                const blob = await api.exportConfig()
                const url = URL.createObjectURL(blob)
                const a = document.createElement('a')
                a.href = url
                a.download = 'craftctl-config.zip'
                a.click()
                URL.revokeObjectURL(url)
                this.$emit('toast', { msg: 'Config exported!', type: 'success' })
            } catch (e) {
                this.$emit('toast', { msg: `Export failed: ${e.message}`, type: 'danger' })
            } finally {
                this.exportingConfig = false
            }
        },
        onConfigFileChange(e) {
            const input = e.target
            this.configFile = input.files?.[0] ?? null
            this.configFileName = this.configFile?.name || ''
        },
        async handleImportConfig() {
            if (!this.configFile) return
            this.importingConfig = true
            this.importedFiles = []
            try {
                const data = await api.importConfig(this.configFile)
                this.importedFiles = data.files ?? []
                this.$emit('toast', { msg: `Imported ${this.importedFiles.length} file(s)`, type: 'success' })
            } catch (e) {
                this.$emit('toast', { msg: `Import failed: ${e.message}`, type: 'danger' })
            } finally {
                this.importingConfig = false
            }
        },
    },
}
</script>

<style scoped>
.backups-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;
}

.backups-summary {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: var(--muted);
}

.summary-sep {
    opacity: 0.4;
}

.backups-actions {
    display: flex;
    gap: 10px;
}

.backups-empty {
    text-align: center;
    padding: 40px 20px;
    color: var(--muted);
    font-size: 14px;
}

.backups-table {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
}

.backup-row {
    display: grid;
    grid-template-columns: 2fr 80px 100px 160px 140px;
    align-items: center;
    padding: 10px 16px;
    border-bottom: 1px solid var(--border);
    font-size: 13px;
}

.backup-row:last-child {
    border-bottom: none;
}

.backup-row.header {
    font-weight: 600;
    color: var(--muted);
    font-size: 11px;
    letter-spacing: 0.5px;
    text-transform: uppercase;
    background: var(--bg3);
}

.backup-col.name {
    word-break: break-all;
}

.backup-col.actions {
    display: flex;
    gap: 6px;
    justify-content: flex-end;
}

.type-badge {
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 3px;
    font-weight: 700;
    letter-spacing: 0.5px;
    text-transform: uppercase;
}

.type-badge.full {
    background: rgba(234, 179, 8, 0.15);
    color: #eab308;
}

.type-badge.world {
    background: rgba(74, 222, 128, 0.15);
    color: var(--green);
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
    margin-bottom: 12px;
}

.modal-text {
    font-size: 13px;
    color: var(--text2);
    line-height: 1.5;
    margin-bottom: 20px;
}

.modal-actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
}

.config-section {
    margin-top: 24px;
}

.config-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid rgba(30, 45, 61, 0.4);
}

.config-row:last-child {
    border-bottom: none;
}

.config-import-row {
    display: flex;
    align-items: center;
    gap: 10px;
}

.file-label {
    cursor: pointer;
}

.config-result {
    padding: 12px 16px;
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
}

.result-label {
    font-size: 12px;
    color: #22c55e;
    font-weight: 600;
    letter-spacing: 0.5px;
    text-transform: uppercase;
}

.result-file {
    font-size: 12px;
    color: var(--text2);
    background: var(--bg);
    padding: 2px 8px;
    border-radius: 4px;
}
</style>
