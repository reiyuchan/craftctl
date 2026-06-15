package server

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/reiyuchan/ctlcraft/internal/mc"
)

var (
	errNoServerJar = errors.New("no server.jar found. Please download a server first")
)

var httpClient = resty.New().
	SetHeader("User-Agent", "ctlcraft/0.1.0").
	SetTimeout(30_000_000_000)

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

func javaSearchPaths() []string {
	var paths []string
	if jh := os.Getenv("JAVA_HOME"); jh != "" {
		paths = append(paths, jh)
	}
	paths = append(paths,
		"/usr/lib/jvm",
		"/opt/java",
		"/opt/openjdk",
		"C:\\Program Files\\Java",
		"C:\\Program Files\\Eclipse Adoptium",
		"C:\\Program Files\\Amazon Corretto",
	)
	return paths
}

func javaBinPath(base, name string) string {
	java := filepath.Join(base, name, "bin", "java")
	if runtime.GOOS == "windows" {
		java += ".exe"
	}
	return java
}

func javaVersion(javaBin string) string {
	out, _ := exec.Command(javaBin, "-version").CombinedOutput()
	lines := strings.SplitN(string(out), "\n", 2)
	ver := lines[0]
	if len(lines) > 1 {
		ver = strings.Join(lines[:min(2, len(lines))], " | ")
	}
	return fmt.Sprintf("%s: %s", javaBin, ver)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

func modrinthSearch(query string, loaders []string, gameVersion string) ([]fiber.Map, error) {
	var facets []string
	if len(loaders) > 0 {
		lfs := make([]string, len(loaders))
		for i, l := range loaders {
			lfs[i] = fmt.Sprintf("categories:%s", strings.ToLower(l))
		}
		facets = append(facets, fmt.Sprintf("(%s)", strings.Join(lfs, ",")))
	}
	if gameVersion != "" {
		facets = append(facets, fmt.Sprintf("(versions:%s)", gameVersion))
	}

	params := map[string]string{"query": query, "limit": "30"}
	if len(facets) > 0 {
		params["facets"] = strings.Join(facets, "")
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
		latest := ""
		if len(hit.Versions) > 0 {
			latest = hit.Versions[0]
		}
		items[i] = fiber.Map{
			"id":             hit.ProjectID,
			"slug":           hit.Slug,
			"title":          hit.Title,
			"description":    hit.Description,
			"author":         hit.Author,
			"downloads":      fmt.Sprintf("%d", hit.Downloads),
			"latest_version": latest,
			"icon_url":       iconURL,
			"categories":     hit.Categories,
			"loaders":        hit.Loaders,
			"source":         "Modrinth",
		}
	}
	return items, nil
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
		Icon        string   `json:"icon"`
		Categories  []string `json:"categories"`
		Loaders     []string `json:"loaders"`
	} `json:"hits"`
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

	resp, err = httpClient.R().SetOutput(outPath).Get(file.URL)
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
	resp, err := httpClient.R().SetOutput(outPath).Get(downloadURL)
	if err != nil || !resp.IsSuccess() {
		return "", errors.New("download failed")
	}
	return outPath, nil
}

// ── Installed JARs ────────────────────────────────────────────────────────

func installedJars(serverDir, subdir string) ([]fiber.Map, error) {
	dir := filepath.Join(serverDir, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []fiber.Map{}, nil
		}
		return nil, err
	}

	var items []fiber.Map
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jar") {
			continue
		}
		info, _ := entry.Info()
		size := "unknown"
		if info != nil {
			size = formatBytes(info.Size())
		}
		items = append(items, fiber.Map{
			"file_name": entry.Name(),
			"name":      strings.TrimSuffix(entry.Name(), ".jar"),
			"version":   "unknown",
			"size":      size,
			"source":    "Local",
		})
	}
	if items == nil {
		items = []fiber.Map{}
	}
	return items, nil
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

	resp, err = httpClient.R().SetOutput(fabricPath).Get(jarURL)
	if err != nil {
		return "", fmt.Errorf("fabric download: %w", err)
	}

	// Try server launcher
	var launcher struct{ URL string }
	resp2, _ := httpClient.R().SetResult(&launcher).Get(
		"https://meta.fabricmc.cn/v2/versions/loader/" + mcVersion + "/" + loader.Loader.Version + "/server/legacy")
	if resp2.IsSuccess() && launcher.URL != "" {
		httpClient.R().SetOutput(filepath.Join(modsDir, "fabric-server-launch.jar")).Get(launcher.URL)
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

	resp, err = httpClient.R().SetOutput(outPath).Get(link)
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
