<template>
    <div class="tab-content">

        <div class="settings-grid">
            <div class="card settings-card" v-for="section in settingsSections" :key="section.title">
                <div class="card-header">
                    <span class="card-title">{{ section.title }}</span>
                </div>
                <div class="settings-body">
                    <div v-for="setting in section.settings" :key="setting.key" class="setting-row">

                        <div class="setting-info">
                            <span class="setting-name">{{ setting.name }}</span>
                            <span class="setting-desc">{{ setting.desc }}</span>
                        </div>

                        <div class="setting-control">
                            <input v-if="setting.type === 'text'" :value="(store.serverProps as any)[setting.key]"
                                @input="(store.serverProps as any)[setting.key] = ($event.target as HTMLInputElement).value"
                                class="setting-input" />
                            <input v-else-if="setting.type === 'number'" :value="(store.serverProps as any)[setting.key]"
                                @input="(store.serverProps as any)[setting.key] = Number(($event.target as HTMLInputElement).value)"
                                type="number" class="setting-input sm" />
                            <label v-else-if="setting.type === 'toggle'" class="toggle">
                                <input type="checkbox" :checked="(store.serverProps as any)[setting.key]"
                                    @change="(store.serverProps as any)[setting.key] = ($event.target as HTMLInputElement).checked" />
                                <span class="toggle-track">
                                    <span class="toggle-thumb"></span>
                                </span>
                            </label>
                            <select v-else-if="setting.type === 'select'" :value="(store.serverProps as any)[setting.key]"
                                @change="(store.serverProps as any)[setting.key] = ($event.target as HTMLSelectElement).value"
                                class="setting-select">
                                <option v-for="opt in setting.options" :key="opt" :value="opt">{{ opt }}</option>
                            </select>
                        </div>

                    </div>
                </div>
            </div>
        </div>

        <div class="jvm-section card settings-card">
            <div class="card-header">
                <span class="card-title">JVM SETTINGS</span>
            </div>
            <div class="settings-body">
                <div class="setting-row">
                    <div class="setting-info">
                        <span class="setting-name">Min RAM</span>
                        <span class="setting-desc">Minimum heap size (e.g. 2G, 512M)</span>
                    </div>
                    <div class="setting-control">
                        <input v-model="jvmSettings.minRAM" class="setting-input sm" />
                    </div>
                </div>
                <div class="setting-row">
                    <div class="setting-info">
                        <span class="setting-name">Max RAM</span>
                        <span class="setting-desc">Maximum heap size (e.g. 4G, 8G)</span>
                    </div>
                    <div class="setting-control">
                        <input v-model="jvmSettings.maxRAM" class="setting-input sm" />
                    </div>
                </div>
                <div class="setting-row">
                    <div class="setting-info">
                        <span class="setting-name">Java Path</span>
                        <span class="setting-desc">Path to java binary (leave empty for default)</span>
                    </div>
                    <div class="setting-control">
                        <input v-model="jvmSettings.javaPath" class="setting-input" placeholder="java" />
                    </div>
                </div>
                <div class="setting-row">
                    <div class="setting-info">
                        <span class="setting-name">JVM Flags</span>
                        <span class="setting-desc">Additional JVM arguments</span>
                    </div>
                    <div class="setting-control">
                        <textarea v-model="jvmSettings.jvmFlags" class="setting-textarea" rows="3"></textarea>
                    </div>
                </div>
            </div>
        </div>

        <div class="settings-actions">
            <button class="btn btn-outline" @click="resetDefaults">RESET DEFAULTS</button>
            <button class="btn btn-primary" @click="saveJVM">💾 SAVE JVM</button>
            <button class="btn btn-primary" @click="saveSettings">💾 SAVE server.properties</button>
        </div>

    </div>
</template>

<script lang="ts">
import { api } from '../api.js'
import { store } from '../store.js'

const DEFAULT_PROPS = {
    serverName: 'My Minecraft Server',
    motd: 'A Minecraft Server',
    maxPlayers: 20,
    difficulty: 'normal',
    gamemode: 'survival',
    pvp: true,
    onlineMode: true,
    hardcore: false,
    whiteList: false,
    spawnAnimals: true,
    spawnMonsters: true,
    spawnNpcs: true,
    viewDistance: 10,
    simulationDistance: 10,
    port: 25565,
    levelType: 'minecraft:default',
}

export default {
    name: 'SettingsPage',
    emits: ['toast'],
    data() {
        return {
            store,
            jvmSettings: {
                minRAM: store.jvmSettings.minRAM,
                maxRAM: store.jvmSettings.maxRAM,
                jvmFlags: store.jvmSettings.jvmFlags,
                javaPath: store.jvmSettings.javaPath,
            },
            settingsSections: [
                {
                    title: 'GENERAL',
                    settings: [
                        { key: 'serverName', name: 'Server Name', desc: 'Display name for your server', type: 'text' },
                        { key: 'motd', name: 'MOTD', desc: 'Message shown in server list', type: 'text' },
                        { key: 'maxPlayers', name: 'Max Players', desc: 'Maximum concurrent players', type: 'number' },
                        { key: 'port', name: 'Port', desc: 'Server port (default 25565)', type: 'number' },
                    ],
                },
                {
                    title: 'GAMEPLAY',
                    settings: [
                        { key: 'difficulty', name: 'Difficulty', desc: 'Game difficulty', type: 'select', options: ['peaceful', 'easy', 'normal', 'hard'] },
                        { key: 'gamemode', name: 'Default Gamemode', desc: 'Default gamemode for new players', type: 'select', options: ['survival', 'creative', 'adventure', 'spectator'] },
                        { key: 'pvp', name: 'PvP', desc: 'Allow player vs player combat', type: 'toggle' },
                        { key: 'hardcore', name: 'Hardcore', desc: 'Hardcore mode (permanent death)', type: 'toggle' },
                    ],
                },
                {
                    title: 'WORLD',
                    settings: [
                        { key: 'levelType', name: 'Level Type', desc: 'World generation type', type: 'select', options: ['minecraft:default', 'minecraft:flat', 'minecraft:large_biomes', 'minecraft:amplified'] },
                        { key: 'viewDistance', name: 'View Distance', desc: 'Chunks loaded per player (2-32)', type: 'number' },
                        { key: 'simulationDistance', name: 'Sim. Distance', desc: 'Entity simulation distance', type: 'number' },
                        { key: 'spawnAnimals', name: 'Spawn Animals', desc: 'Allow passive mob spawning', type: 'toggle' },
                        { key: 'spawnMonsters', name: 'Spawn Monsters', desc: 'Allow hostile mob spawning', type: 'toggle' },
                    ],
                },
                {
                    title: 'SECURITY',
                    settings: [
                        { key: 'onlineMode', name: 'Online Mode', desc: 'Verify players against Mojang servers', type: 'toggle' },
                        { key: 'whiteList', name: 'Whitelist', desc: 'Only allow whitelisted players', type: 'toggle' },
                    ],
                },
            ],
        }
    },
    mounted() {
        store.fetchServerProps()
    },
    methods: {
        saveJVM() {
            store.jvmSettings.minRAM = this.jvmSettings.minRAM
            store.jvmSettings.maxRAM = this.jvmSettings.maxRAM
            store.jvmSettings.jvmFlags = this.jvmSettings.jvmFlags
            store.jvmSettings.javaPath = this.jvmSettings.javaPath
            this.$emit('toast', { msg: 'JVM settings saved!', type: 'success' })
        },
        async saveSettings() {
            try {
                const props = {
                    server_name: this.store.serverProps.serverName,
                    motd: this.store.serverProps.motd,
                    max_players: this.store.serverProps.maxPlayers,
                    difficulty: this.store.serverProps.difficulty,
                    gamemode: this.store.serverProps.gamemode,
                    pvp: this.store.serverProps.pvp,
                    online_mode: this.store.serverProps.onlineMode,
                    hardcore: this.store.serverProps.hardcore,
                    white_list: this.store.serverProps.whiteList,
                    spawn_animals: this.store.serverProps.spawnAnimals,
                    spawn_monsters: this.store.serverProps.spawnMonsters,
                    spawn_npcs: this.store.serverProps.spawnNpcs,
                    view_distance: this.store.serverProps.viewDistance,
                    simulation_distance: this.store.serverProps.simulationDistance,
                    port: this.store.serverProps.port,
                    level_type: this.store.serverProps.levelType,
                }
                await api.saveServerProps(props)
                this.$emit('toast', { msg: 'server.properties saved!', type: 'success' })
            } catch (e: any) {
                this.$emit('toast', { msg: `Save failed: ${e}`, type: 'danger' })
            }
        },
        resetDefaults() {
            Object.assign(this.store.serverProps, DEFAULT_PROPS)
            this.$emit('toast', { msg: 'Reset to defaults', type: 'warn' })
        },
    },
}
</script>

<style scoped>
.settings-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    margin-bottom: 20px;
}

.settings-body {
    padding: 4px 0;
}

.setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 16px;
    border-bottom: 1px solid rgba(30, 45, 61, 0.4);
}

.setting-row:last-child {
    border-bottom: none;
}

.setting-name {
    display: block;
    font-size: 13px;
}

.setting-desc {
    display: block;
    font-size: 11px;
    color: var(--muted);
    margin-top: 2px;
}

.setting-control {
    flex-shrink: 0;
    margin-left: 16px;
}

.setting-input {
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 5px 10px;
    color: var(--text);
    font-family: 'Share Tech Mono', monospace;
    font-size: 13px;
    width: 180px;
    outline: none;
}

.setting-input.sm {
    width: 80px;
}

.setting-input:focus {
    border-color: var(--green);
}

.setting-select {
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 5px 10px;
    color: var(--text);
    font-family: 'Share Tech Mono', monospace;
    font-size: 13px;
    outline: none;
    width: 180px;
    cursor: pointer;
}

.setting-textarea {
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 8px 10px;
    color: var(--text);
    font-family: 'Share Tech Mono', monospace;
    font-size: 12px;
    width: 320px;
    outline: none;
    resize: vertical;
    min-height: 60px;
    line-height: 1.5;
}

.setting-textarea:focus {
    border-color: var(--green);
}

.settings-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding-top: 8px;
}
</style>