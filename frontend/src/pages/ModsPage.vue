<template>
    <div class="tab-content">

        <!-- ══ CONFLICT WALL — plugins are installed ══ -->
        <div v-if="store.hasPlugins" class="conflict-wall">
            <div class="conflict-inner">
                <div class="conflict-icon">⚠</div>
                <div class="conflict-title">MOD LOADER CONFLICT DETECTED</div>
                <div class="conflict-body">
                    Your server has <strong>{{ store.installedPlugins.length }} plugin(s)</strong> installed
                    ({{ pluginLoaders }}).
                    Mods require a mod loader (Fabric / Forge / NeoForge), which is
                    <strong>incompatible with Bukkit-based plugin loaders</strong>.
                    Running both will cause your server to crash or behave unpredictably.
                </div>
                <div class="conflict-options">
                    <div class="conflict-option">
                        <div class="option-label">KEEP PLUGINS</div>
                        <div class="option-desc">Remove your mod loader and stay on a plugin-based server (Paper /
                            Spigot / Purpur).</div>
                        <button class="btn btn-outline btn-sm" @click="$emit('navigate', 'plugins')">← Go to Plugins
                            page</button>
                    </div>
                    <div class="conflict-divider">OR</div>
                    <div class="conflict-option">
                        <div class="option-label">SWITCH TO MODS</div>
                        <div class="option-desc">Uninstall all plugins first, then install a mod loader to use Fabric /
                            Forge mods.</div>
                        <button class="btn btn-danger btn-sm" @click="uninstallAllPlugins">✕ Uninstall all
                            plugins</button>
                    </div>
                </div>
                <div class="conflict-footer">
                    <span>💡 Hybrid solutions like Mohist or Arclight exist but are officially unsupported and
                        unstable.</span>
                </div>
            </div>
        </div>

        <!-- ══ NORMAL MOD MANAGER ══ -->
        <template v-else>

            <!-- Loader badge bar -->
            <div class="loader-bar">
                <div class="loader-badge" :class="store.installedModLoader?.toLowerCase() ?? 'none'">
                    <span class="loader-badge-icon">{{ loaderIcon }}</span>
                    <div>
                        <span class="loader-badge-name">{{ store.installedModLoader ?? 'No Mod Loader' }}</span>
                        <span class="loader-badge-sub">{{ store.installedModLoader ? 'Mod loader active' : 'Install a loader to use mods' }}</span>
                    </div>
                </div>
                <div class="loader-switcher">
                    <span class="loader-switch-label">Switch:</span>
                    <button v-for="l in modLoaders" :key="l"
                        :class="['loader-switch-btn', { active: store.installedModLoader === l }]"
                        @click="store.installedModLoader = l">{{ l }}</button>
                    <button :class="['loader-switch-btn', 'vanilla-btn', { active: store.installedModLoader === null }]"
                        @click="store.installedModLoader = null">Vanilla</button>
                </div>
            </div>

            <!-- Toolbar -->
            <div class="mods-toolbar">
                <div class="loader-pills">
                    <button v-for="l in loaderFilters" :key="l" :class="['loader-pill', { active: activeLoader === l }]"
                        @click="activeLoader = l">{{ l }}</button>
                </div>
                <div class="toolbar-right">
                    <button class="btn btn-sm btn-outline" @click="openFolder">📁 OPEN FOLDER</button>
                    <button class="btn btn-sm btn-outline" @click="checkUpdates">🔄 UPDATES</button>
                    <button class="btn btn-sm btn-primary"
                        @click="view = view === 'installed' ? 'browse' : 'installed'">
                        {{ view === 'installed' ? '＋ BROWSE' : '← INSTALLED' }}
                    </button>
                </div>
            </div>

            <!-- ── INSTALLED VIEW ── -->
            <div v-if="view === 'installed'">
                <div class="summary-bar">
                    <div class="summary-chip">
                        <span class="chip-val">{{ store.installedMods.length }}</span>
                        <span class="chip-lbl">Installed</span>
                    </div>
                    <div class="summary-chip warn" v-if="updatesAvailable > 0">
                        <span class="chip-val">{{ updatesAvailable }}</span>
                        <span class="chip-lbl">Updates</span>
                    </div>
                    <div class="summary-chip danger" v-if="errorsCount > 0">
                        <span class="chip-val">{{ errorsCount }}</span>
                        <span class="chip-lbl">Errors</span>
                    </div>
                    <div class="search-wrap">
                        <input v-model="installedSearch" class="search-input" placeholder="🔍 Filter mods..." />
                    </div>
                </div>

                <div class="card mods-list-card">
                    <div class="card-header">
                        <span class="card-title">INSTALLED MODS</span>
                        <div class="category-filters">
                            <button v-for="c in categoryFilters" :key="c"
                                :class="['cat-btn', { active: activeCategory === c }]" @click="activeCategory = c">{{ c
                                }}</button>
                        </div>
                    </div>
                    <div class="mods-list">
                        <transition-group name="mod-row">
                            <div v-for="mod in filteredInstalled" :key="mod.id" class="mod-row" :class="mod.status">
                                <div class="mod-icon">{{ mod.icon }}</div>
                                <div class="mod-meta">
                                    <div class="mod-name-row">
                                        <span class="mod-name">{{ mod.name }}</span>
                                        <span class="mod-version">v{{ mod.version }}</span>
                                        <span class="mod-loader-tag" :class="mod.loader.toLowerCase()">{{ mod.loader
                                        }}</span>
                                        <span class="mod-source-tag">{{ mod.source }}</span>
                                    </div>
                                    <span class="mod-desc">{{ mod.description }}</span>
                                    <span class="mod-file">{{ mod.fileName }} · {{ mod.fileSize }}</span>
                                </div>
                                <div class="mod-category-badge">{{ mod.category }}</div>
                                <div class="mod-status-col">
                                    <span class="mod-status-pill" :class="mod.status">{{ statusLabel(mod.status)
                                    }}</span>
                                </div>
                                <div class="mod-actions">
                                    <button v-if="mod.status === 'update-available'" class="tbl-btn update"
                                        @click="updateItem(mod)">↑ Update</button>
                                    <button class="tbl-btn" :disabled="mod.status === 'error'"
                                        @click="store.toggleMod(mod.id)">
                                        {{ mod.status === 'disabled' ? 'Enable' : 'Disable' }}
                                    </button>
                                    <button class="tbl-btn danger" @click="confirmTarget = mod">✕</button>
                                </div>
                            </div>
                        </transition-group>
                        <div v-if="filteredInstalled.length === 0" class="empty-state">
                            <span class="empty-icon">🧩</span>
                            <span>No mods installed yet</span>
                            <button class="btn btn-sm btn-primary" @click="view = 'browse'">Browse Mods</button>
                        </div>
                    </div>
                </div>

                <div class="drop-zone" :class="{ dragging: isDragging }" @dragover.prevent="isDragging = true"
                    @dragleave="isDragging = false" @drop.prevent="handleDrop">
                    <span class="drop-icon">📂</span>
                    <span>Drop <strong>.jar</strong> mod files here to install locally</span>
                </div>
            </div>

            <!-- ── BROWSE VIEW ── -->
            <div v-else>
                <div class="browse-search-row">
                    <div class="browse-search-wrap">
                        <span class="browse-search-icon">🔍</span>
                        <input v-model="browseQuery" class="browse-search-input"
                            placeholder="Search Modrinth & CurseForge for mods..." @keydown.enter="doSearch" />
                        <button class="btn btn-primary btn-sm" @click="doSearch" :disabled="isSearchingMods">
                            <span v-if="isSearchingMods" class="spinner">◌</span>
                            <span v-else>SEARCH</span>
                        </button>
                    </div>
                    <div class="source-filters">
                        <button v-for="s in ['All', 'Modrinth']" :key="s"
                            :class="['source-btn', { active: activeSource === s }]" @click="activeSource = s">{{ s
                            }}</button>
                    </div>
                </div>

                <div v-if="isSearchingMods" class="searching-state">
                    <span class="spinner-lg">◌</span><span>Searching mod repositories...</span>
                </div>

                <div v-else-if="searchResults.length > 0" class="browse-results">
                    <div v-for="result in filteredResults" :key="result.id" class="result-card">
                        <div class="result-icon">{{ result.icon }}</div>
                        <div class="result-meta">
                            <div class="result-name-row">
                                <span class="result-name">{{ result.title || result.name }}</span>
                                <span class="result-author">by {{ result.author }}</span>
                                <span class="result-source-tag">{{ result.source }}</span>
                            </div>
                            <span class="result-desc">{{ result.description }}</span>
                            <div class="result-tags">
                                <span class="result-tag">{{ result.categories?.[0] }}</span>
                                <span class="result-tag" v-for="l in result.loaders" :key="l"
                                    :class="(l || '').toLowerCase()">{{ l }}</span>
                            </div>
                        </div>
                        <div class="result-stats">
                            <span class="result-downloads">⬇ {{ result.downloads }}</span>
                            <span class="result-version">v{{ result.latest_version || result.latestVersion }}</span>
                        </div>
                        <div class="result-action">
                            <button v-if="result.installed" class="btn btn-sm btn-outline installed-btn" disabled>✓
                                Installed</button>
                            <button v-else class="btn btn-sm btn-primary" @click="installMod(result)">⬇ Install</button>
                        </div>
                    </div>
                </div>

                <div v-else class="browse-empty">
                    <div class="browse-empty-inner">
                        <span class="empty-icon">🔭</span>
                        <span class="browse-empty-title">SEARCH FOR MODS</span>
                        <span class="browse-empty-sub">Modrinth & CurseForge — Fabric, Forge, NeoForge</span>
                        <div class="featured-label">POPULAR MODS</div>
                        <div class="quick-searches">
                            <button v-for="q in quickSearches" :key="q" class="quick-btn" @click="quickSearch(q)">{{ q
                            }}</button>
                        </div>
                    </div>
                </div>
            </div>

        </template>

        <!-- Uninstall modal -->
        <div v-if="confirmTarget" class="modal-overlay" @click.self="confirmTarget = null">
            <div class="modal">
                <div class="modal-title">UNINSTALL MOD</div>
                <div class="modal-body">
                    Remove <strong>{{ confirmTarget.name }}</strong>?
                    <br><span class="modal-hint">The .jar file will be deleted from your mods folder.</span>
                </div>
                <div class="modal-actions">
                    <button class="btn btn-outline btn-sm" @click="confirmTarget = null">CANCEL</button>
                    <button class="btn btn-danger btn-sm" @click="doUninstall">UNINSTALL</button>
                </div>
            </div>
        </div>

    </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { api } from '../api'
import { store } from '../store'
import type { InstalledMod, ModSearchResult, ModLoaderType, ItemStatus } from '../store'

export default defineComponent({
    name: 'ModsPage',
    emits: ['toast', 'navigate'],
    data() {
        return {
            store,
            view: 'installed' as 'installed' | 'browse',
            activeLoader: 'All',
            activeCategory: 'All',
            activeSource: 'All',
            installedSearch: '',
            browseQuery: '',
            isDragging: false,
            confirmTarget: null as InstalledMod | null,
            modLoaders: ['Fabric', 'Forge', 'NeoForge'] as ModLoaderType[],
            loaderFilters: ['All', 'Fabric', 'Forge', 'NeoForge'],
            categoryFilters: ['All', 'Performance', 'Gameplay', 'Utility', 'World Generation'],
            quickSearches: ['Sodium', 'Lithium', 'Create', 'WorldEdit', 'Iris', 'Fabric API', 'Chunky', 'JEI'],
            searchResults: [] as any[],
            isSearchingMods: false,
        }
    },
    computed: {
        loaderIcon(): string {
            return ({ Fabric: '🧵', Forge: '🔨', NeoForge: '⚒' } as Record<string, string>)[this.store.installedModLoader ?? ''] ?? '🪨'
        },
        pluginLoaders(): string {
            return [...new Set(this.store.installedPlugins.map((p: any) => p.loader))].join(', ')
        },
        updatesAvailable(): number { return this.store.installedMods.filter((m: any) => m.status === 'update-available').length },
        errorsCount(): number { return this.store.installedMods.filter((m: any) => m.status === 'error').length },
        filteredInstalled(): InstalledMod[] {
            return this.store.installedMods.filter((mod: any) => {
                const matchLoader = this.activeLoader === 'All' || mod.loader === this.activeLoader
                const matchCategory = this.activeCategory === 'All' || mod.category === this.activeCategory
                const matchSearch = !this.installedSearch || mod.name.toLowerCase().includes(this.installedSearch.toLowerCase())
                return matchLoader && matchCategory && matchSearch
            })
        },
        filteredResults(): any[] {
            let results = this.searchResults
            if (this.activeSource !== 'All') {
                results = results.filter((r: any) => r.source === this.activeSource)
            }
            if (this.activeLoader !== 'All') {
                results = results.filter((r: any) => r.loaders?.includes(this.activeLoader))
            }
            return results
        },
    },
    methods: {
        statusLabel(status: ItemStatus): string {
            return ({ enabled: '● Active', disabled: '○ Disabled', error: '✕ Error', 'update-available': '↑ Update' } as Record<ItemStatus, string>)[status]
        },
        async doSearch(): Promise<void> {
            if (!this.browseQuery.trim()) return
            this.isSearchingMods = true
            this.searchResults = []
            try {
                const activeBuild = this.store.activeServerBuild
                const mcVersion = activeBuild?.mcVersion || '1.21'
                const loader = this.activeLoader !== 'All' ? [this.activeLoader.toLowerCase()] : undefined
                const results = await api.searchMods(this.browseQuery, loader, mcVersion)
                this.searchResults = results.map((r: any) => ({
                    ...r,
                    installed: this.store.installedMods.some((m: any) => m.id === r.id),
                    icon: this.getModIcon(r),
                }))
            } catch (e: any) {
                this.$emit('toast', { msg: `Search failed: ${e}`, type: 'danger' })
            } finally {
                this.isSearchingMods = false
            }
        },
        getModIcon(r: any): string {
            const icons: Record<string, string> = { 'Performance': '⚡', 'Gameplay': '🎮', 'Utility': '🔧', 'World Generation': '🗺', 'Admin': '🛡', 'Economy': '💰', 'Chat': '💬', 'Protection': '🛡' }
            const cat = r.categories?.[0] || 'Utility'
            return icons[cat] || '🧩'
        },
        quickSearch(q: string): void { this.browseQuery = q; this.doSearch() },
        async installMod(result: any): Promise<void> {
            try {
                this.$emit('toast', { msg: `Downloading ${result.title || result.name}...`, type: 'success' })
                await api.downloadMod(result.id)
                const existing = this.store.installedMods.find((m: any) => m.id === result.id)
                if (!existing) {
                    this.store.installedMods.push({
                        id: result.id,
                        name: result.title || result.name,
                        version: result.latest_version || result.latestVersion,
                        latestVersion: result.latest_version || result.latestVersion,
                        author: result.author,
                        description: result.description,
                        category: result.categories?.[0] || 'Utility',
                        loader: result.loaders?.[0] || 'Fabric',
                        fileSize: '—',
                        status: 'enabled',
                        source: result.source,
                        icon: this.getModIcon(result),
                        fileName: `${result.id}.jar`,
                    })
                }
                this.searchResults = this.searchResults.map((r: any) => r.id === result.id ? { ...r, installed: true } : r)
                this.$emit('toast', { msg: `${result.title || result.name} installed!`, type: 'success' })
            } catch (e: any) {
                this.$emit('toast', { msg: `Install failed: ${e}`, type: 'danger' })
            }
        },
        async updateItem(mod: InstalledMod): Promise<void> {
            try {
                this.$emit('toast', { msg: `Downloading ${mod.name} v${mod.latestVersion}...`, type: 'success' })
                await api.updateMod(mod.id, mod.fileName)
                mod.version = mod.latestVersion
                mod.status = 'enabled'
                this.$emit('toast', { msg: `${mod.name} updated to v${mod.latestVersion}`, type: 'success' })
            } catch (e: any) {
                this.$emit('toast', { msg: `Update failed: ${e}`, type: 'danger' })
            }
        },
        async doUninstall(): Promise<void> {
            if (!this.confirmTarget) return
            try {
                await api.deleteMod(this.confirmTarget.fileName)
                this.store.installedMods = this.store.installedMods.filter((m: any) => m.id !== this.confirmTarget?.id)
                this.$emit('toast', { msg: `Removed ${this.confirmTarget.name}`, type: 'danger' })
            } catch (e: any) {
                this.$emit('toast', { msg: `Uninstall failed: ${e}`, type: 'danger' })
            }
            this.confirmTarget = null
        },
        async openFolder(): Promise<void> {
            try {
                const dir = await api.getServerDirPath()
                await api.openFolder(`${dir}/mods`)
            } catch (e: any) {
                this.$emit('toast', { msg: `Could not open folder: ${e}`, type: 'danger' })
            }
        },
        async checkUpdates(): Promise<void> {
            this.$emit('toast', { msg: 'Checking for updates...', type: 'success' })
            try {
                const items = await api.checkModUpdates()
                for (const item of items) {
                    const existing = this.store.installedMods.find(m => m.id === item.file_name)
                    if (existing) {
                        existing.latestVersion = item.latestVersion
                        existing.status = item.hasUpdate ? 'update-available' : 'enabled'
                    }
                }
                this.$emit('toast', { msg: `${this.updatesAvailable} update(s) available`, type: this.updatesAvailable > 0 ? 'warn' : 'success' })
            } catch (e: any) {
                this.$emit('toast', { msg: `Update check failed: ${e}`, type: 'danger' })
            }
        },
        handleDrop(e: DragEvent): void {
            this.isDragging = false
            const jars = Array.from(e.dataTransfer?.files ?? []).filter(f => f.name.endsWith('.jar'))
            if (!jars.length) { this.$emit('toast', { msg: 'Only .jar files supported', type: 'danger' }); return }
            jars.forEach(f => { this.$emit('toast', { msg: `Installing ${f.name}...`, type: 'success' }); this.store.addLog('INFO', 'info', `Local mod install: ${f.name}`) })
        },
        uninstallAllPlugins(): void {
            this.store.installedPlugins = []
            this.store.addLog('INFO', 'warn', 'Removed all plugins to allow mod loader')
            this.$emit('toast', { msg: 'All plugins removed', type: 'danger' })
        },
    },

    mounted() {
        store.fetchInstalledMods()
    },
})
</script>

<style scoped>
/* ── Conflict Wall ── */
.conflict-wall {
    background: rgba(248, 113, 113, 0.04);
    border: 1px solid rgba(248, 113, 113, 0.3);
    border-radius: var(--radius);
    padding: 32px;
    margin-bottom: 20px;
}

.conflict-inner {
    max-width: 680px;
    margin: 0 auto;
}

.conflict-icon {
    font-size: 36px;
    margin-bottom: 10px;
}

.conflict-title {
    font-family: 'VT323', monospace;
    font-size: 24px;
    letter-spacing: 3px;
    color: var(--red);
    margin-bottom: 12px;
}

.conflict-body {
    font-size: 13px;
    color: var(--text2);
    line-height: 1.7;
    margin-bottom: 24px;
}

.conflict-body strong {
    color: var(--text);
}

.conflict-options {
    display: flex;
    align-items: center;
    gap: 24px;
    margin-bottom: 20px;
}

.conflict-option {
    flex: 1;
}

.option-label {
    font-size: 12px;
    letter-spacing: 2px;
    color: var(--text);
    margin-bottom: 6px;
}

.option-desc {
    font-size: 12px;
    color: var(--muted);
    margin-bottom: 12px;
    line-height: 1.5;
}

.conflict-divider {
    font-family: 'VT323', monospace;
    font-size: 20px;
    color: var(--muted);
    padding: 0 8px;
    flex-shrink: 0;
}

.conflict-footer {
    font-size: 11px;
    color: var(--muted);
    border-top: 1px solid var(--border);
    padding-top: 14px;
}

/* ── Loader bar ── */
.loader-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 12px 16px;
    margin-bottom: 16px;
    gap: 16px;
    flex-wrap: wrap;
}

.loader-badge {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 14px;
    border-radius: var(--radius);
    border: 1px solid var(--border);
}

.loader-badge.fabric {
    border-color: #60a5fa55;
    background: rgba(96, 165, 250, 0.06);
}

.loader-badge.forge {
    border-color: #fb923c55;
    background: rgba(251, 146, 60, 0.06);
}

.loader-badge.neoforge {
    border-color: #a78bfa55;
    background: rgba(167, 139, 250, 0.06);
}

.loader-badge-icon {
    font-size: 22px;
}

.loader-badge-name {
    display: block;
    font-size: 14px;
    font-weight: 600;
}

.loader-badge.fabric .loader-badge-name {
    color: var(--blue);
}

.loader-badge.forge .loader-badge-name {
    color: #fb923c;
}

.loader-badge.neoforge .loader-badge-name {
    color: var(--purple);
}

.loader-badge-sub {
    display: block;
    font-size: 10px;
    color: var(--muted);
}

.loader-switcher {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
}

.loader-switch-label {
    font-size: 11px;
    color: var(--muted);
    margin-right: 4px;
}

.loader-switch-btn {
    padding: 4px 12px;
    background: none;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--muted);
    cursor: pointer;
    font-family: 'Share Tech Mono', monospace;
    font-size: 11px;
    transition: all 0.15s;
}

.loader-switch-btn.active {
    border-color: var(--green);
    color: var(--green);
}

.loader-switch-btn.vanilla-btn.active {
    border-color: var(--muted);
    color: var(--muted);
}

.loader-switch-btn:hover:not(.active) {
    border-color: var(--text2);
    color: var(--text);
}

/* ── Toolbar ── */
.mods-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    gap: 12px;
    flex-wrap: wrap;
}

.loader-pills {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
}

.loader-pill {
    padding: 4px 12px;
    background: none;
    border: 1px solid var(--border);
    border-radius: 20px;
    color: var(--muted);
    cursor: pointer;
    font-family: 'Share Tech Mono', monospace;
    font-size: 11px;
    transition: all 0.15s;
}

.loader-pill.active {
    border-color: var(--green);
    color: var(--green);
    background: rgba(74, 222, 128, 0.07);
}

.toolbar-right {
    display: flex;
    gap: 8px;
}

/* ── Summary ── */
.summary-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 14px;
    flex-wrap: wrap;
}

.summary-chip {
    display: flex;
    align-items: center;
    gap: 6px;
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 6px 12px;
}

.summary-chip.warn {
    border-color: rgba(251, 191, 36, 0.4);
}

.summary-chip.danger {
    border-color: rgba(248, 113, 113, 0.4);
}

.chip-val {
    font-family: 'VT323', monospace;
    font-size: 20px;
    line-height: 1;
}

.chip-lbl {
    font-size: 11px;
    color: var(--muted);
}

.summary-chip.warn .chip-val {
    color: var(--yellow);
}

.summary-chip.danger .chip-val {
    color: var(--red);
}

.search-wrap {
    flex: 1;
    min-width: 160px;
}

/* ── Category ── */
.category-filters {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
}

.cat-btn {
    padding: 3px 9px;
    background: none;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--muted);
    cursor: pointer;
    font-family: 'Share Tech Mono', monospace;
    font-size: 10px;
    transition: all 0.15s;
}

.cat-btn.active {
    border-color: var(--green);
    color: var(--green);
}

/* ── Mod rows ── */
.mods-list-card {
    margin-bottom: 16px;
}

.mods-list {
    padding: 6px 0;
    min-height: 80px;
}

.mod-row {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 10px 16px;
    border-bottom: 1px solid rgba(30, 45, 61, 0.4);
    transition: background 0.1s;
}

.mod-row:last-child {
    border-bottom: none;
}

.mod-row:hover {
    background: var(--bg3);
}

.mod-row.disabled {
    opacity: 0.5;
}

.mod-row.error {
    border-left: 3px solid var(--red);
}

.mod-icon {
    font-size: 22px;
    flex-shrink: 0;
    width: 32px;
    text-align: center;
}

.mod-meta {
    flex: 1;
    min-width: 0;
}

.mod-name-row {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    margin-bottom: 2px;
}

.mod-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--text);
}

.mod-version {
    font-size: 11px;
    color: var(--muted);
}

.mod-desc {
    display: block;
    font-size: 11px;
    color: var(--text2);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 400px;
}

.mod-file {
    display: block;
    font-size: 10px;
    color: var(--muted);
    margin-top: 2px;
}

.mod-loader-tag {
    font-size: 10px;
    padding: 1px 6px;
    border-radius: 3px;
    border: 1px solid var(--border);
    color: var(--muted);
}

.mod-loader-tag.fabric {
    border-color: #60a5fa44;
    color: var(--blue);
}

.mod-loader-tag.forge {
    border-color: #fb923c44;
    color: #fb923c;
}

.mod-loader-tag.neoforge {
    border-color: #a78bfa44;
    color: var(--purple);
}

.mod-source-tag {
    font-size: 10px;
    color: var(--muted);
}

.mod-category-badge {
    font-size: 10px;
    padding: 2px 8px;
    background: var(--bg3);
    border: 1px solid var(--border);
    border-radius: 3px;
    color: var(--muted);
    white-space: nowrap;
    flex-shrink: 0;
}

.mod-status-col {
    width: 96px;
    flex-shrink: 0;
}

.mod-status-pill {
    font-size: 11px;
    padding: 3px 8px;
    border-radius: 3px;
    display: inline-block;
    white-space: nowrap;
}

.mod-status-pill.enabled {
    background: rgba(74, 222, 128, 0.1);
    color: var(--green);
}

.mod-status-pill.disabled {
    background: rgba(75, 85, 99, 0.2);
    color: var(--muted);
}

.mod-status-pill.error {
    background: rgba(248, 113, 113, 0.1);
    color: var(--red);
}

.mod-status-pill.update-available {
    background: rgba(251, 191, 36, 0.1);
    color: var(--yellow);
}

.mod-actions {
    display: flex;
    gap: 5px;
    flex-shrink: 0;
}

.tbl-btn.update {
    border-color: rgba(251, 191, 36, 0.4);
    color: var(--yellow);
}

.tbl-btn.update:hover {
    border-color: var(--yellow);
}

/* ── Drop zone ── */
.drop-zone {
    border: 2px dashed var(--border);
    border-radius: var(--radius);
    padding: 20px;
    text-align: center;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    color: var(--muted);
    font-size: 13px;
    transition: all 0.2s;
}

.drop-zone.dragging {
    border-color: var(--green);
    color: var(--green);
    background: rgba(74, 222, 128, 0.05);
}

.drop-icon {
    font-size: 20px;
}

/* ── Browse ── */
.browse-search-row {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-bottom: 18px;
}

.browse-search-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 0 14px;
    transition: border-color 0.15s;
}

.browse-search-wrap:focus-within {
    border-color: var(--green);
}

.browse-search-icon {
    color: var(--muted);
    font-size: 14px;
    flex-shrink: 0;
}

.browse-search-input {
    flex: 1;
    background: none;
    border: none;
    outline: none;
    color: var(--text);
    font-family: 'Share Tech Mono', monospace;
    font-size: 13px;
    padding: 10px 0;
}

.browse-search-input::placeholder {
    color: var(--muted);
}

.source-filters {
    display: flex;
    gap: 6px;
}

.source-btn {
    padding: 4px 12px;
    background: none;
    border: 1px solid var(--border);
    border-radius: 20px;
    color: var(--muted);
    cursor: pointer;
    font-family: 'Share Tech Mono', monospace;
    font-size: 11px;
    transition: all 0.15s;
}

.source-btn.active {
    border-color: var(--blue);
    color: var(--blue);
    background: rgba(96, 165, 250, 0.07);
}

.browse-results {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.result-card {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 14px 16px;
    display: flex;
    align-items: center;
    gap: 14px;
    transition: border-color 0.15s;
}

.result-card:hover {
    border-color: var(--blue);
}

.result-icon {
    font-size: 24px;
    flex-shrink: 0;
    width: 36px;
    text-align: center;
}

.result-meta {
    flex: 1;
    min-width: 0;
}

.result-name-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 3px;
    flex-wrap: wrap;
}

.result-name {
    font-size: 14px;
    font-weight: 600;
}

.result-author {
    font-size: 11px;
    color: var(--muted);
}

.result-source-tag {
    font-size: 10px;
    color: var(--muted);
}

.result-desc {
    display: block;
    font-size: 11px;
    color: var(--text2);
    margin-bottom: 5px;
}

.result-tags {
    display: flex;
    gap: 5px;
    flex-wrap: wrap;
}

.result-tag {
    font-size: 10px;
    padding: 1px 7px;
    background: var(--bg3);
    border: 1px solid var(--border);
    border-radius: 3px;
    color: var(--muted);
}

.result-tag.fabric {
    border-color: #60a5fa44;
    color: var(--blue);
}

.result-tag.forge {
    border-color: #fb923c44;
    color: #fb923c;
}

.result-tag.neoforge {
    border-color: #a78bfa44;
    color: var(--purple);
}

.result-stats {
    text-align: right;
    flex-shrink: 0;
}

.result-downloads {
    display: block;
    font-size: 12px;
    color: var(--text2);
}

.result-version {
    display: block;
    font-size: 11px;
    color: var(--muted);
    margin-top: 2px;
}

.result-action {
    flex-shrink: 0;
}

.installed-btn {
    opacity: 0.5;
}

.browse-empty {
    display: flex;
    justify-content: center;
    padding: 40px 0;
}

.browse-empty-inner {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    text-align: center;
}

.browse-empty-title {
    font-family: 'VT323', monospace;
    font-size: 24px;
    letter-spacing: 3px;
    color: var(--text);
    margin-top: 8px;
}

.browse-empty-sub {
    font-size: 12px;
    color: var(--muted);
}

.featured-label {
    font-size: 10px;
    letter-spacing: 2px;
    color: var(--muted);
    margin-top: 16px;
}

.quick-searches {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    justify-content: center;
    margin-top: 6px;
}

.quick-btn {
    padding: 5px 14px;
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: 20px;
    color: var(--text2);
    cursor: pointer;
    font-family: 'Share Tech Mono', monospace;
    font-size: 12px;
    transition: all 0.15s;
}

.quick-btn:hover {
    border-color: var(--green);
    color: var(--green);
}

.searching-state {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 48px;
    color: var(--muted);
}

.spinner-lg {
    font-size: 28px;
    animation: spin 1s linear infinite;
    display: inline-block;
}

.empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 32px;
    color: var(--muted);
}

.empty-icon {
    font-size: 32px;
}

/* ── Modal ── */
.modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    z-index: 300;
    display: flex;
    align-items: center;
    justify-content: center;
}

.modal {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 24px;
    width: 380px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.6);
}

.modal-title {
    font-family: 'VT323', monospace;
    font-size: 20px;
    letter-spacing: 2px;
    color: var(--red);
    margin-bottom: 14px;
}

.modal-body {
    font-size: 13px;
    color: var(--text2);
    line-height: 1.6;
    margin-bottom: 20px;
}

.modal-body strong {
    color: var(--text);
}

.modal-hint {
    font-size: 11px;
    color: var(--muted);
}

.modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
}

.mod-row-enter-active,
.mod-row-leave-active {
    transition: all 0.25s;
}

.mod-row-enter-from,
.mod-row-leave-to {
    opacity: 0;
    transform: translateX(-8px);
}
</style>