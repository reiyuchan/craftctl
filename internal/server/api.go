package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	*Server
}

func newHandler(s *Server) Handler {
	return Handler{Server: s}
}

func (h Handler) routes(app *fiber.App) {
	g := app.Group("/api")

	g.Get("/server/dir", h.serverDir)
	g.Post("/server/dir/ensure", h.serverDirEnsure)
	g.Get("/server/info", h.serverInfo)
	g.Get("/server/props", h.readProps)
	g.Post("/server/props", h.saveProps)
	g.Get("/server/eula", h.checkEula)
	g.Post("/server/eula/accept", h.acceptEula)
	g.Post("/server/jar/download", h.downloadJar)
	g.Post("/server/start", h.startServer)
	g.Post("/server/stop", h.stopServer)
	g.Post("/server/command", h.sendCommand)
	g.Get("/server/logs", h.getServerLogs)
	g.Get("/server/logs/download", h.downloadServerLogs)
	g.Get("/events/server-stats", h.events.Handler("server-stats"))
	g.Get("/server/stats", h.getServerStats)

	g.Get("/java/detect", h.detectJava)
	g.Get("/java/versions", h.javaVersions)
	g.Post("/java/download", h.downloadJava)

	g.Post("/folder/open", h.openFolder)

	g.Get("/mods/popular", h.popularMods)
	g.Post("/mods/search", h.searchMods)
	g.Get("/mods/versions/:id", h.modVersions)
	g.Post("/mods/download", h.downloadMod)
	g.Get("/mods/installed", h.installedMods)
	g.Post("/mods/delete", h.deleteMod)
	g.Get("/mods/updates", h.checkModUpdates)
	g.Post("/mods/update", h.updateMod)

	g.Get("/plugins/popular", h.popularPlugins)
	g.Post("/plugins/search", h.searchPlugins)
	g.Post("/plugins/download", h.downloadPlugin)
	g.Get("/plugins/installed", h.installedPlugins)
	g.Post("/plugins/delete", h.deletePlugin)
	g.Get("/plugins/updates", h.checkPluginUpdates)
	g.Post("/plugins/update", h.updatePlugin)

	g.Get("/versions/paper/:mcVersion/builds", h.paperBuilds)
	g.Get("/versions/paper/:mcVersion/build/:build/url", h.paperDownloadURL)
	g.Get("/versions/vanilla", h.vanillaVersions)
	g.Post("/versions/vanilla/url", h.vanillaDownloadURL)
	g.Post("/versions/fabric/install", h.installFabric)

	g.Get("/versions/purpur/:mcVersion", h.purpurVersions)
	g.Get("/versions/folia/:mcVersion", h.foliaVersions)
	g.Get("/versions/folia/:mcVersion/build/:build/url", h.foliaDownloadURL)
	g.Get("/versions/neoforge/:mcVersion", h.neoforgeVersions)
	g.Get("/versions/forge/:mcVersion", h.forgeVersions)
	g.Get("/versions/quilt/:mcVersion", h.quiltVersions)
	g.Get("/versions/magma/:mcVersion", h.magmaVersions)
	g.Get("/versions/spigot", h.spigotInfo)

	g.Post("/server/install", h.installServer)

	g.Get("/worlds", h.listWorlds)
	g.Post("/worlds/load", h.loadWorld)
	g.Post("/worlds/backup", h.backupWorld)
	g.Delete("/worlds/:name", h.deleteWorld)

	g.Get("/backups", h.listBackups)
	g.Post("/backups/full", h.createFullBackup)
	g.Post("/backups/restore", h.restoreBackup)
	g.Delete("/backups/:name", h.deleteBackup)

	h.registerPlayerRoutes(g)

	g.Get("/scheduler/tasks", h.listScheduledTasks)
	g.Post("/scheduler/tasks", h.createScheduledTask)
	g.Put("/scheduler/tasks/:id", h.updateScheduledTask)
	g.Delete("/scheduler/tasks/:id", h.deleteScheduledTask)
	g.Post("/scheduler/tasks/:id/run", h.runScheduledTask)

	g.Get("/events/server-log", h.events.Handler("server-log"))
	g.Get("/events/server-stopped", h.events.Handler("server-stopped"))
	g.Get("/events/server-error", h.events.Handler("server-error"))
}

// ── Server lifecycle ──────────────────────────────────────────────────────────

func (h Handler) serverDir(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"server_dir": h.cfg.ServerDir})
}

func (h Handler) serverDirEnsure(c *fiber.Ctx) error {
	os.MkdirAll(h.cfg.ServerDir, 0o755)
	return c.JSON(fiber.Map{"server_dir": h.cfg.ServerDir})
}

func (h Handler) serverInfo(c *fiber.Ctx) error {
	dir := h.cfg.ServerDir
	return c.JSON(fiber.Map{
		"server_dir":     dir,
		"has_server_jar": exists(dir, "server.jar"),
		"has_eula":       exists(dir, "eula.txt"),
		"has_properties": exists(dir, "server.properties"),
	})
}

func (h Handler) readProps(c *fiber.Ctx) error {
	props, err := readServerProperties(h.cfg.ServerDir)
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(props)
}

func (h Handler) saveProps(c *fiber.Ctx) error {
	var props serverProperties
	if err := c.BodyParser(&props); err != nil {
		return errorResp(c, 400, err)
	}
	if err := writeServerProperties(h.cfg.ServerDir, props); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h Handler) checkEula(c *fiber.Ctx) error {
	data, err := os.ReadFile(filePath(h.cfg.ServerDir, "eula.txt"))
	if err != nil {
		return c.JSON(false)
	}
	return c.JSON(contains(string(data), "eula=true"))
}

func (h Handler) acceptEula(c *fiber.Ctx) error {
	os.MkdirAll(h.cfg.ServerDir, 0o755)
	path := filePath(h.cfg.ServerDir, "eula.txt")
	if err := os.WriteFile(path, []byte("eula=true\n"), 0o644); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"path": path})
}

func (h Handler) downloadJar(c *fiber.Ctx) error {
	var body struct{ URL string }
	if err := c.BodyParser(&body); err != nil {
		return errorResp(c, 400, err)
	}
	path, err := downloadFile(downloadClient, body.URL, filePath(h.cfg.ServerDir, "server.jar"))
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"path": path})
}

func (h Handler) startServer(c *fiber.Ctx) error {
	var opts startOpts
	c.BodyParser(&opts)

	java := opts.JavaPath
	if java == "" {
		java = "java"
	}
	minRam := opts.MinRAM
	if minRam == "" {
		minRam = h.cfg.MinRAM
	}
	maxRam := opts.MaxRAM
	if maxRam == "" {
		maxRam = h.cfg.MaxRAM
	}
	flags := opts.JVMFlags
	if flags == "" {
		flags = h.cfg.JVMFlags
	}

	args := []string{
		"-Xms" + minRam,
		"-Xmx" + maxRam,
	}
	args = append(args, fields(flags)...)
	args = append(args, "-jar", "server.jar", "nogui")

	eulaPath := filePath(h.cfg.ServerDir, "eula.txt")
	if !existsFile(eulaPath) {
		os.WriteFile(eulaPath, []byte("eula=true\n"), 0o644)
	}

	if !existsFile(filePath(h.cfg.ServerDir, "server.jar")) {
		return errorResp(c, 400, errNoServerJar)
	}

	if err := h.ws.Start(java, h.cfg.ServerDir, args...); err != nil {
		return errorResp(c, 500, err)
	}
	pid := h.mc.PID()
	if pid > 0 {
		h.stats.Start(pid)
	}
	return c.JSON(fiber.Map{"status": "starting"})
}

func (h Handler) stopServer(c *fiber.Ctx) error {
	h.stats.Stop()
	if err := h.ws.Stop(); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "stopping"})
}

func (h Handler) sendCommand(c *fiber.Ctx) error {
	var body struct{ Command string }
	c.BodyParser(&body)
	if err := h.ws.Send(body.Command); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "sent"})
}

func (h Handler) getServerLogs(c *fiber.Ctx) error {
	return c.JSON(h.mc.GetLogs())
}

func (h Handler) downloadServerLogs(c *fiber.Ctx) error {
	logs := h.mc.GetLogs()
	data := strings.Join(logs, "")
	c.Set("Content-Type", "text/plain; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="server.log"`)
	return c.SendString(data)
}

func (h Handler) getServerStats(c *fiber.Ctx) error {
	current := h.stats.Snapshot()
	history := h.stats.History()
	return c.JSON(fiber.Map{
		"current": fiber.Map{
			"cpu":        current.CPU,
			"ram":        current.RAM,
			"ramPercent": current.RAMPercent,
			"threads":    current.Threads,
		},
		"history": history,
	})
}

// ── Java ─────────────────────────────────────────────────────────────────────

func (h Handler) detectJava(c *fiber.Ctx) error {
	runtimes := detectJavaRuntimes(filepath.Join(h.cfg.DataDir, "java"))
	if runtimes == nil {
		runtimes = []javaInstallation{}
	}
	return c.JSON(runtimes)
}

func (h Handler) javaVersions(c *fiber.Ctx) error {
	versions, err := adoptiumReleases()
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(versions)
}

func (h Handler) downloadJava(c *fiber.Ctx) error {
	var body struct {
		Version string `json:"version"`
	}
	if err := c.BodyParser(&body); err != nil {
		return errorResp(c, 400, err)
	}

	javaDir := filepath.Join(h.cfg.DataDir, "java")
	path, err := adoptiumDownload(body.Version, javaDir)
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"path": path})
}

// ── Folders ─────────────────────────────────────────────────────────────────

func (h Handler) openFolder(c *fiber.Ctx) error {
	var body struct{ Path string }
	c.BodyParser(&body)

	dir := body.Path
	if dir == "" {
		dir = h.cfg.ServerDir
	}
	os.MkdirAll(dir, 0o755)
	openFolder(dir)
	return c.JSON(fiber.Map{"status": "ok"})
}

// ── Mods ────────────────────────────────────────────────────────────────────

func (h Handler) searchMods(c *fiber.Ctx) error {
	var body modSearchReq
	c.BodyParser(&body)

	items, err := modrinthSearch(body.Query, body.Loaders, body.GameVersion)
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(items)
}

func (h Handler) modVersions(c *fiber.Ctx) error {
	id := c.Params("id")
	versions, err := modrinthVersions(id)
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(versions)
}

func (h Handler) downloadMod(c *fiber.Ctx) error {
	var body struct {
		ProjectID string `json:"projectId"`
		VersionID string `json:"versionId"`
	}
	c.BodyParser(&body)

	path, err := modrinthDownload(body.ProjectID, body.VersionID, filePath(h.cfg.ServerDir, "mods"))
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"path": path})
}

func (h Handler) installedMods(c *fiber.Ctx) error {
	items, err := installedJars(h.cfg.ServerDir, "mods")
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(items)
}

func (h Handler) deleteMod(c *fiber.Ctx) error {
	var body struct{ FileName string }
	c.BodyParser(&body)
	os.Remove(filePath(h.cfg.ServerDir, "mods", body.FileName))
	return c.JSON(fiber.Map{"status": "deleted"})
}

// ── Plugins ────────────────────────────────────────────────────────────────

func (h Handler) searchPlugins(c *fiber.Ctx) error {
	var body struct{ Query string }
	c.BodyParser(&body)

	items, err := pluginSearch(body.Query)
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(items)
}

func (h Handler) downloadPlugin(c *fiber.Ctx) error {
	var body pluginDownloadReq
	c.BodyParser(&body)

	path, err := pluginDownload(body.Slug, body.Version, body.Source,
		filePath(h.cfg.ServerDir, "plugins"))
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"path": path})
}

func (h Handler) installedPlugins(c *fiber.Ctx) error {
	items, err := installedJars(h.cfg.ServerDir, "plugins")
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(items)
}

func (h Handler) deletePlugin(c *fiber.Ctx) error {
	var body struct{ FileName string }
	c.BodyParser(&body)
	os.Remove(filePath(h.cfg.ServerDir, "plugins", body.FileName))
	return c.JSON(fiber.Map{"status": "deleted"})
}

func (h Handler) checkModUpdates(c *fiber.Ctx) error {
	items, err := installedJars(h.cfg.ServerDir, "mods")
	if err != nil {
		return errorResp(c, 500, err)
	}
	result, err := checkModrinthUpdates(items)
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(result)
}

func (h Handler) updateMod(c *fiber.Ctx) error {
	var body struct {
		ProjectID string `json:"projectId"`
		FileName  string `json:"fileName"`
	}
	if err := c.BodyParser(&body); err != nil {
		return errorResp(c, 400, err)
	}
	path, err := modrinthDownloadLatest(body.ProjectID, filePath(h.cfg.ServerDir, "mods"))
	if err != nil {
		return errorResp(c, 500, err)
	}
	// Remove old file
	os.Remove(filePath(h.cfg.ServerDir, "mods", body.FileName))
	return c.JSON(fiber.Map{"path": path})
}

func (h Handler) checkPluginUpdates(c *fiber.Ctx) error {
	items, err := installedJars(h.cfg.ServerDir, "plugins")
	if err != nil {
		return errorResp(c, 500, err)
	}
	result, err := checkPluginUpdatesRemote(items)
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(result)
}

func (h Handler) updatePlugin(c *fiber.Ctx) error {
	var body struct {
		Slug     string `json:"slug"`
		FileName string `json:"fileName"`
		Source   string `json:"source"`
	}
	if err := c.BodyParser(&body); err != nil {
		return errorResp(c, 400, err)
	}
	path, err := pluginDownloadLatest(body.Slug, body.Source, filePath(h.cfg.ServerDir, "plugins"))
	if err != nil {
		return errorResp(c, 500, err)
	}
	os.Remove(filePath(h.cfg.ServerDir, "plugins", body.FileName))
	return c.JSON(fiber.Map{"path": path})
}

func (h Handler) popularMods(c *fiber.Ctx) error {
	items, err := popularMods()
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(items)
}

func (h Handler) popularPlugins(c *fiber.Ctx) error {
	items, err := popularPlugins()
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(items)
}

// ── Versions ───────────────────────────────────────────────────────────────

func (h Handler) paperBuilds(c *fiber.Ctx) error {
	mcVersion := c.Params("mcVersion")
	builds, err := paperBuilds(mcVersion)
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(builds)
}

func (h Handler) paperDownloadURL(c *fiber.Ctx) error {
	mcVersion := c.Params("mcVersion")
	build := c.Params("build")
	url, err := paperDownloadURL(mcVersion, build)
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(url)
}

func (h Handler) vanillaVersions(c *fiber.Ctx) error {
	versions, err := vanillaVersions()
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(versions)
}

func (h Handler) vanillaDownloadURL(c *fiber.Ctx) error {
	var body struct{ VersionURL string }
	c.BodyParser(&body)
	url, err := vanillaDownloadURL(body.VersionURL)
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(url)
}

func (h Handler) installFabric(c *fiber.Ctx) error {
	var body struct{ MCVersion string }
	c.BodyParser(&body)
	path, err := fabricInstall(body.MCVersion, filePath(h.cfg.ServerDir, "mods"))
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"path": path})
}

// ── Scheduler ─────────────────────────────────────────────────────────────

func (h Handler) listScheduledTasks(c *fiber.Ctx) error {
	return c.JSON(h.scheduler.GetTasks())
}

func (h Handler) createScheduledTask(c *fiber.Ctx) error {
	var task ScheduledTask
	if err := c.BodyParser(&task); err != nil {
		return errorResp(c, 400, err)
	}
	if task.Name == "" || task.Type == "" || task.Interval == "" {
		return errorResp(c, 400, fmt.Errorf("name, type, and interval are required"))
	}
	if task.Type != "backup" && task.Type != "restart" && task.Type != "stop" {
		return errorResp(c, 400, fmt.Errorf("type must be backup, restart, or stop"))
	}
	if _, err := parseInterval(task.Interval); err != nil {
		return errorResp(c, 400, fmt.Errorf("invalid interval: %w", err))
	}
	created := h.scheduler.AddTask(task)
	return c.Status(201).JSON(created)
}

func (h Handler) updateScheduledTask(c *fiber.Ctx) error {
	id := c.Params("id")
	var task ScheduledTask
	if err := c.BodyParser(&task); err != nil {
		return errorResp(c, 400, err)
	}
	if err := h.scheduler.UpdateTask(id, task); err != nil {
		return errorResp(c, 404, err)
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h Handler) deleteScheduledTask(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.scheduler.RemoveTask(id); err != nil {
		return errorResp(c, 404, err)
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

func (h Handler) runScheduledTask(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.scheduler.RunTask(id); err != nil {
		return errorResp(c, 404, err)
	}
	return c.JSON(fiber.Map{"status": "ok"})
}
