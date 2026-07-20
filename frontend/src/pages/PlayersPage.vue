<template>
    <div class="tab-content">

        <div class="players-toolbar">
            <input v-model="playerSearch" class="search-input" placeholder="🔍 Search players..." />
            <div class="toolbar-right">
                <button class="btn btn-sm btn-outline" @click="showBanModal = true">+ BAN</button>
                <button class="btn btn-sm btn-primary" @click="showWhitelistModal = true">WHITELIST</button>
            </div>
        </div>

        <div class="tabs-inner">
            <button v-for="t in playerTabs" :key="t" :class="['inner-tab', { active: activePlayerTab === t }]"
                @click="activePlayerTab = t">{{ t }}</button>
        </div>

        <div class="card">
            <table class="players-table">
                <thead>
                    <tr>
                        <th>PLAYER</th>
                        <th>STATUS</th>
                        <th>ACTIONS</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="p in filteredPlayers" :key="p.name">
                        <td>
                            <div class="td-player">
                                <div class="player-avatar sm" :style="{ background: p.color }">{{ p.name[0] }}</div>
                                <span>{{ p.name }}</span>
                                <span v-if="p.op" class="op-badge">OP</span>
                            </div>
                        </td>
                        <td>
                            <span class="status-pill" :class="p.online ? 'online' : 'offline'">
                                {{ p.online ? 'Online' : 'Offline' }}
                            </span>
                        </td>
                        <td>
                            <div class="action-row">
                                <button class="tbl-btn" @click="opToggle(p)">{{ p.op ? 'Deop' : 'Op' }}</button>
                                <button v-if="p.online" class="tbl-btn warn" @click="kickPlayer(p.name)">Kick</button>
                                <button class="tbl-btn danger" @click="banPlayer(p.name)">Ban</button>
                            </div>
                        </td>
                    </tr>
                    <tr v-if="filteredPlayers.length === 0">
                        <td colspan="3" class="empty-row">No players found</td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-if="showBanModal" class="modal-overlay" @click.self="showBanModal = false">
            <div class="modal-card">
                <h3>Ban Player</h3>
                <input v-model="banName" class="search-input" placeholder="Player name" @keyup.enter="confirmBan" />
                <div class="modal-actions">
                    <button class="btn btn-sm btn-outline" @click="showBanModal = false">Cancel</button>
                    <button class="btn btn-sm btn-danger" @click="confirmBan">Ban</button>
                </div>
            </div>
        </div>

        <div v-if="showWhitelistModal" class="modal-overlay" @click.self="showWhitelistModal = false">
            <div class="modal-card">
                <h3>Whitelist</h3>
                <div class="whitelist-list">
                    <div v-for="w in store.whitelistPlayers" :key="w.name" class="whitelist-item">
                        <span>{{ w.name }}</span>
                        <button class="tbl-btn danger" @click="removeFromWhitelist(w.name)">Remove</button>
                    </div>
                    <p v-if="store.whitelistPlayers.length === 0" class="empty-row">No whitelisted players</p>
                </div>
                <div class="whitelist-add">
                    <input v-model="whitelistName" class="search-input" placeholder="Add player..." @keyup.enter="addWhitelist" />
                    <button class="btn btn-sm btn-primary" @click="addWhitelist">Add</button>
                </div>
                <div class="modal-actions">
                    <button class="btn btn-sm btn-outline" @click="showWhitelistModal = false">Close</button>
                </div>
            </div>
        </div>

    </div>
</template>

<script>
import { store } from '../store.js'

export default {
    name: 'PlayersPage',
    emits: ['toast'],
    data() {
        return {
            store,
            playerSearch: '',
            activePlayerTab: 'All Players',
            playerTabs: ['All Players', 'Online', 'Banned', 'Whitelist'],
            showBanModal: false,
            banName: '',
            showWhitelistModal: false,
            whitelistName: '',
            refreshTimer: null,
        }
    },
    computed: {
        filteredPlayers() {
            let list = this.store.allPlayers
            if (this.activePlayerTab === 'Online') list = list.filter(p => p.online)
            if (this.activePlayerTab === 'Whitelist') {
                return this.store.whitelistPlayers
                    .filter(w => !this.playerSearch || w.name.toLowerCase().includes(this.playerSearch.toLowerCase()))
                    .map(w => ({
                        name: w.name,
                        color: `hsl(${this.hashCode(w.name) % 360}, 50%, 45%)`,
                        online: this.store.allPlayers.some(p => p.name === w.name && p.online),
                        op: false,
                        lastSeen: '',
                        playtime: '',
                    }))
            }
            if (this.playerSearch) list = list.filter(p =>
                p.name.toLowerCase().includes(this.playerSearch.toLowerCase())
            )
            return list
        },
    },
    methods: {
        hashCode(s) {
            let h = 0
            for (let i = 0; i < s.length; i++) {
                h = ((h << 5) - h + s.charCodeAt(i)) | 0
            }
            return Math.abs(h)
        },
        async kickPlayer(name) {
            try {
                await this.store.kickPlayerAction(name)
                this.$emit('toast', { msg: `Kicked ${name}`, type: 'warn' })
            } catch {
                this.$emit('toast', { msg: `Failed to kick ${name}`, type: 'danger' })
            }
        },
        async banPlayer(name) {
            try {
                await this.store.banPlayerAction(name)
                this.$emit('toast', { msg: `Banned ${name}`, type: 'danger' })
            } catch {
                this.$emit('toast', { msg: `Failed to ban ${name}`, type: 'danger' })
            }
        },
        async confirmBan() {
            if (this.banName.trim()) {
                await this.banPlayer(this.banName.trim())
                this.banName = ''
                this.showBanModal = false
            }
        },
        async opToggle(player) {
            try {
                if (player.op) {
                    await this.store.deopPlayer(player.name)
                } else {
                    await this.store.opPlayer(player.name)
                }
                this.$emit('toast', {
                    msg: `${player.op ? 'Deopped' : 'Opped'} ${player.name}`,
                    type: 'success',
                })
            } catch {
                this.$emit('toast', { msg: `Failed to update ops for ${player.name}`, type: 'danger' })
            }
        },
        async addWhitelist() {
            if (this.whitelistName.trim()) {
                try {
                    await this.store.addToWhitelist(this.whitelistName.trim())
                    this.$emit('toast', { msg: `Added ${this.whitelistName.trim()} to whitelist`, type: 'success' })
                    this.whitelistName = ''
                } catch {
                    this.$emit('toast', { msg: 'Failed to add to whitelist', type: 'danger' })
                }
            }
        },
        async removeFromWhitelist(name) {
            try {
                await this.store.removeFromWhitelist(name)
                this.$emit('toast', { msg: `Removed ${name} from whitelist`, type: 'warn' })
            } catch {
                this.$emit('toast', { msg: 'Failed to remove from whitelist', type: 'danger' })
            }
        },
        async refreshPlayers() {
            try {
                await this.store.fetchPlayers()
            } catch {
                // ignore
            }
        },
    },
    mounted() {
        this.refreshPlayers()
        this.refreshTimer = setInterval(() => this.refreshPlayers(), 10000)
    },
    beforeUnmount() {
        if (this.refreshTimer) clearInterval(this.refreshTimer)
    },
}
</script>

<style scoped>
.players-toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 14px;
}

.toolbar-right {
    display: flex;
    gap: 8px;
}

.players-table {
    width: 100%;
    border-collapse: collapse;
}

.players-table th {
    padding: 10px 16px;
    text-align: left;
    font-size: 11px;
    letter-spacing: 1.5px;
    color: var(--muted);
    border-bottom: 1px solid var(--border);
}

.players-table td {
    padding: 10px 16px;
    border-bottom: 1px solid rgba(30, 45, 61, 0.5);
}

.players-table tr:hover td {
    background: var(--bg3);
}

.td-player {
    display: flex;
    align-items: center;
    gap: 8px;
}

.td-muted {
    color: var(--text2);
    font-size: 12px;
}

.action-row {
    display: flex;
    gap: 6px;
}

.empty-row {
    text-align: center;
    color: var(--muted);
    padding: 24px !important;
}

.modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
}

.modal-card {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 20px;
    min-width: 320px;
    max-width: 420px;
}

.modal-card h3 {
    margin: 0 0 12px;
    font-size: 14px;
}

.modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 14px;
}

.whitelist-list {
    max-height: 200px;
    overflow-y: auto;
    margin-bottom: 10px;
}

.whitelist-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 6px 0;
    border-bottom: 1px solid rgba(30, 45, 61, 0.5);
}

.whitelist-add {
    display: flex;
    gap: 8px;
}

.whitelist-add .search-input {
    flex: 1;
}
</style>