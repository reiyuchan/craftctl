<template>
    <div class="tab-content">

        <div class="worlds-header">
            <button class="btn btn-outline" @click="store.fetchWorlds()">Refresh</button>
        </div>

        <div v-if="store.worlds.length === 0" class="worlds-empty">
            <span>No worlds found in server directory</span>
        </div>

        <div class="worlds-grid">
            <div v-for="world in store.worlds" :key="world.name" class="world-card">

                <div class="world-thumb" :style="{ background: world.gradient }">
                    <span class="world-biome">{{ world.active ? 'ACTIVE' : 'World' }}</span>
                    <span v-if="world.active" class="world-active-badge">ACTIVE</span>
                </div>

                <div class="world-info">
                    <span class="world-name">{{ world.name }}</span>
                    <span class="world-meta">{{ world.size }}</span>
                </div>

                <div class="world-actions">
                    <button class="tbl-btn" :disabled="world.active || world.loading" @click="handleLoad(world)">
                        {{ world.loading ? 'Loading...' : 'Load' }}
                    </button>
                    <button class="tbl-btn" :disabled="world.backingUp" @click="handleBackup(world)">
                        {{ world.backingUp ? 'Backing up...' : 'Backup' }}
                    </button>
                    <button class="tbl-btn danger" :disabled="world.active" @click="handleDelete(world)">Delete</button>
                </div>

            </div>
        </div>

    </div>
</template>

<script>
import { store } from '../store.js'

export default {
    name: 'WorldsPage',
    emits: ['toast'],
    data() {
        return {
            store,
        }
    },
    async mounted() {
        await this.store.fetchWorlds()
    },
    methods: {
        async handleLoad(world) {
            try {
                await this.store.loadWorld(world.name)
                this.$emit('toast', { msg: `Loaded world: ${world.name}`, type: 'success' })
            } catch (e) {
                this.$emit('toast', { msg: `Failed to load world: ${e.message}`, type: 'danger' })
            }
        },
        async handleBackup(world) {
            try {
                this.$emit('toast', { msg: `Backup started for ${world.name}`, type: 'success' })
                await this.store.backupWorld(world.name)
                this.$emit('toast', { msg: `Backup complete: ${world.name}`, type: 'success' })
            } catch (e) {
                this.$emit('toast', { msg: `Backup failed: ${e.message}`, type: 'danger' })
            }
        },
        async handleDelete(world) {
            if (!confirm(`Delete world "${world.name}"? This cannot be undone.`)) return
            try {
                await this.store.deleteWorld(world.name)
                this.$emit('toast', { msg: `Deleted ${world.name}`, type: 'danger' })
            } catch (e) {
                this.$emit('toast', { msg: `Delete failed: ${e.message}`, type: 'danger' })
            }
        },
    },
}
</script>

<style scoped>
.worlds-header {
    display: flex;
    gap: 10px;
    margin-bottom: 20px;
}

.worlds-empty {
    text-align: center;
    padding: 40px 20px;
    color: var(--muted);
    font-size: 14px;
}

.worlds-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 16px;
}

.world-card {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
    transition: border-color 0.2s;
}

.world-card:hover {
    border-color: var(--green2);
}

.world-thumb {
    height: 100px;
    display: flex;
    align-items: flex-start;
    justify-content: flex-end;
    padding: 10px;
    position: relative;
}

.world-biome {
    position: absolute;
    bottom: 10px;
    left: 12px;
    font-size: 13px;
}

.world-active-badge {
    background: var(--green);
    color: #000;
    font-size: 10px;
    padding: 2px 7px;
    font-weight: 700;
    letter-spacing: 1px;
    border-radius: 2px;
}

.world-info {
    padding: 12px;
}

.world-name {
    display: block;
    font-size: 14px;
    font-weight: 600;
}

.world-meta {
    display: block;
    font-size: 11px;
    color: var(--muted);
    margin-top: 3px;
}

.world-actions {
    display: flex;
    gap: 6px;
    padding: 8px 12px 12px;
}
</style>
