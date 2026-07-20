<template>
  <div class="app">
    <div class="scanlines"></div>
    <div class="noise"></div>

    <!-- ── Sidebar ── -->
    <aside class="sidebar">
      <div class="logo">
        <span class="logo-icon">⛏</span>
        <div class="logo-text">
          <span class="logo-title">CRAFTCTL</span>
          <span class="logo-sub">Server Manager</span>
        </div>
      </div>

      <nav class="nav">
        <button v-for="item in navItems" :key="item.id" class="nav-item" :class="{ active: activeTab === item.id }"
          @click="activeTab = item.id">
          <span class="nav-icon">{{ item.icon }}</span>
          <span class="nav-label">{{ item.label }}</span>
          <span v-if="item.badge" class="nav-badge">{{ item.badge }}</span>
        </button>
      </nav>

      <div class="sidebar-footer">
        <div class="version-chip">v1.21.4</div>
        <div class="status-dot" :class="store.serverStatus"></div>
      </div>
    </aside>

    <!-- ── Main ── -->
    <main class="content">

      <!-- Topbar -->
      <header class="topbar">
        <div class="topbar-left">
          <h1 class="page-title">{{ currentPage.label }}</h1>
          <span class="page-subtitle">{{ currentPage.subtitle }}</span>
        </div>
        <div class="topbar-right">
          <div class="server-badge" :class="store.serverStatus">
            <span class="pulse-dot"></span>
            {{ serverStatusLabel }}
          </div>
          <button class="btn btn-danger" v-if="store.serverStatus === 'running'" @click="stopServer">
            ■ STOP
          </button>
          <button class="btn btn-primary" v-else @click="startServer" :disabled="store.serverStatus === 'starting'">
            <span v-if="store.serverStatus === 'starting'" class="spinner">◌</span>
            <span v-else>▶ START</span>
          </button>
        </div>
      </header>

      <!-- Dynamic page component -->
      <component :is="currentComponent" @toast="showToast" @navigate="activeTab = $event" />

    </main>

    <!-- ── Toast ── -->
    <transition name="toast">
      <div v-if="toast" class="toast" :class="toast.type">
        {{ toast.msg }}
      </div>
    </transition>
  </div>
</template>

<script>
import { store } from './store.js'
import DashboardPage from './pages/DashboardPage.vue'
import ConsolePage from './pages/ConsolePage.vue'
import PlayersPage from './pages/PlayersPage.vue'
import WorldsPage from './pages/WorldsPage.vue'
import ModsPage from './pages/ModsPage.vue'
import PluginsPage from './pages/PluginsPage.vue'
import JavaPage from './pages/JavaPage.vue'
import ServerVersionsPage from './pages/ServerVersionsPage.vue'
import SettingsPage from './pages/SettingsPage.vue'

export default {
  name: 'App',
  components: { DashboardPage, ConsolePage, PlayersPage, WorldsPage, ModsPage, PluginsPage, JavaPage, ServerVersionsPage, SettingsPage },

  data() {
    return {
      store,
      activeTab: 'dashboard',
      toast: null,
      _toastTimer: null,

      navItems: [
        { id: 'dashboard', icon: '▦', label: 'Dashboard', subtitle: 'Server overview & stats', component: DashboardPage },
        { id: 'console', icon: '⌨', label: 'Console', subtitle: 'Live server console', component: ConsolePage },
        { id: 'players', icon: '👤', label: 'Players', subtitle: 'Manage players', component: PlayersPage },
        { id: 'worlds', icon: '🌍', label: 'Worlds', subtitle: 'World management', component: WorldsPage },
        { id: 'mods', icon: '🧩', label: 'Mods', subtitle: 'Fabric / Forge / NeoForge', component: ModsPage },
        { id: 'plugins', icon: '🔌', label: 'Plugins', subtitle: 'Paper / Spigot / Purpur', component: PluginsPage },
        { id: 'versions', icon: '📦', label: 'Versions', subtitle: 'Server jar downloads', component: ServerVersionsPage },
        { id: 'java', icon: '☕', label: 'Java', subtitle: 'Java runtime manager', component: JavaPage },
        { id: 'settings', icon: '⚙', label: 'Settings', subtitle: 'Server & JVM config', component: SettingsPage },
      ],
    }
  },

  computed: {
    currentPage() {
      return this.navItems.find(n => n.id === this.activeTab) || {}
    },
    currentComponent() {
      return this.currentPage.component
    },
    serverStatusLabel() {
      return { running: 'Online', stopped: 'Offline', starting: 'Starting...' }[this.store.serverStatus]
    },
  },

  methods: {
    async startServer() {
      this.store.serverStatus = 'starting'
      this.store.addLog('INFO', 'info', 'Starting Minecraft server...')
      try {
        const { api, events } = await import('./api')
        const unlisten = await events.onServerStopped(() => {
          this.store.serverStatus = 'stopped'
          this.store.addLog('INFO', 'warn', 'Server process exited.')
          unlisten()
        })
        await api.startServer({
          javaPath: this.store.jvmSettings.javaPath || undefined,
          minRam: this.store.jvmSettings.minRAM,
          maxRam: this.store.jvmSettings.maxRAM,
          jvmFlags: this.store.jvmSettings.jvmFlags,
        })
        this.store.serverStatus = 'running'
        this.store.addLog('INFO', 'info', 'Server process started successfully.')
        this.showToast({ msg: 'Server started!', type: 'success' })
      } catch (e) {
        this.store.serverStatus = 'stopped'
        this.store.addLog('ERROR', 'error', `Failed to start: ${e}`)
        this.showToast({ msg: `Start failed: ${e}`, type: 'danger' })
      }
    },
    async stopServer() {
      try {
        const { api } = await import('./api')
        await api.stopServer()
        this.store.serverStatus = 'stopped'
        this.store.addLog('INFO', 'warn', 'Server stopping...')
        this.showToast({ msg: 'Server stopped.', type: 'warn' })
      } catch (e) {
        this.showToast({ msg: `Stop failed: ${e}`, type: 'danger' })
      }
    },
    showToast({ msg, type = 'success' }) {
      this.toast = { msg, type }
      clearTimeout(this._toastTimer)
      this._toastTimer = setTimeout(() => { this.toast = null }, 3000)
    },
  },

  mounted() {
    store.fetchServerInfo()
    store.fetchServerProps()
  },
}
</script>

<style>
/* Global ambient styles (non-scoped) */
@import url('https://fonts.googleapis.com/css2?family=VT323&family=Share+Tech+Mono&display=swap');
@import './assets/styles.css';

.app {
  display: flex;
  height: 100vh;
  position: relative;
}

/* Scanlines */
.scanlines {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 100;
  background: repeating-linear-gradient(0deg,
      rgba(0, 0, 0, 0) 0px, rgba(0, 0, 0, 0) 2px,
      rgba(0, 0, 0, 0.08) 2px, rgba(0, 0, 0, 0.08) 4px);
}

.noise {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 99;
  opacity: 0.025;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
}

/* ── Sidebar ── */
.sidebar {
  width: var(--sidebar-w);
  background: var(--bg2);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 20px 0;
  flex-shrink: 0;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 20px 24px;
  border-bottom: 1px solid var(--border);
}

.logo-icon {
  font-size: 28px;
}

.logo-title {
  display: block;
  font-family: 'VT323', monospace;
  font-size: 22px;
  color: var(--green);
  letter-spacing: 2px;
  line-height: 1;
  text-shadow: 0 0 10px rgba(74, 222, 128, 0.5);
}

.logo-sub {
  display: block;
  font-size: 10px;
  color: var(--muted);
  letter-spacing: 1px;
}

.nav {
  flex: 1;
  padding: 16px 0;
}

.nav-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 20px;
  background: none;
  border: none;
  color: var(--text2);
  cursor: pointer;
  font-family: 'Share Tech Mono', monospace;
  font-size: 13px;
  letter-spacing: 0.5px;
  transition: all 0.15s;
  position: relative;
}

.nav-item:hover {
  color: var(--text);
  background: var(--bg3);
}

.nav-item.active {
  color: var(--green);
  background: rgba(74, 222, 128, 0.07);
}

.nav-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: var(--green);
  box-shadow: 0 0 8px var(--green);
}

.nav-icon {
  font-size: 16px;
  width: 20px;
  text-align: center;
}

.nav-label {
  flex: 1;
}

.nav-badge {
  background: var(--red);
  color: #fff;
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 9px;
}

.sidebar-footer {
  padding: 16px 20px;
  border-top: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.version-chip {
  background: var(--bg3);
  border: 1px solid var(--border);
  padding: 3px 8px;
  font-size: 11px;
  color: var(--muted);
  border-radius: var(--radius);
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--muted);
  transition: all 0.3s;
}

.status-dot.running {
  background: var(--green);
  box-shadow: 0 0 8px var(--green);
}

.status-dot.starting {
  background: var(--yellow);
  box-shadow: 0 0 8px var(--yellow);
  animation: blink 1s infinite;
}

/* ── Main content ── */
.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--bg);
}

/* ── Topbar ── */
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 28px;
  border-bottom: 1px solid var(--border);
  background: var(--bg2);
  flex-shrink: 0;
}

.page-title {
  font-family: 'VT323', monospace;
  font-size: 26px;
  letter-spacing: 2px;
  color: var(--text);
  line-height: 1;
}

.page-subtitle {
  display: block;
  font-size: 11px;
  color: var(--muted);
  margin-top: 2px;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.server-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.server-badge.running {
  border-color: var(--green2);
  color: var(--green);
}

.server-badge.stopped {
  color: var(--muted);
}

.server-badge.starting {
  border-color: var(--yellow);
  color: var(--yellow);
}

.pulse-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
  animation: blink 1.2s infinite;
}

/* ── Toast ── */
.toast {
  position: fixed;
  bottom: 24px;
  right: 24px;
  background: var(--bg3);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 12px 20px;
  font-size: 13px;
  z-index: 200;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
}

.toast.success {
  border-color: var(--green);
  color: var(--green);
}

.toast.warn {
  border-color: var(--yellow);
  color: var(--yellow);
}

.toast.danger {
  border-color: var(--red);
  color: var(--red);
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.25s;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>