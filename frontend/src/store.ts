import { reactive } from 'vue'
import { api } from './api'
import type { BackupInfo, ScheduledTask } from './api'


export type ServerStatus = 'running' | 'stopped' | 'starting'

export type StatTrend = 'up' | 'down' | 'neutral'

export interface Stat {
  icon: string
  label: string
  value: string
  trend: StatTrend
  trendVal: string
}

export interface ChartData {
  TPS: number[]
  RAM: number[]
  CPU: number[]
}

export interface CurrentStats {
  cpu: number
  ram: number
  ramPercent: number
  threads: number
}

export type ChartMetric = keyof ChartData

export interface OnlinePlayer {
  name: string
  color: string
  time: string
  ping: number
}

export interface Player {
  name: string
  color: string
  online: boolean
  op: boolean
  lastSeen: string
  playtime: string
}

export interface JVMSettings {
  minRAM: string
  maxRAM: string
  jvmFlags: string
  javaPath: string
}

export interface WhitelistEntry {
  name: string
  uuid: string
}


export type LogLevel = 'INFO' | 'WARN' | 'ERROR'

export type LogType = 'info' | 'warn' | 'error' | 'join' | 'chat' | 'cmd'

export interface LogEntry {
  time: string
  level: LogLevel
  type: LogType
  msg: string
}

export type Difficulty = 'peaceful' | 'easy' | 'normal' | 'hard'

export type Gamemode = 'survival' | 'creative' | 'adventure' | 'spectator'

export type LevelType =
  | 'minecraft:default'
  | 'minecraft:flat'
  | 'minecraft:large_biomes'
  | 'minecraft:amplified'

export interface ServerProps {
  serverName: string
  motd: string
  maxPlayers: number
  difficulty: Difficulty
  gamemode: Gamemode
  pvp: boolean
  onlineMode: boolean
  hardcore: boolean
  whiteList: boolean
  spawnAnimals: boolean
  spawnMonsters: boolean
  spawnNpcs: boolean
  viewDistance: number
  simulationDistance: number
  port: number
  levelType: LevelType
}

export interface World {
  name: string
  biome: string
  size: string
  seed: string
  active: boolean
  gradient: string
  loading?: boolean
  backingUp?: boolean
}

export type ModLoaderType = 'Fabric' | 'Forge' | 'NeoForge'

export type PluginLoaderType = 'Paper' | 'Spigot' | 'Purpur'

export type ModLoader = ModLoaderType | PluginLoaderType | 'Vanilla'

export type InstalledModLoader = ModLoaderType | null

export type ItemCategory =
  | 'Performance'
  | 'World Generation'
  | 'Gameplay'
  | 'Utility'
  | 'Admin'
  | 'Economy'
  | 'Protection'
  | 'Chat'

export type ItemStatus = 'enabled' | 'disabled' | 'error' | 'update-available'

export type ModSource = 'Modrinth' | 'CurseForge' | 'Hangar' | 'Local'

export interface InstalledMod {
  id: string
  name: string
  version: string
  latestVersion: string
  author: string
  description: string
  category: ItemCategory
  loader: ModLoaderType
  fileSize: string
  status: ItemStatus
  source: ModSource
  icon: string
  fileName: string
}

export interface InstalledPlugin {
  id: string
  name: string
  version: string
  latestVersion: string
  author: string
  description: string
  category: ItemCategory
  loader: PluginLoaderType
  fileSize: string
  status: ItemStatus
  source: ModSource
  icon: string
  fileName: string
}

export interface ModSearchResult {
  id: string
  name: string
  author: string
  description: string
  category: ItemCategory
  downloads: string
  latestVersion: string
  loaders: ModLoaderType[]
  source: ModSource
  icon: string
  installed: boolean
}

export interface PluginSearchResult {
  id: string
  name: string
  author: string
  description: string
  category: ItemCategory
  downloads: string
  latestVersion: string
  loaders: PluginLoaderType[]
  source: ModSource
  icon: string
  installed: boolean
}

export type JavaVendor =
  | 'Adoptium'
  | 'Oracle'
  | 'Microsoft'
  | 'Amazon Corretto'
  | 'Azul Zulu'

export type JavaInstallStatus =
  | 'installed'
  | 'installing'
  | 'update-available'
  | 'error'

export type JavaArch = 'x64' | 'aarch64'

export interface JavaInstallation {
  id: string
  vendor: JavaVendor
  majorVersion: number
  fullVersion: string
  latestVersion: string
  arch: JavaArch
  installPath: string
  sizeOnDisk: string
  status: JavaInstallStatus
  isActive: boolean
  releaseType: 'LTS' | 'STS'
}

export interface JavaRelease {
  version: number
  lts: boolean
}

export type ServerSoftware =
  | 'Vanilla'
  | 'Paper'
  | 'Spigot'
  | 'Purpur'
  | 'Fabric'
  | 'Forge'
  | 'NeoForge'
  | 'Quilt'
  | 'Folia'
  | 'Magma'

export type ServerVersionStatus = 'installed' | 'downloading' | 'available'

export type ReleaseChannel = 'release' | 'snapshot' | 'beta' | 'alpha'

export interface ServerBuild {
  id: string
  software: ServerSoftware
  mcVersion: string
  build: string
  releaseDate: string
  channel: ReleaseChannel
  fileSize: string
  sha256: string
  changelog: string
  javaRequired: number
  status: ServerVersionStatus
  isActive: boolean
  downloadUrl: string
}

export interface ServerSoftwareMeta {
  id: ServerSoftware
  name: string
  icon: string
  description: string
  type: 'vanilla' | 'plugin' | 'mod' | 'hybrid'
  recommendedFor: string
  color: string
}

export interface Store {
  serverStatus: ServerStatus
  stats: Stat[]
  chartData: ChartData
  currentStats: CurrentStats
  onlinePlayers: OnlinePlayer[]
  maxPlayers: number
  allPlayers: Player[]
  whitelistPlayers: WhitelistEntry[]
  logs: LogEntry[]
  serverProps: ServerProps
  jvmSettings: JVMSettings
  worlds: World[]
  installedModLoader: InstalledModLoader
  installedMods: InstalledMod[]
  modSearchResults: ModSearchResult[]
  isSearchingMods: boolean
  installedPlugins: InstalledPlugin[]
  pluginSearchResults: PluginSearchResult[]
  isSearchingPlugins: boolean
  javaInstallations: JavaInstallation[]
  javaReleases: JavaRelease[]
  isInstallingJava: boolean
  serverBuilds: ServerBuild[]
  isDownloadingServer: boolean
  downloadingBuildId: string | null
  addLog(level: LogLevel, type: LogType, msg: string): void
  kickPlayer(name: string): void
  installMod(result: ModSearchResult): void
  uninstallMod(id: string): void
  toggleMod(id: string): void
  installPlugin(result: PluginSearchResult): void
  uninstallPlugin(id: string): void
  togglePlugin(id: string): void
  setActiveJava(id: string): void
  uninstallJava(id: string): void
  installJava(version: string): Promise<string>
  downloadServerBuild(software: string, mcVersion: string, build: string): Promise<void>
  deleteServerBuild(id: string): void
  setActiveServerBuild(id: string): void
  readonly hasModLoader: boolean
  readonly hasMods: boolean
  readonly hasPlugins: boolean
  readonly activeJava: JavaInstallation | undefined
  readonly activeServerBuild: ServerBuild | undefined
  // New fetch methods
  fetchJavaInstallations(): Promise<void>
  fetchJavaReleases(): Promise<void>
  fetchServerBuilds(): Promise<void>
  fetchInstalledMods(): Promise<void>
  fetchInstalledPlugins(): Promise<void>
  fetchServerProps(): Promise<void>
  fetchServerInfo(): Promise<void>
  fetchServerStats(): Promise<void>

  backups: BackupInfo[]
  fetchBackups(): Promise<void>
  createFullBackup(): Promise<void>
  restoreBackup(name: string): Promise<void>
  deleteBackup(name: string): Promise<void>

  scheduledTasks: ScheduledTask[]
  fetchScheduledTasks(): Promise<void>
  createScheduledTask(task: { name: string; type: string; interval: string }): Promise<void>
  updateScheduledTask(id: string, task: { name: string; type: string; interval: string; enabled: boolean }): Promise<void>
  deleteScheduledTask(id: string): Promise<void>
  runScheduledTask(id: string): Promise<void>
  toggleScheduledTask(id: string, enabled: boolean): Promise<void>

  fetchWorlds(): Promise<void>
  loadWorld(name: string): Promise<void>
  backupWorld(name: string): Promise<void>
  deleteWorld(name: string): Promise<void>
  fetchPlayers(): Promise<void>
  fetchOps(): Promise<void>
  opPlayer(name: string): Promise<void>
  deopPlayer(name: string): Promise<void>
  kickPlayerAction(name: string): Promise<void>
  banPlayerAction(name: string): Promise<void>
  pardonPlayer(name: string): Promise<void>
  fetchWhitelist(): Promise<void>
  addToWhitelist(name: string): Promise<void>
  removeFromWhitelist(name: string): Promise<void>
}


function hashCode(s: string): number {
  let h = 0
  for (let i = 0; i < s.length; i++) {
    h = ((h << 5) - h + s.charCodeAt(i)) | 0
  }
  return Math.abs(h)
}


const defaultJVMSettings = (): JVMSettings => ({
  minRAM: '2G',
  maxRAM: '4G',
  jvmFlags: '-XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200 -XX:+UnlockExperimentalVMOptions -XX:+DisableExplicitGC -XX:+AlwaysPreTouch -XX:G1NewSizePercent=30 -XX:G1MaxNewSizePercent=40',
  javaPath: '',
})

const defaultProps = (): ServerProps => ({
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
})

export const store = reactive<Store>({
  serverStatus: 'stopped',
  stats: [],
  chartData: { TPS: [], RAM: [], CPU: [] },
  currentStats: { cpu: 0, ram: 0, ramPercent: 0, threads: 0 },
  onlinePlayers: [],
  maxPlayers: 20,
  allPlayers: [],
  whitelistPlayers: [],
  logs: [],
  serverProps: defaultProps(),
  jvmSettings: defaultJVMSettings(),
  worlds: [],
  backups: [],
  scheduledTasks: [],
  installedModLoader: null,
  installedMods: [],
  modSearchResults: [],
  isSearchingMods: false,
  installedPlugins: [],
  pluginSearchResults: [],
  isSearchingPlugins: false,
  javaInstallations: [],
  javaReleases: [],
  isInstallingJava: false,
  serverBuilds: [],
  isDownloadingServer: false,
  downloadingBuildId: null,

  get hasModLoader(): boolean {
    return this.installedModLoader !== null
  },
  get hasMods(): boolean {
    return this.installedMods.length > 0
  },
  get hasPlugins(): boolean {
    return this.installedPlugins.length > 0
  },
  get activeJava(): JavaInstallation | undefined {
    return this.javaInstallations.find(j => j.isActive)
  },
  get activeServerBuild(): ServerBuild | undefined {
    return this.serverBuilds.find(b => b.isActive)
  },

  addLog(level: LogLevel, type: LogType, msg: string): void {
    const time = new Date().toTimeString().slice(0, 8)
    this.logs.push({ time, level, type, msg })
  },

  kickPlayer(name: string): void {
    this.onlinePlayers = this.onlinePlayers.filter(p => p.name !== name)
    const player = this.allPlayers.find(p => p.name === name)
    if (player) player.online = false
    this.addLog('INFO', 'warn', `${name} was kicked from the server.`)
  },

  installMod(result: ModSearchResult): void {
    const mod: InstalledMod = {
      id: result.id,
      name: result.name,
      version: result.latestVersion,
      latestVersion: result.latestVersion,
      author: result.author,
      description: result.description,
      category: result.category,
      loader: result.loaders[0] ?? 'Fabric',
      fileSize: 'Downloading...',
      status: 'enabled',
      source: result.source,
      icon: result.icon,
      fileName: `${result.id}-${result.latestVersion}.jar`,
    }
    this.installedMods.push(mod)
    result.installed = true
    this.addLog('INFO', 'info', `Installed mod ${result.name} v${result.latestVersion}`)
  },

  uninstallMod(id: string): void {
    const mod = this.installedMods.find(m => m.id === id)
    if (mod) this.addLog('INFO', 'warn', `Uninstalled mod ${mod.name}`)
    this.installedMods = this.installedMods.filter(m => m.id !== id)
    const result = this.modSearchResults.find(r => r.id === id)
    if (result) result.installed = false
  },

  toggleMod(id: string): void {
    const mod = this.installedMods.find(m => m.id === id)
    if (!mod || mod.status === 'error') return
    mod.status = mod.status === 'enabled' ? 'disabled' : 'enabled'
    this.addLog('INFO', 'info', `Mod ${mod.name} ${mod.status}`)
  },

  installPlugin(result: PluginSearchResult): void {
    const plugin: InstalledPlugin = {
      id: result.id,
      name: result.name,
      version: result.latestVersion,
      latestVersion: result.latestVersion,
      author: result.author,
      description: result.description,
      category: result.category,
      loader: result.loaders[0] ?? 'Paper',
      fileSize: 'Downloading...',
      status: 'enabled',
      source: result.source,
      icon: result.icon,
      fileName: `${result.id}-${result.latestVersion}.jar`,
    }
    this.installedPlugins.push(plugin)
    result.installed = true
    this.addLog('INFO', 'info', `Installed plugin ${result.name} v${result.latestVersion}`)
  },

  uninstallPlugin(id: string): void {
    const plugin = this.installedPlugins.find(p => p.id === id)
    if (plugin) this.addLog('INFO', 'warn', `Uninstalled plugin ${plugin.name}`)
    this.installedPlugins = this.installedPlugins.filter(p => p.id !== id)
    const result = this.pluginSearchResults.find(r => r.id === id)
    if (result) result.installed = false
  },

  togglePlugin(id: string): void {
    const plugin = this.installedPlugins.find(p => p.id === id)
    if (!plugin || plugin.status === 'error') return
    plugin.status = plugin.status === 'enabled' ? 'disabled' : 'enabled'
    this.addLog('INFO', 'info', `Plugin ${plugin.name} ${plugin.status}`)
  },

  setActiveJava(id: string): void {
    this.javaInstallations.forEach(j => { j.isActive = j.id === id })
    const java = this.javaInstallations.find(j => j.id === id)
    if (java)
      this.addLog('INFO', 'info', `Active Java set to ${java.vendor} ${java.majorVersion} (${java.fullVersion})`)
  },

  uninstallJava(id: string): void {
    const java = this.javaInstallations.find(j => j.id === id)
    if (!java) return
    if (id.startsWith('sys-')) {
      this.addLog('WARN', 'warn', `Cannot uninstall system Java (${java.vendor} ${java.majorVersion})`)
      return
    }
    this.javaInstallations = this.javaInstallations.filter(j => j.id !== id)
    this.addLog('INFO', 'warn', `Uninstalled ${java.vendor} Java ${java.majorVersion}`)
  },

  async installJava(version: string): Promise<string> {
    this.isInstallingJava = true
    try {
      const result = await api.downloadJava(version)
      await this.fetchJavaInstallations()
      this.addLog('INFO', 'info', `Java ${version} installed successfully`)
      return result.path
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Java install failed: ${e.message ?? e}`)
      throw e
    } finally {
      this.isInstallingJava = false
    }
  },

  async downloadServerBuild(software: string, mcVersion: string, build: string): Promise<void> {
    this.isDownloadingServer = true
    this.downloadingBuildId = `${software.toLowerCase()}-${mcVersion}-${build}`
    const existing = this.serverBuilds.find(
      b => b.software === software && b.mcVersion === mcVersion && b.build === build)
    if (existing) existing.status = 'downloading'
    try {
      await api.installServerSoftware(software, mcVersion, build)
      await this.fetchServerBuilds()
      this.addLog('INFO', 'info', `${software} ${mcVersion} (build ${build}) downloaded`)
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Download failed: ${e.message ?? e}`)
      throw e
    } finally {
      this.isDownloadingServer = false
      this.downloadingBuildId = null
    }
  },

  deleteServerBuild(id: string): void {
    const build = this.serverBuilds.find(b => b.id === id)
    if (!build) return
    this.serverBuilds = this.serverBuilds.filter(b => b.id !== id)
    this.addLog('INFO', 'warn', `Deleted ${build.software} ${build.mcVersion} build ${build.build}`)
  },

  setActiveServerBuild(id: string): void {
    this.serverBuilds.forEach(b => { b.isActive = b.id === id })
    const build = this.serverBuilds.find(b => b.id === id)
    if (build)
      this.addLog('INFO', 'info', `Active server set to ${build.software} ${build.mcVersion}`)
  },

  async fetchJavaInstallations(): Promise<void> {
    try {
      const runtimes = await api.detectJava()
      this.javaInstallations = runtimes.map((r, i) => ({
        id: r.id,
        vendor: r.vendor as JavaVendor,
        majorVersion: r.majorVersion,
        fullVersion: r.fullVersion,
        latestVersion: r.latestVersion,
        arch: r.arch as JavaArch,
        installPath: r.installPath,
        sizeOnDisk: r.sizeOnDisk,
        status: r.status as JavaInstallStatus,
        isActive: i === 0,
        releaseType: r.releaseType as 'LTS' | 'STS',
      }))
    } catch {
      this.javaInstallations = []
    }
  },

  async fetchJavaReleases(): Promise<void> {
    try {
      const releases = await api.javaVersions()
      this.javaReleases = releases
    } catch {
      this.javaReleases = []
    }
  },

  async fetchServerBuilds(): Promise<void> {
    try {
      const info = await api.getActiveInfo()
      const hasJar = info.has_server_jar
      if (hasJar) {
        this.serverBuilds = [{
          id: 'current',
          software: 'Paper' as ServerSoftware,
          mcVersion: '',
          build: '1',
          releaseDate: '',
          channel: 'release' as ReleaseChannel,
          fileSize: '',
          sha256: '',
          changelog: '',
          javaRequired: 17,
          status: 'installed' as ServerVersionStatus,
          isActive: true,
          downloadUrl: '',
        }]
      } else {
        this.serverBuilds = []
      }
    } catch {
      this.serverBuilds = []
    }
  },

  async fetchInstalledMods(): Promise<void> {
    try {
      const items = await api.getInstalledMods()
      this.installedMods = items.map(item => ({
        id: item.file_name,
        name: item.name ?? item.file_name,
        version: item.version,
        latestVersion: item.latestVersion || item.version,
        author: '',
        description: '',
        category: 'Utility' as ItemCategory,
        loader: 'Fabric' as ModLoaderType,
        fileSize: item.size,
        status: item.hasUpdate ? 'update-available' as ItemStatus : 'enabled' as ItemStatus,
        source: (item.source as ModSource) ?? 'Local',
        icon: '📦',
        fileName: item.file_name,
      }))
    } catch {
      this.installedMods = []
    }
  },

  async fetchInstalledPlugins(): Promise<void> {
    try {
      const items = await api.getInstalledPlugins()
      this.installedPlugins = items.map(item => ({
        id: item.file_name,
        name: item.name ?? item.file_name,
        version: item.version,
        latestVersion: item.latestVersion || item.version,
        author: '',
        description: '',
        category: 'Utility' as ItemCategory,
        loader: 'Paper' as PluginLoaderType,
        fileSize: item.size,
        status: item.hasUpdate ? 'update-available' as ItemStatus : 'enabled' as ItemStatus,
        source: (item.source as ModSource) ?? 'Local',
        icon: '📦',
        fileName: item.file_name,
      }))
    } catch {
      this.installedPlugins = []
    }
  },

  async fetchServerProps(): Promise<void> {
    try {
      const props = await api.readServerProps()
      this.serverProps = {
        serverName: props.server_name,
        motd: props.motd,
        maxPlayers: props.max_players,
        difficulty: props.difficulty as Difficulty,
        gamemode: props.gamemode as Gamemode,
        pvp: props.pvp,
        onlineMode: props.online_mode,
        hardcore: props.hardcore,
        whiteList: props.white_list,
        spawnAnimals: props.spawn_animals,
        spawnMonsters: props.spawn_monsters,
        spawnNpcs: props.spawn_npcs,
        viewDistance: props.view_distance,
        simulationDistance: props.simulation_distance,
        port: props.port,
        levelType: props.level_type as LevelType,
      }
    } catch {
      // keep defaults
    }
  },

  async fetchServerStats(): Promise<void> {
    try {
      const result = await api.getServerStats()
      this.currentStats = result.current
      this.chartData.CPU = result.history.map(h => h.cpu)
      this.chartData.RAM = result.history.map(h => h.ram / (1024 * 1024 * 1024))
      this.chartData.TPS = result.history.map(() => 20)
    } catch {
      // ignore
    }
  },

  async fetchServerInfo(): Promise<void> {
    try {
      const info = await api.getActiveInfo()
      if (info.has_server_jar) {
        await this.fetchServerBuilds()
      }
    } catch {
      // ignore
    }
  },

  async fetchWorlds(): Promise<void> {
    try {
      const items = await api.getWorlds()
      const gradients = [
        'linear-gradient(135deg, #1a472a 0%, #2d6a4f 100%)',
        'linear-gradient(135deg, #1e3a5f 0%, #2563eb 100%)',
        'linear-gradient(135deg, #7c2d12 0%, #dc2626 100%)',
        'linear-gradient(135deg, #581c87 0%, #9333ea 100%)',
        'linear-gradient(135deg, #065f46 0%, #10b981 100%)',
      ]
      this.worlds = items.map((w, i) => ({
        name: w.name,
        biome: w.active ? 'Active' : 'World',
        size: w.size,
        sizeBytes: w.sizeBytes,
        modifiedDate: w.modifiedDate,
        seed: '',
        active: w.active,
        gradient: gradients[i % gradients.length],
      }))
    } catch {
      this.worlds = []
    }
  },

  async loadWorld(name: string): Promise<void> {
    try {
      const world = this.worlds.find(w => w.name === name)
      if (world) world.loading = true
      await api.loadWorld(name)
      this.worlds.forEach(w => { w.active = w.name === name })
      this.addLog('INFO', 'info', `Loaded world: ${name}`)
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Failed to load world: ${e.message ?? e}`)
      throw e
    } finally {
      const world = this.worlds.find(w => w.name === name)
      if (world) world.loading = false
    }
  },

  async backupWorld(name: string): Promise<void> {
    try {
      const world = this.worlds.find(w => w.name === name)
      if (world) world.backingUp = true
      const result = await api.backupWorld(name)
      this.addLog('INFO', 'info', `Backup created: ${result.name}`)
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Backup failed: ${e.message ?? e}`)
      throw e
    } finally {
      const world = this.worlds.find(w => w.name === name)
      if (world) world.backingUp = false
    }
  },

  async deleteWorld(name: string): Promise<void> {
    try {
      await api.deleteWorld(name)
      this.worlds = this.worlds.filter(w => w.name !== name)
      this.addLog('INFO', 'warn', `Deleted world: ${name}`)
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Delete failed: ${e.message ?? e}`)
      throw e
    }
  },

  async fetchBackups(): Promise<void> {
    try {
      this.backups = await api.getBackups()
    } catch {
      this.backups = []
    }
  },

  async createFullBackup(): Promise<void> {
    try {
      const result = await api.createFullBackup()
      this.addLog('INFO', 'info', `Full backup created: ${result.name}`)
      await this.fetchBackups()
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Full backup failed: ${e.message ?? e}`)
      throw e
    }
  },

  async restoreBackup(name: string): Promise<void> {
    try {
      await api.restoreBackup(name)
      this.addLog('INFO', 'info', `Restored from backup: ${name}`)
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Restore failed: ${e.message ?? e}`)
      throw e
    }
  },

  async deleteBackup(name: string): Promise<void> {
    try {
      await api.deleteBackup(name)
      this.backups = this.backups.filter(b => b.name !== name)
      this.addLog('INFO', 'warn', `Deleted backup: ${name}`)
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Delete backup failed: ${e.message ?? e}`)
      throw e
    }
  },

  async fetchScheduledTasks(): Promise<void> {
    try {
      this.scheduledTasks = await api.getScheduledTasks()
    } catch {
      this.scheduledTasks = []
    }
  },

  async createScheduledTask(task: { name: string; type: string; interval: string }): Promise<void> {
    try {
      const created = await api.createScheduledTask(task)
      this.scheduledTasks.push(created)
      this.addLog('INFO', 'info', `Scheduled task created: ${task.name}`)
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Create task failed: ${e.message ?? e}`)
      throw e
    }
  },

  async updateScheduledTask(id: string, task: { name: string; type: string; interval: string; enabled: boolean }): Promise<void> {
    try {
      await api.updateScheduledTask(id, task)
      await this.fetchScheduledTasks()
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Update task failed: ${e.message ?? e}`)
      throw e
    }
  },

  async deleteScheduledTask(id: string): Promise<void> {
    try {
      await api.deleteScheduledTask(id)
      this.scheduledTasks = this.scheduledTasks.filter(t => t.id !== id)
      this.addLog('INFO', 'warn', 'Scheduled task deleted')
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Delete task failed: ${e.message ?? e}`)
      throw e
    }
  },

  async runScheduledTask(id: string): Promise<void> {
    try {
      await api.runScheduledTask(id)
      await this.fetchScheduledTasks()
      this.addLog('INFO', 'info', 'Task executed manually')
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Run task failed: ${e.message ?? e}`)
      throw e
    }
  },

  async toggleScheduledTask(id: string, enabled: boolean): Promise<void> {
    const task = this.scheduledTasks.find(t => t.id === id)
    if (!task) return
    try {
      await api.updateScheduledTask(id, { name: task.name, type: task.type, interval: task.interval, enabled })
      task.enabled = enabled
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Toggle task failed: ${e.message ?? e}`)
      throw e
    }
  },

  async fetchPlayers(): Promise<void> {
    try {
      const result = await api.getPlayers()
      const ops = await api.getOps().catch(() => [])
      const opNames = new Set(ops.map(o => o.name))
      this.allPlayers = result.players.map(name => ({
        name,
        color: `hsl(${hashCode(name) % 360}, 50%, 45%)`,
        online: true,
        op: opNames.has(name),
        lastSeen: '',
        playtime: '',
      }))
      this.maxPlayers = result.total
    } catch {
      this.allPlayers = []
    }
  },

  async fetchOps(): Promise<void> {
    try {
      const ops = await api.getOps()
      const opNames = new Set(ops.map(o => o.name))
      for (const player of this.allPlayers) {
        player.op = opNames.has(player.name)
      }
    } catch {
      // ignore
    }
  },

  async opPlayer(name: string): Promise<void> {
    try {
      await api.opPlayer(name)
      this.addLog('INFO', 'info', `Opped ${name}`)
      await this.fetchOps()
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Op failed: ${e.message ?? e}`)
      throw e
    }
  },

  async deopPlayer(name: string): Promise<void> {
    try {
      await api.deopPlayer(name)
      this.addLog('INFO', 'info', `Deopped ${name}`)
      await this.fetchOps()
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Deop failed: ${e.message ?? e}`)
      throw e
    }
  },

  async kickPlayerAction(name: string): Promise<void> {
    try {
      await api.kickPlayer(name)
      this.kickPlayer(name)
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Kick failed: ${e.message ?? e}`)
      throw e
    }
  },

  async banPlayerAction(name: string): Promise<void> {
    try {
      await api.banPlayer(name)
      this.addLog('INFO', 'warn', `Banned ${name}`)
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Ban failed: ${e.message ?? e}`)
      throw e
    }
  },

  async pardonPlayer(name: string): Promise<void> {
    try {
      await api.pardonPlayer(name)
      this.addLog('INFO', 'info', `Pardoned ${name}`)
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Pardon failed: ${e.message ?? e}`)
      throw e
    }
  },

  async fetchWhitelist(): Promise<void> {
    try {
      const list = await api.getWhitelist()
      this.whitelistPlayers = list.map(e => ({
        name: e.name,
        uuid: e.uuid ?? '',
      }))
    } catch {
      this.whitelistPlayers = []
    }
  },

  async addToWhitelist(name: string): Promise<void> {
    try {
      await api.whitelistAdd(name)
      this.addLog('INFO', 'info', `Added ${name} to whitelist`)
      await this.fetchWhitelist()
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Whitelist add failed: ${e.message ?? e}`)
      throw e
    }
  },

  async removeFromWhitelist(name: string): Promise<void> {
    try {
      await api.whitelistRemove(name)
      this.addLog('INFO', 'warn', `Removed ${name} from whitelist`)
      await this.fetchWhitelist()
    } catch (e: any) {
      this.addLog('ERROR', 'error', `Whitelist remove failed: ${e.message ?? e}`)
      throw e
    }
  },
})