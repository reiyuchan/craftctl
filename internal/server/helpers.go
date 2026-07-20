package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/reiyuchan/craftctl/internal/mc"
)

var (
	errNoServerJar = errors.New("no server.jar found. Please download a server first")
)

var httpClient = resty.New().
	SetHeader("User-Agent", "craftctl/0.1.0").
	SetTimeout(30_000_000_000)

var downloadClient = resty.New().
	SetHeader("User-Agent", "craftctl/0.1.0").
	SetTimeout(0) // no timeout — TCP keepalive handles dead connections

// ── HTTP helpers ───────────────────────────────────────────────────────────

func errorResp(c *fiber.Ctx, code int, err error) error {
	return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}

// ── File helpers ──────────────────────────────────────────────────────────

func exists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func existsFile(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func filePath(parts ...string) string {
	return filepath.Join(parts...)
}

// ── Server properties ──────────────────────────────────────────────────────

type serverProperties = mc.ServerProperties

var readServerProperties = mc.ReadServerProperties
var writeServerProperties = mc.WriteServerProperties

// ── Start options ─────────────────────────────────────────────────────────

type startOpts struct {
	JavaPath string `json:"javaPath"`
	MinRAM   string `json:"minRam"`
	MaxRAM   string `json:"maxRam"`
	JVMFlags string `json:"jvmFlags"`
}

// ── Java detection ────────────────────────────────────────────────────────

type javaInstallation struct {
	ID           string `json:"id"`
	Vendor       string `json:"vendor"`
	MajorVersion int    `json:"majorVersion"`
	FullVersion  string `json:"fullVersion"`
	LatestVersion string `json:"latestVersion"`
	Arch         string `json:"arch"`
	InstallPath  string `json:"installPath"`
	SizeOnDisk   string `json:"sizeOnDisk"`
	Status       string `json:"status"`
	IsActive     bool   `json:"isActive"`
	ReleaseType  string `json:"releaseType"`
}

func detectJavaRuntimes(javaDir string) []javaInstallation {
	seen := make(map[string]bool)
	var runtimes []javaInstallation

	// System paths
	paths := []string{
		os.Getenv("JAVA_HOME"),
		"/usr/lib/jvm",
		"/opt/java",
		"/opt/openjdk",
		"/usr/local/opt/openjdk",
		"C:\\Program Files\\Java",
		"C:\\Program Files\\Eclipse Adoptium",
		"C:\\Program Files\\Amazon Corretto",
	}
	for _, dir := range paths {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			bin := filepath.Join(dir, e.Name(), "bin", "java")
			if runtime.GOOS == "windows" {
				bin += ".exe"
			}
			if existsFile(bin) && !seen[bin] {
				seen[bin] = true
				info := readJavaInfo(bin)
				info.ID = fmt.Sprintf("sys-%d", len(runtimes))
				info.InstallPath = bin
				runtimes = append(runtimes, info)
			}
		}
	}

	// Downloaded runtimes
	if javaDir != "" {
		entries, err := os.ReadDir(javaDir)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				bin := findJavaBin(filepath.Join(javaDir, e.Name()))
				if bin != "" && !seen[bin] {
					seen[bin] = true
					info := readJavaInfo(bin)
					info.ID = fmt.Sprintf("dl-%d", len(runtimes))
					info.InstallPath = bin
					if size, err := dirSize(filepath.Dir(filepath.Dir(bin))); err == nil {
						info.SizeOnDisk = formatBytes(size)
					}
					runtimes = append(runtimes, info)
				}
			}
		}
	}

	return runtimes
}

func findJavaBin(dir string) string {
	bin := filepath.Join(dir, "bin", "java")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if existsFile(bin) {
		return bin
	}
	// Check one level deeper (Adoptium tarballs have a nested dir)
	subs, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, s := range subs {
		if !s.IsDir() {
			continue
		}
		bin = filepath.Join(dir, s.Name(), "bin", "java")
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}
		if existsFile(bin) {
			return bin
		}
	}
	return ""
}

func readJavaInfo(javaBin string) javaInstallation {
	out, _ := exec.Command(javaBin, "-version").CombinedOutput()
	output := string(out)
	lines := strings.SplitN(output, "\n", 3)

	verLine := ""
	if len(lines) > 0 {
		verLine = strings.TrimSpace(lines[0])
	}
	fullVersion := verLine
	if len(lines) > 1 {
		fullVersion = verLine + " | " + strings.TrimSpace(lines[1])
	}

	majorVersion := 0
	if idx := strings.Index(verLine, `version "`); idx >= 0 {
		rest := verLine[idx+len(`version "`):]
		if end := strings.Index(rest, `"`); end >= 0 {
			verStr := rest[:end]
			parts := strings.SplitN(verStr, ".", 2)
			if v, err := strconv.Atoi(parts[0]); err == nil {
				majorVersion = v
				if majorVersion == 1 && len(parts) > 1 {
					if parts2 := strings.SplitN(parts[1], ".", 2); len(parts2) > 0 {
						if v2, err := strconv.Atoi(parts2[0]); err == nil {
							majorVersion = v2
						}
					}
				}
			}
		}
	}

	vendor := "OpenJDK"
	low := strings.ToLower(output)
	switch {
	case strings.Contains(low, " adoptium") || strings.Contains(low, "adoptium") || strings.Contains(low, "temurin"):
		vendor = "Adoptium"
	case strings.Contains(low, "corretto"):
		vendor = "Amazon Corretto"
	case strings.Contains(low, "zulu"):
		vendor = "Azul Zulu"
	case strings.Contains(low, "java(tm)") || strings.Contains(low, "hotspot"):
		vendor = "Oracle"
	case strings.Contains(low, "openj9"):
		vendor = "IBM Semeru"
	case strings.Contains(low, "microsoft"):
		vendor = "Microsoft"
	}

	arch := runtime.GOARCH
	switch {
	case strings.Contains(output, "AArch64") || strings.Contains(output, "aarch64"):
		arch = "aarch64"
	case strings.Contains(output, "64-Bit"):
		arch = "x64"
	case strings.Contains(output, "32-Bit"):
		arch = "x32"
	}

	releaseType := "STS"
	if majorVersion >= 17 || strings.Contains(output, "LTS") || strings.Contains(low, "lts") {
		releaseType = "LTS"
	}

	return javaInstallation{
		Vendor:       vendor,
		MajorVersion: majorVersion,
		FullVersion:  fullVersion,
		Arch:         arch,
		Status:       "installed",
		ReleaseType:  releaseType,
	}
}

func dirSize(path string) (int64, error) {
	var total int64
	err := filepath.Walk(path, func(_ string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			total += fi.Size()
		}
		return nil
	})
	return total, err
}

// ── Folder ────────────────────────────────────────────────────────────────

func openFolder(dir string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd, args = "explorer", []string{"/select,", dir}
	case "darwin":
		cmd, args = "open", []string{dir}
	default:
		cmd, args = "xdg-open", []string{dir}
	}
	exec.Command(cmd, args...).Start()
}

// ── Modrinth ─────────────────────────────────────────────────────────────

type modSearchReq struct {
	Query       string   `json:"query"`
	Loaders     []string `json:"loaders"`
	GameVersion string   `json:"gameVersion"`
}

func buildModrinthFacets(loaders []string, gameVersion string) [][]string {
	var facets [][]string
	if len(loaders) > 0 {
		loaderFacets := make([]string, len(loaders))
		for i, l := range loaders {
			loaderFacets[i] = fmt.Sprintf("categories:%s", strings.ToLower(l))
		}
		facets = append(facets, loaderFacets)
	}
	if gameVersion != "" {
		facets = append(facets, []string{fmt.Sprintf("versions:%s", gameVersion)})
	}
	return facets
}

func modrinthSearch(query string, loaders []string, gameVersion string) ([]fiber.Map, error) {
	params := map[string]string{"query": query, "limit": "30"}
	if facets := buildModrinthFacets(loaders, gameVersion); len(facets) > 0 {
		b, err := json.Marshal(facets)
		if err != nil {
			return nil, fmt.Errorf("modrinth facets marshal: %w", err)
		}
		params["facets"] = string(b)
	}

	var result modrinthSearchResult
	resp, err := httpClient.R().SetQueryParams(params).SetResult(&result).Get(
		"https://api.modrinth.com/v2/search")
	if err != nil {
		return nil, fmt.Errorf("modrinth search: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("modrinth: status %d", resp.StatusCode())
	}

	items := make([]fiber.Map, len(result.Hits))
	for i, hit := range result.Hits {
		iconURL := ""
		if hit.Icon != "" {
			iconURL = "https://cdn.modrinth.com/" + hit.Icon
		}
		items[i] = fiber.Map{
			"id":             hit.ProjectID,
			"slug":           hit.Slug,
			"title":          hit.Title,
			"description":    hit.Description,
			"author":         hit.Author,
			"downloads":      fmt.Sprintf("%d", hit.Downloads),
			"latest_version": hit.LatestVersion,
			"icon_url":       iconURL,
			"categories":     hit.Categories,
			"loaders":        hit.Loaders,
			"source":         "Modrinth",
		}
	}
	return items, nil
}

func modrinthSearchURL(query string, loaders []string, gameVersion string) string {
	base := "https://api.modrinth.com/v2/search"
	params := map[string]string{"query": query, "limit": "30"}
	if facets := buildModrinthFacets(loaders, gameVersion); len(facets) > 0 {
		b, _ := json.Marshal(facets)
		params["facets"] = string(b)
	}
	// Build URL manually to show the facets JSON for debugging
	url := base + "?query=" + query + "&limit=30"
	if f, ok := params["facets"]; ok {
		url += "&facets=" + f
	}
	return url
}

type modrinthSearchResult struct {
	Hits []struct {
		ProjectID   string   `json:"project_id"`
		Slug        string   `json:"slug"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Author      string   `json:"author"`
		Downloads   int64    `json:"downloads"`
		Versions    []string `json:"versions"`
		Icon           string   `json:"icon_url"`
		LatestVersion  string   `json:"latest_version"`
		Categories     []string `json:"categories"`
		Loaders        []string `json:"loaders"` // always nil from search endpoint; frontend handles gracefully
	} `json:"hits"`
}

func popularMods() ([]fiber.Map, error) {
	facets := `[["categories:fabric"],["project_type:mod"]]`
	var result modrinthSearchResult
	resp, err := httpClient.R().
		SetQueryParams(map[string]string{"limit": "12", "index": "downloads", "facets": facets}).
		SetResult(&result).
		Get("https://api.modrinth.com/v2/search")
	if err != nil {
		return nil, fmt.Errorf("popular mods: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("popular mods: status %d", resp.StatusCode())
	}

	items := make([]fiber.Map, len(result.Hits))
	for i, hit := range result.Hits {
		iconURL := ""
		if hit.Icon != "" {
			iconURL = "https://cdn.modrinth.com/" + hit.Icon
		}
		items[i] = fiber.Map{
			"id":             hit.ProjectID,
			"slug":           hit.Slug,
			"title":          hit.Title,
			"description":    hit.Description,
			"author":         hit.Author,
			"downloads":      fmt.Sprintf("%d", hit.Downloads),
			"latest_version": hit.LatestVersion,
			"icon_url":       iconURL,
			"categories":     hit.Categories,
			"loaders":        hit.Loaders,
			"source":         "Modrinth",
		}
	}
	return items, nil
}

func popularPlugins() ([]fiber.Map, error) {
	var items []fiber.Map

	facets := `[["categories:paper"],["project_type:plugin"]]`
	var mr modrinthSearchResult
	resp, _ := httpClient.R().
		SetQueryParams(map[string]string{"limit": "12", "index": "downloads", "facets": facets}).
		SetResult(&mr).
		Get("https://api.modrinth.com/v2/search")

	if resp != nil && resp.IsSuccess() {
		for _, hit := range mr.Hits {
			iconURL := ""
			if hit.Icon != "" {
				iconURL = "https://cdn.modrinth.com/" + hit.Icon
			}
			latest := hit.LatestVersion
			items = append(items, fiber.Map{
				"id":             hit.ProjectID,
				"slug":           hit.Slug,
				"name":           hit.Title,
				"description":    hit.Description,
				"author":         hit.Author,
				"downloads":      fmt.Sprintf("%d", hit.Downloads),
				"latest_version": latest,
				"icon_url":       iconURL,
				"categories":     hit.Categories,
				"loaders":        hit.Loaders,
				"source":         "Modrinth",
			})
		}
	}

	var hg struct {
		Result []struct {
			Slug        string `json:"slug"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Namespace   struct {
				Owner string `json:"owner"`
				Slug  string `json:"slug"`
			} `json:"namespace"`
			AvatarURL string `json:"avatarUrl"`
			Stats     struct {
				Downloads int `json:"downloads"`
			} `json:"stats"`
		} `json:"result"`
	}
	resp2, _ := httpClient.R().
		SetQueryParams(map[string]string{"platform": "PAPER", "sort": "downloads", "limit": "12"}).
		SetResult(&hg).
		Get("https://hangar.papermc.io/api/v1/projects")

	if resp2 != nil && resp2.IsSuccess() {
		for _, p := range hg.Result {
			items = append(items, fiber.Map{
				"id":             p.Slug,
				"slug":           p.Slug,
				"name":           p.Name,
				"description":    p.Description,
				"author":         p.Namespace.Owner,
				"downloads":      fmt.Sprintf("%d", p.Stats.Downloads),
				"latest_version": "",
				"icon_url":       p.AvatarURL,
				"categories":     []string{},
				"loaders":        []string{"Paper"},
				"source":         "Hangar",
			})
		}
	}

	if items == nil {
		items = []fiber.Map{}
	}
	return items, nil
}

func modrinthVersions(id string) ([]modrinthVersion, error) {
	var versions []modrinthVersion
	resp, err := httpClient.R().SetResult(&versions).Get(
		"https://api.modrinth.com/v2/project/" + id + "/version")
	if err != nil {
		return nil, fmt.Errorf("modrinth versions: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("modrinth: status %d", resp.StatusCode())
	}
	return versions, nil
}

type modrinthVersion struct {
	ID            string   `json:"id"`
	VersionNumber string   `json:"version_number"`
	GameVersions  []string `json:"game_versions"`
	Loaders       []string `json:"loaders"`
}

func modrinthDownload(projectID, versionID, modsDir string) (string, error) {
	var versions []struct {
		ID    string `json:"id"`
		Files []struct {
			URL      string `json:"url"`
			Filename string `json:"filename"`
			Primary  bool   `json:"primary"`
		} `json:"files"`
	}

	resp, err := httpClient.R().SetResult(&versions).Get(
		"https://api.modrinth.com/v2/project/" + projectID + "/version")
	if err != nil {
		return "", fmt.Errorf("modrinth versions: %w", err)
	}
	if !resp.IsSuccess() || len(versions) == 0 {
		return "", errors.New("no versions found")
	}

	target := versions[0]
	if versionID != "" {
		for _, v := range versions {
			if v.ID == versionID {
				target = v
				break
			}
		}
	}

	file := target.Files[0]
	for _, f := range target.Files {
		if f.Primary {
			file = f
			break
		}
	}

	os.MkdirAll(modsDir, 0o755)
	outPath := filepath.Join(modsDir, file.Filename)

	resp, err = downloadClient.R().SetOutput(outPath).Get(file.URL)
	if err != nil {
		return "", fmt.Errorf("download mod: %w", err)
	}
	if !resp.IsSuccess() {
		return "", errors.New("download failed")
	}
	return outPath, nil
}

// ── Plugin search ─────────────────────────────────────────────────────────

func pluginSearch(query string) ([]fiber.Map, error) {
	var items []fiber.Map

	// Modrinth
	var mr struct {
		Hits []struct {
			ProjectID   string   `json:"project_id"`
			Slug        string   `json:"slug"`
			Title       string   `json:"title"`
			Description string   `json:"description"`
			Author      string   `json:"author"`
			Downloads   int64    `json:"downloads"`
			Versions    []string `json:"versions"`
			Icon        string   `json:"icon"`
			Categories  []string `json:"categories"`
			Loaders     []string `json:"loaders"`
		} `json:"hits"`
	}
	resp, _ := httpClient.R().
		SetQueryParams(map[string]string{"query": query, "limit": "15"}).
		SetResult(&mr).
		Get("https://api.modrinth.com/v2/search")

	if resp != nil && resp.IsSuccess() {
		for _, hit := range mr.Hits {
			iconURL := ""
			if hit.Icon != "" {
				iconURL = "https://cdn.modrinth.com/" + hit.Icon
			}
			latest := ""
			if len(hit.Versions) > 0 {
				latest = hit.Versions[0]
			}
			items = append(items, fiber.Map{
				"id":             hit.ProjectID,
				"slug":           hit.Slug,
				"name":           hit.Title,
				"description":    hit.Description,
				"author":         hit.Author,
				"downloads":      fmt.Sprintf("%d", hit.Downloads),
				"latest_version": latest,
				"icon_url":       iconURL,
				"categories":     hit.Categories,
				"loaders":        hit.Loaders,
				"source":         "Modrinth",
			})
		}
	}

	// Hangar
	var hg struct {
		Result []struct {
			Slug        string `json:"slug"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Owner       string `json:"owner"`
			VersionTag  string `json:"version_tag"`
			AvatarURL   string `json:"avatar_url"`
			Stats       struct {
				Downloads int `json:"downloads"`
			} `json:"stats"`
		} `json:"result"`
	}
	resp2, _ := httpClient.R().
		SetQueryParams(map[string]string{"search": query, "platform": "PAPER", "page": "0", "size": "10"}).
		SetResult(&hg).
		Get("https://hangar.papermc.io/api/v1/projects")

	if resp2 != nil && resp2.IsSuccess() {
		for _, p := range hg.Result {
			items = append(items, fiber.Map{
				"id":             p.Slug,
				"slug":           p.Slug,
				"name":           p.Name,
				"description":    p.Description,
				"author":         p.Owner,
				"downloads":      fmt.Sprintf("%d", p.Stats.Downloads),
				"latest_version": p.VersionTag,
				"icon_url":       p.AvatarURL,
				"categories":     []string{},
				"loaders":        []string{"Paper"},
				"source":         "Hangar",
			})
		}
	}

	if items == nil {
		items = []fiber.Map{}
	}
	return items, nil
}

type pluginDownloadReq struct {
	Slug    string `json:"slug"`
	Version string `json:"version"`
	Source  string `json:"source"`
}

func pluginDownload(slug, version, source, pluginsDir string) (string, error) {
	os.MkdirAll(pluginsDir, 0o755)

	var downloadURL, filename string

	if source == "Hangar" {
		ver := version
		if ver == "" {
			ver = "latest"
		}
		var result struct {
			DownloadURL string `json:"downloadUrl"`
		}
		resp, err := httpClient.R().SetResult(&result).Get(
			"https://hangar.papermc.io/api/v1/projects/" + slug + "/version/" + ver)
		if err != nil || !resp.IsSuccess() {
			return "", errors.New("hangar download failed")
		}
		downloadURL = result.DownloadURL
		filename = slug + ".jar"
	} else {
		var versions []struct {
			Files []struct {
				URL      string `json:"url"`
				Filename string `json:"filename"`
				Primary  bool   `json:"primary"`
			} `json:"files"`
		}
		resp, err := httpClient.R().SetResult(&versions).Get(
			"https://api.modrinth.com/v2/project/" + slug + "/version")
		if err != nil || !resp.IsSuccess() || len(versions) == 0 {
			return "", errors.New("plugin not found")
		}

		file := versions[0].Files[0]
		for _, f := range versions[0].Files {
			if f.Primary {
				file = f
				break
			}
		}
		downloadURL = file.URL
		filename = file.Filename
	}

	outPath := filepath.Join(pluginsDir, filename)
	resp, err := downloadClient.R().SetOutput(outPath).Get(downloadURL)
	if err != nil || !resp.IsSuccess() {
		return "", errors.New("download failed")
	}
	return outPath, nil
}

// ── Installed JARs ────────────────────────────────────────────────────────

type InstalledJarInfo struct {
	FileName      string `json:"file_name"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Size          string `json:"size"`
	Source        string `json:"source"`
	ProjectID     string `json:"projectId,omitempty"`
	Slug          string `json:"slug,omitempty"`
	HasUpdate     bool   `json:"hasUpdate"`
	LatestVersion string `json:"latestVersion"`
}

func installedJars(serverDir, subdir string) ([]InstalledJarInfo, error) {
	dir := filepath.Join(serverDir, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []InstalledJarInfo{}, nil
		}
		return nil, err
	}

	var items []InstalledJarInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jar") {
			continue
		}
		info, _ := entry.Info()
		size := "unknown"
		if info != nil {
			size = formatBytes(info.Size())
		}
		items = append(items, InstalledJarInfo{
			FileName:      entry.Name(),
			Name:          strings.TrimSuffix(entry.Name(), ".jar"),
			Version:       "unknown",
			Size:          size,
			Source:        "Local",
			HasUpdate:     false,
			LatestVersion: "",
		})
	}
	if items == nil {
		items = []InstalledJarInfo{}
	}
	return items, nil
}

func checkModrinthUpdates(items []InstalledJarInfo) ([]InstalledJarInfo, error) {
	result := make([]InstalledJarInfo, len(items))
	copy(result, items)

	for i, item := range items {
		slug := strings.TrimSuffix(item.FileName, ".jar")
		versions, err := modrinthVersions(slug)
		if err != nil {
			continue
		}
		if len(versions) == 0 {
			continue
		}
		latest := versions[0].VersionNumber
		result[i].LatestVersion = latest
		result[i].ProjectID = slug
		if item.Version != "unknown" && item.Version != latest {
			result[i].HasUpdate = true
		}
	}

	return result, nil
}

func checkPluginUpdatesRemote(items []InstalledJarInfo) ([]InstalledJarInfo, error) {
	result := make([]InstalledJarInfo, len(items))
	copy(result, items)

	for i, item := range items {
		slug := strings.TrimSuffix(item.FileName, ".jar")

		// Try Modrinth first
		mrVersions, err := modrinthVersions(slug)
		if err == nil && len(mrVersions) > 0 {
			latest := mrVersions[0].VersionNumber
			result[i].LatestVersion = latest
			result[i].ProjectID = slug
			result[i].Source = "Modrinth"
			if item.Version != "unknown" && item.Version != latest {
				result[i].HasUpdate = true
			}
			continue
		}

		// Try Hangar
		var hgResult struct {
			Result []struct {
				Slug       string `json:"slug"`
				VersionTag string `json:"version_tag"`
			} `json:"result"`
		}
		resp, err := httpClient.R().SetResult(&hgResult).Get(
			"https://hangar.papermc.io/api/v1/projects/" + slug)
		if err == nil && resp.IsSuccess() && len(hgResult.Result) > 0 {
			latest := hgResult.Result[0].VersionTag
			result[i].LatestVersion = latest
			result[i].Slug = slug
			result[i].Source = "Hangar"
			if item.Version != "unknown" && item.Version != latest {
				result[i].HasUpdate = true
			}
			continue
		}
	}

	return result, nil
}

func modrinthDownloadLatest(projectID, modsDir string) (string, error) {
	return modrinthDownload(projectID, "", modsDir)
}

func pluginDownloadLatest(slug, source, pluginsDir string) (string, error) {
	return pluginDownload(slug, "", source, pluginsDir)
}

// ── Paper ────────────────────────────────────────────────────────────────

func paperBuilds(mcVersion string) ([]int, error) {
	var result struct {
		Builds []int `json:"builds"`
	}
	resp, err := httpClient.R().SetResult(&result).Get(
		"https://api.papermc.io/v2/projects/paper/versions/" + mcVersion)
	if err != nil {
		return nil, fmt.Errorf("paper builds: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("paper: version not found")
	}
	return result.Builds, nil
}

func paperDownloadURL(mcVersion, build string) (string, error) {
	var result struct {
		Builds []struct {
			Build     int `json:"build"`
			Downloads struct {
				Application struct {
					Name string `json:"name"`
				} `json:"application"`
			} `json:"downloads"`
		} `json:"builds"`
	}
	resp, err := httpClient.R().SetResult(&result).Get(
		fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s/builds/%s", mcVersion, build))
	if err != nil {
		return "", fmt.Errorf("paper download url: %w", err)
	}
	if !resp.IsSuccess() || len(result.Builds) == 0 {
		return "", errors.New("build not found")
	}
	target := result.Builds[len(result.Builds)-1]
	return fmt.Sprintf(
		"https://api.papermc.io/v2/projects/paper/versions/%s/builds/%s/downloads/%s",
		mcVersion, build, target.Downloads.Application.Name), nil
}

// ── Vanilla ─────────────────────────────────────────────────────────────

func vanillaVersions() ([]fiber.Map, error) {
	var result struct {
		Versions []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			URL  string `json:"url"`
		} `json:"versions"`
	}
	resp, err := httpClient.R().SetResult(&result).Get(
		"https://launchermeta.mojang.com/mc/game/version_manifest.json")
	if err != nil {
		return nil, fmt.Errorf("vanilla versions: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, errors.New("failed to fetch versions")
	}

	items := make([]fiber.Map, 0, 20)
	for i, v := range result.Versions {
		if i >= 20 {
			break
		}
		items = append(items, fiber.Map{"id": v.ID, "type": v.Type, "url": v.URL})
	}
	return items, nil
}

func vanillaDownloadURL(versionURL string) (string, error) {
	var result struct {
		Downloads struct {
			Server struct {
				URL string `json:"url"`
			} `json:"server"`
		} `json:"downloads"`
	}
	resp, err := httpClient.R().SetResult(&result).Get(versionURL)
	if err != nil {
		return "", fmt.Errorf("vanilla url: %w", err)
	}
	if !resp.IsSuccess() {
		return "", errors.New("version not found")
	}
	return result.Downloads.Server.URL, nil
}

// ── Fabric ───────────────────────────────────────────────────────────────

func fabricInstall(mcVersion, modsDir string) (string, error) {
	os.MkdirAll(modsDir, 0o755)

	var loaderResult [][]struct {
		Loader struct {
			Version string `json:"version"`
			Hash    string `json:"hash"`
		} `json:"loader"`
		Maven string `json:"maven"`
	}
	resp, err := httpClient.R().SetResult(&loaderResult).Get(
		"https://meta.fabricmc.cn/v2/versions/loader/" + mcVersion)
	if err != nil {
		return "", fmt.Errorf("fabric loader: %w", err)
	}
	if !resp.IsSuccess() || len(loaderResult) == 0 || len(loaderResult[0]) == 0 {
		return "", fmt.Errorf("no fabric version found for %s", mcVersion)
	}

	loader := loaderResult[0][0]
	jarURL := fmt.Sprintf("%s/%s/%s-%s.jar",
		loader.Maven, loader.Loader.Version, loader.Loader.Version, loader.Loader.Hash)
	fabricPath := filepath.Join(modsDir, fmt.Sprintf("fabric-%s.jar", loader.Loader.Version))

	resp, err = downloadClient.R().SetOutput(fabricPath).Get(jarURL)
	if err != nil {
		return "", fmt.Errorf("fabric download: %w", err)
	}

	// Try server launcher
	var launcher struct{ URL string }
	resp2, _ := httpClient.R().SetResult(&launcher).Get(
		"https://meta.fabricmc.cn/v2/versions/loader/" + mcVersion + "/" + loader.Loader.Version + "/server/legacy")
	if resp2.IsSuccess() && launcher.URL != "" {
		downloadClient.R().SetOutput(filepath.Join(modsDir, "fabric-server-launch.jar")).Get(launcher.URL)
	}

	return fabricPath, nil
}

// ── Download ─────────────────────────────────────────────────────────────

func downloadFile(client *resty.Client, url, dest string) (string, error) {
	os.MkdirAll(filepath.Dir(dest), 0o755)
	resp, err := client.R().SetOutput(dest).Get(url)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	if !resp.IsSuccess() {
		return "", fmt.Errorf("download: status %d", resp.StatusCode())
	}
	return dest, nil
}

// ── Misc ──────────────────────────────────────────────────────────────────

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func fields(s string) []string {
	return strings.Fields(s)
}

func formatBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1fG", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1fK", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%dB", n)
	}
}

// ── Java Downloads (Adoptium) ─────────────────────────────────────────────────

func adoptiumReleases() ([]fiber.Map, error) {
	var releases struct {
		AvailableLTS      []int `json:"available_lts_releases"`
		AvailableReleases []int `json:"available_releases"`
		MostRecentLTS     int   `json:"most_recent_lts"`
		MostRecentFeature int   `json:"most_recent_feature_release"`
	}
	resp, err := httpClient.R().SetResult(&releases).Get(
		"https://api.adoptium.net/v3/info/available_releases")
	if err != nil {
		return nil, fmt.Errorf("adoptium releases: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("adoptium: status %d", resp.StatusCode())
	}

	ltsSet := make(map[int]bool)
	for _, v := range releases.AvailableLTS {
		ltsSet[v] = true
	}

	var items []fiber.Map
	for _, ver := range releases.AvailableReleases {
		items = append(items, fiber.Map{
			"version": ver,
			"lts":     ltsSet[ver],
		})
	}

	return items, nil
}

type assetRelease struct {
	Binary struct {
		Package struct {
			Link string `json:"link"`
			Name string `json:"name"`
		} `json:"package"`
	} `json:"binary"`
}

func adoptiumDownload(version, installDir string) (string, error) {
	osName := "linux"
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		osName = "windows"
		ext = "zip"
	}

	var assets []assetRelease
	resp, err := httpClient.R().SetResult(&assets).Get(
		fmt.Sprintf("https://api.adoptium.net/v3/assets/latest/%s/hotspot?architecture=x64&image_type=jdk&os=%s&vendor=eclipse",
			version, osName))
	if err != nil {
		return "", fmt.Errorf("adoptium lookup: %w", err)
	}
	if !resp.IsSuccess() || len(assets) == 0 {
		return "", fmt.Errorf("adoptium: no asset found for version %s", version)
	}

	link := assets[0].Binary.Package.Link
	if link == "" {
		return "", fmt.Errorf("adoptium: empty download link for version %s", version)
	}

	os.MkdirAll(installDir, 0o755)

	filename := fmt.Sprintf("jdk-%s.%s", version, ext)
	outPath := filepath.Join(installDir, filename)

	resp, err = downloadClient.R().SetOutput(outPath).Get(link)
	if err != nil {
		return "", fmt.Errorf("download java: %w", err)
	}
	if !resp.IsSuccess() {
		return "", fmt.Errorf("download failed: %d", resp.StatusCode())
	}

	if ext == "zip" {
		unzipPath := filepath.Join(installDir, fmt.Sprintf("jdk-%s", version))
		if err := unzip(outPath, unzipPath); err != nil {
			return "", fmt.Errorf("unzip: %w", err)
		}
		os.Remove(outPath)
		return unzipPath, nil
	}

	extractPath := filepath.Join(installDir, fmt.Sprintf("jdk-%s", version))
	os.MkdirAll(extractPath, 0o755)
	if err := untargz(outPath, extractPath); err != nil {
		return "", fmt.Errorf("extract: %w", err)
	}
	os.Remove(outPath)
	return extractPath, nil
}

func unzip(src, dest string) error {
	cmd := exec.Command("unzip", "-q", src, "-d", dest)
	return cmd.Run()
}

func untargz(src, dest string) error {
	cmd := exec.Command("tar", "-xzf", src, "-C", dest)
	return cmd.Run()
}
