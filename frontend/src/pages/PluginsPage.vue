<template>
    <div class="tab-content">

        <!-- ══ CONFLICT WALL — mod loader is installed ══ -->
        <div v-if="store.hasModLoader" class="conflict-wall">
            <div class="conflict-inner">
                <div class="conflict-icon">⚠</div>
                <div class="conflict-title">MOD LOADER CONFLICT DETECTED</div>
                <div class="conflict-body">
                    Your server has <strong>{{ store.installedModLoader }}</strong> installed as a mod loader.
                    Bukkit-based plugins (Paper / Spigot / Purpur) are
                    <strong>incompatible with Fabric / Forge / NeoForge mod loaders</strong>.
                    Installing plugins alongside mods will prevent your server from starting.
                </div>

                <!-- show any lingering plugins already installed -->
                <div v-if="store.hasMods" class="existing-mods-warning">
                    <span class="warn-label">INSTALLED MODS ({{ store.installedMods.length }})</span>
                    <div class="existing-list">
                        <span v-for="m in store.installedMods" :key="m.id" class="existing-chip">
                            {{ m.icon }} {{ m.name }}
                        </span>
                    </div>
                </div>

                <div class="conflict-options">
                    <div class="conflict-option">
                        <div class="option-label">KEEP MODS</div>
                        <div class="option-desc">Stay on your current mod loader. Remove the mod loader to switch to a
                            plugin-based server.</div>
                        <button class="btn btn-outline btn-sm" @click="$emit('navigate', 'mods')">← Go to Mods
                            page</button>
                    </div>
                    <div class="conflict-divider">OR</div>
                    <div class="conflict-option">
                        <div class="option-label">SWITCH TO PLUGINS</div>
                        <div class="option-desc">Remove your mod loader and all mods, then switch to Paper / Spigot /
                            Purpur.</div>
                        <button class="btn btn-danger btn-sm" @click="switchToPlugins">✕ Remove mod loader &
                            mods</button>
                    </div>
                </div>
                <div class="conflict-footer">
                    💡 Want both mods and plugins? Consider <strong>Purpur</strong> with Fabric via
                    <strong>Cardboard</strong> — though stability is not guaranteed.
                </div>
            </div>
        </div>

        <!-- ══ NORMAL PLUGIN MANAGER ══ -->
        <template v-else>

            <!-- Server type bar -->
            <div class="loader-bar">
                <div class="loader-badge paper">
                    <span class="loader-badge-icon">📄</span>
                    <div>
                        <span class="loader-badge-name">Plugin Server</span>
                        <span class="loader-badge-sub">Paper / Spigot / Purpur — Bukkit API</span>
                    </div>
                </div>
                <div class="loader-switcher">
                    <span class="loader-switch-label">Server software:</span>
                    <button v-for="s in pluginLoaders" :key="s"
                        :class="['loader-switch-btn', { active: activeServerSoftware === s }]"
                        @click="activeServerSoftware = s">{{ s }}</button>
                </div>
            </div>

            <!-- Toolbar -->
            <div class="mods-toolbar">
                <div class="loader-pills">
                    <button v-for="c in categoryFilters" :key="c"
                        :class="['loader-pill', { active: activeCategory === c }]" @click="activeCategory = c">{{ c
                        }}</button>
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
                        <span class="chip-val">{{ store.installedPlugins.length }}</span>
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
                        <input v-model="installedSearch" class="search-input" placeholder="🔍 Filter plugins..." />
                    </div>
                </div>

                <div class="card mods-list-card">
                    <div class="card-header">
                        <span class="card-title">INSTALLED PLUGINS</span>
                        <div class="category-filters">
                            <button v-for="c in allCategoryFilters" :key="c"
                                :class="['cat-btn', { active: activeCatFilter === c }]" @click="activeCatFilter = c">{{
                                c }}</button>
                        </div>
                    </div>
                    <div class="mods-list">
                        <transition-group name="mod-row">
                            <div v-for="plugin in filteredInstalled" :key="plugin.id" class="mod-row"
                                :class="plugin.status">
                                <div class="mod-icon">{{ plugin.icon }}</div>
                                <div class="mod-meta">
                                    <div class="mod-name-row">
                                        <span class="mod-name">{{ plugin.name }}</span>
                                        <span class="mod-version">v{{ plugin.version }}</span>
                                        <span class="mod-loader-tag" :class="plugin.loader.toLowerCase()">{{
                                            plugin.loader }}</span>
                                        <span class="mod-source-tag">{{ plugin.source }}</span>
                                    </div>
                                    <span class="mod-desc">{{ plugin.description }}</span>
                                    <span class="mod-file">{{ plugin.fileName }} · {{ plugin.fileSize }}</span>
                                </div>
                                <div class="mod-category-badge">{{ plugin.category }}</div>
                                <div class="mod-status-col">
                                    <span class="mod-status-pill" :class="plugin.status">{{ statusLabel(plugin.status)
                                        }}</span>
                                </div>
                                <div class="mod-actions">
                                    <button v-if="plugin.status === 'update-available'" class="tbl-btn update"
                                        @click="updateItem(plugin)">↑ Update</button>
                                    <button class="tbl-btn" :disabled="plugin.status === 'error'"
                                        @click="store.togglePlugin(plugin.id)">
                                        {{ plugin.status === 'disabled' ? 'Enable' : 'Disable' }}
                                    </button>
                                    <button class="tbl-btn danger" @click="confirmTarget = plugin">✕</button>
                                </div>
                            </div>
                        </transition-group>
                        <div v-if="filteredInstalled.length === 0" class="empty-state">
                            <span class="empty-icon">🔌</span>
                            <span>No plugins installed yet</span>
                            <button class="btn btn-sm btn-primary" @click="view = 'browse'">Browse Plugins</button>
                        </div>
                    </div>
                </div>

                <div class="drop-zone" :class="{ dragging: isDragging }" @dragover.prevent="isDragging = true"
                    @dragleave="isDragging = false" @drop.prevent="handleDrop">
                    <span class="drop-icon">📂</span>
                    <span>Drop <strong>.jar</strong> plugin files here to install locally</span>
                </div>
            </div>

            <!-- ── BROWSE VIEW ── -->
            <div v-else>
                <div class="browse-search-row">
                    <div class="browse-search-wrap">
                        <span class="browse-search-icon">🔍</span>
                        <input v-model="browseQuery" class="browse-search-input"
                            placeholder="Search Hangar, CurseForge & Modrinth for plugins..."
                            @keydown.enter="doSearch" />
                        <button class="btn btn-primary btn-sm" @click="doSearch" :disabled="isSearchingPlugins">
                            <span v-if="isSearchingPlugins" class="spinner">◌</span>
                            <span v-else>SEARCH</span>
                        </button>
                    </div>
                    <div class="source-filters">
                        <button v-for="s in ['All', 'Hangar', 'Modrinth', 'CurseForge']" :key="s"
                            :class="['source-btn', { active: activeSource === s }]" @click="activeSource = s">{{ s
                            }}</button>
                    </div>
                </div>

                <div v-if="isSearchingPlugins" class="searching-state">
                    <span class="spinner-lg">◌</span><span>Searching plugin repositories...</span>
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
                            <button v-else class="btn btn-sm btn-primary" @click="installPlugin(result)">⬇
                                Install</button>
                        </div>
                    </div>
                </div>

                <div v-else-if="popularResults.length > 0 && !browseQuery.trim()" class="browse-results">
                    <div class="popular-section-label">POPULAR PLUGINS</div>
                    <div v-for="result in popularResults" :key="result.id" class="result-card">
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
                            <button v-else class="btn btn-sm btn-primary" @click="installPlugin(result)">⬇
                                Install</button>
                        </div>
                    </div>
                </div>

                <div v-else-if="isLoadingPopular" class="searching-state">
                    <span class="spinner-lg">◌</span><span>Loading popular plugins...</span>
                </div>

                <div v-else class="browse-empty">
                    <div class="browse-empty-inner">
                        <span class="empty-icon">🔌</span>
                        <span class="browse-empty-title">SEARCH FOR PLUGINS</span>
                        <span class="browse-empty-sub">Hangar, Modrinth & CurseForge — Paper, Spigot, Purpur</span>
                        <div class="featured-label">QUICK SEARCH</div>
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
                <div class="modal-title">UNINSTALL PLUGIN</div>
                <div class="modal-body">
                    Remove <strong>{{ confirmTarget.name }}</strong>?
                    <br><span class="modal-hint">The .jar file will be deleted from your plugins folder.</span>
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
import type { InstalledPlugin, PluginSearchResult, PluginLoaderType, ItemStatus } from '../store'

export default defineComponent({
    name: 'PluginsPage',
    emits: ['toast', 'navigate'],
    data() {
        return {
            store,
            view: 'installed' as 'installed' | 'browse',
            activeCategory: 'All',
            activeCatFilter: 'All',
            activeSource: 'All',
            activeServerSoftware: 'Paper' as PluginLoaderType,
            installedSearch: '',
            browseQuery: '',
            isDragging: false,
            confirmTarget: null as InstalledPlugin | null,
            pluginLoaders: ['Paper', 'Spigot', 'Purpur'] as PluginLoaderType[],
            categoryFilters: ['All', 'Utility', 'Admin', 'Protection', 'Economy', 'Chat'],
            allCategoryFilters: ['All', 'Utility', 'Admin', 'Protection', 'Economy', 'Chat', 'World Generation'],
            quickSearches: ['EssentialsX', 'LuckPerms', 'WorldGuard', 'Vault', 'Dynmap', 'CoreProtect', 'Chunky', 'ShopGUI+'],
            searchResults: [] as any[],
            isSearchingPlugins: false,
            popularResults: [] as any[],
            isLoadingPopular: false,
        }
    },
    computed: {
        updatesAvailable(): number { return this.store.installedPlugins.filter(p => p.status === 'update-available').length },
        errorsCount(): number { return this.store.installedPlugins.filter(p => p.status === 'error').length },
        filteredInstalled(): InstalledPlugin[] {
            return this.store.installedPlugins.filter(p => {
                const matchCat = this.activeCatFilter === 'All' || p.category === this.activeCatFilter
                const matchSearch = !this.installedSearch || p.name.toLowerCase().includes(this.installedSearch.toLowerCase())
                return matchCat && matchSearch
            })
        },
        filteredResults(): any[] {
            return this.searchResults.filter((r: any) => this.activeSource === 'All' || r.source === this.activeSource)
        },
    },
    methods: {
        statusLabel(status: ItemStatus): string {
            return ({ enabled: '● Active', disabled: '○ Disabled', error: '✕ Error', 'update-available': '↑ Update' } as Record<ItemStatus, string>)[status]
        },
        async doSearch(): Promise<void> {
            if (!this.browseQuery.trim()) return
            this.isSearchingPlugins = true
            this.searchResults = []
            try {
                const results = await api.searchPlugins(this.browseQuery)
                this.searchResults = results.map((r: any) => ({
                    ...r,
                    installed: this.store.installedPlugins.some((p: any) => p.id === r.id),
                    icon: this.getPluginIcon(r),
                }))
            } catch (e: any) {
                this.$emit('toast', { msg: `Search failed: ${e}`, type: 'danger' })
            } finally {
                this.isSearchingPlugins = false
            }
        },
        getPluginIcon(r: any): string {
            const icons: Record<string, string> = { 'Performance': '⚡', 'Gameplay': '🎮', 'Utility': '🔧', 'Protection': '🛡', 'Economy': '💰', 'Chat': '💬', 'Admin': '🛡', 'World Generation': '🗺' }
            const cat = r.categories?.[0] || 'Utility'
            return icons[cat] || '🔌'
        },
        quickSearch(q: string): void { this.browseQuery = q; this.doSearch() },
        async fetchPopularPlugins(): Promise<void> {
            if (this.popularResults.length > 0) return
            this.isLoadingPopular = true
            try {
                const results = await api.getPopularPlugins()
                this.popularResults = results.map((r: any) => ({
                    ...r,
                    installed: this.store.installedPlugins.some((p: any) => p.id === r.id),
                    icon: this.getPluginIcon(r),
                }))
            } catch (e) {
                console.error('Failed to load popular plugins:', e)
            } finally {
                this.isLoadingPopular = false
            }
        },
        async installPlugin(result: any): Promise<void> {
            try {
                this.$emit('toast', { msg: `Downloading ${result.title || result.name}...`, type: 'success' })
                await api.downloadPlugin(result.slug || result.id, undefined, result.source)
                const existing = this.store.installedPlugins.find((p: any) => p.id === result.id)
                if (!existing) {
                    this.store.installedPlugins.push({
                        id: result.id,
                        name: result.title || result.name,
                        version: result.latest_version || result.latestVersion,
                        latestVersion: result.latest_version || result.latestVersion,
                        author: result.author,
                        description: result.description,
                        category: result.categories?.[0] || 'Utility',
                        loader: result.loaders?.[0] || 'Paper',
                        fileSize: '—',
                        status: 'enabled',
                        source: result.source,
                        icon: this.getPluginIcon(result),
                        fileName: `${result.slug || result.id}.jar`,
                    })
                }
                this.searchResults = this.searchResults.map((r: any) => r.id === result.id ? { ...r, installed: true } : r)
                this.popularResults = this.popularResults.map((r: any) => r.id === result.id ? { ...r, installed: true } : r)
                this.$emit('toast', { msg: `${result.title || result.name} installed!`, type: 'success' })
            } catch (e: any) {
                this.$emit('toast', { msg: `Install failed: ${e}`, type: 'danger' })
            }
        },
        async updateItem(plugin: InstalledPlugin): Promise<void> {
            try {
                this.$emit('toast', { msg: `Downloading ${plugin.name} v${plugin.latestVersion}...`, type: 'success' })
                await api.updatePlugin(plugin.id, plugin.fileName, plugin.source)
                plugin.version = plugin.latestVersion
                plugin.status = 'enabled'
                this.$emit('toast', { msg: `${plugin.name} updated to v${plugin.latestVersion}`, type: 'success' })
            } catch (e: any) {
                this.$emit('toast', { msg: `Update failed: ${e}`, type: 'danger' })
            }
        },
        async doUninstall(): Promise<void> {
            if (!this.confirmTarget) return
            try {
                await api.deletePlugin(this.confirmTarget.fileName)
                this.store.installedPlugins = this.store.installedPlugins.filter((p: any) => p.id !== this.confirmTarget?.id)
                this.$emit('toast', { msg: `Removed ${this.confirmTarget.name}`, type: 'danger' })
            } catch (e: any) {
                this.$emit('toast', { msg: `Uninstall failed: ${e}`, type: 'danger' })
            }
            this.confirmTarget = null
        },
        async openFolder(): Promise<void> {
            try {
                const dir = await api.getServerDirPath()
                await api.openFolder(`${dir}/plugins`)
            } catch (e: any) {
                this.$emit('toast', { msg: `Could not open folder: ${e}`, type: 'danger' })
            }
        },
        async checkUpdates(): Promise<void> {
            this.$emit('toast', { msg: 'Checking for updates...', type: 'success' })
            try {
                const items = await api.checkPluginUpdates()
                for (const item of items) {
                    const existing = this.store.installedPlugins.find(p => p.id === item.file_name)
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
            jars.forEach(f => { this.$emit('toast', { msg: `Installing ${f.name}...`, type: 'success' }); this.store.addLog('INFO', 'info', `Local plugin install: ${f.name}`) })
        },
        switchToPlugins(): void {
            this.store.installedMods = []
            this.store.installedModLoader = null
            this.store.addLog('INFO', 'warn', 'Removed mod loader and mods to switch to plugin server')
            this.$emit('toast', { msg: 'Switched to plugin server mode', type: 'success' })
        },
    },

    mounted() {
        store.fetchInstalledPlugins()
        this.fetchPopularPlugins()
    },
})
</script>

<style scoped>
/* ── Conflict Wall ── */
.conflict-wall {
    background: rgba(251, 191, 36, 0.04);
    border: 1px solid rgba(251, 191, 36, 0.3);
    border-radius: var(--radius);
    padding: 32px;
    margin-bottom: 20px;
}

.conflict-inner {
    max-width: 700px;
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
    color: var(--yellow);
    margin-bottom: 12px;
}

.conflict-body {
    font-size: 13px;
    color: var(--text2);
    line-height: 1.7;
    margin-bottom: 16px;
}

.conflict-body strong {
    color: var(--text);
}

.existing-mods-warning {
    background: var(--bg3);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 10px 14px;
    margin-bottom: 20px;
}

.warn-label {
    display: block;
    font-size: 10px;
    letter-spacing: 2px;
    color: var(--yellow);
    margin-bottom: 8px;
}

.existing-list {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
}

.existing-chip {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: 3px;
    padding: 2px 10px;
    font-size: 12px;
    color: var(--text2);
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

.conflict-footer strong {
    color: var(--text2);
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

.loader-badge.paper {
    border-color: #f8717155;
    background: rgba(248, 113, 113, 0.06);
}

.loader-badge-icon {
    font-size: 22px;
}

.loader-badge-name {
    display: block;
    font-size: 14px;
    font-weight: 600;
    color: var(--red);
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
    border-color: var(--red);
    color: var(--red);
    background: rgba(248, 113, 113, 0.07);
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
    border-color: var(--red);
    color: var(--red);
    background: rgba(248, 113, 113, 0.07);
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
    border-color: var(--red);
    color: var(--red);
}

/* ── Plugin rows ── */
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

.mod-loader-tag.paper {
    border-color: #f8717144;
    color: var(--red);
}

.mod-loader-tag.spigot {
    border-color: #fbbf2444;
    color: var(--yellow);
}

.mod-loader-tag.purpur {
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
    border-color: var(--red);
    color: var(--red);
    background: rgba(248, 113, 113, 0.04);
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
    border-color: var(--red);
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
    border-color: var(--red);
    color: var(--red);
    background: rgba(248, 113, 113, 0.07);
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
    border-color: var(--red);
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

.result-tag.paper {
    border-color: #f8717144;
    color: var(--red);
}

.result-tag.spigot {
    border-color: #fbbf2444;
    color: var(--yellow);
}

.result-tag.purpur {
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

.popular-section-label {
    font-family: 'VT323', monospace;
    font-size: 16px;
    letter-spacing: 3px;
    color: var(--red);
    padding: 4px 0 8px;
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
    border-color: var(--red);
    color: var(--red);
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