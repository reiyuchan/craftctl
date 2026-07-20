package server

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestContains(t *testing.T) {
	if !contains("hello world", "world") {
		t.Errorf("contains('hello world', 'world') = false, want true")
	}
	if contains("hello world", "planet") {
		t.Errorf("contains('hello world', 'planet') = true, want false")
	}
}

func TestFields(t *testing.T) {
	got := fields("a b c")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("fields('a b c') = %v, want [a b c]", got)
	}

	got2 := fields("")
	if len(got2) != 0 {
		t.Errorf("fields('') = %v, want []", got2)
	}
}

func TestMin(t *testing.T) {
	if min(1, 2) != 1 {
		t.Errorf("min(1, 2) = %d, want 1", min(1, 2))
	}
	if min(5, 3) != 3 {
		t.Errorf("min(5, 3) = %d, want 3", min(5, 3))
	}
	if min(-1, 0) != -1 {
		t.Errorf("min(-1, 0) = %d, want -1", min(-1, 0))
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0B"},
		{500, "500B"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
		{1572864, "1.5M"},
		{1073741824, "1.0G"},
		{1610612736, "1.5G"},
	}
	for _, tc := range tests {
		got := formatBytes(tc.input)
		if got != tc.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestExists(t *testing.T) {
	tmp := t.TempDir()
	f, _ := os.CreateTemp(tmp, "testfile")
	f.Close()

	if !exists(tmp, filepath.Base(f.Name())) {
		t.Error("exists(tmp, file) = false, want true")
	}
	if exists(tmp, "nonexistent") {
		t.Error("exists(tmp, 'nonexistent') = true, want false")
	}
}

func TestExistsFile(t *testing.T) {
	tmp := t.TempDir()
	f, _ := os.CreateTemp(tmp, "testfile")
	f.Close()

	if !existsFile(f.Name()) {
		t.Error("existsFile(file) = false, want true")
	}
	if existsFile(filepath.Join(tmp, "nonexistent")) {
		t.Error("existsFile('nonexistent') = true, want false")
	}
}

func TestFilePath(t *testing.T) {
	got := filePath("a", "b", "c")
	want := filepath.Join("a", "b", "c")
	if got != want {
		t.Errorf("filePath = %q, want %q", got, want)
	}

	got2 := filePath("single")
	if got2 != "single" {
		t.Errorf("filePath('single') = %q, want 'single'", got2)
	}
}

func TestFindJavaBin(t *testing.T) {
	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	os.MkdirAll(binDir, 0755)
	javaPath := filepath.Join(binDir, "java")
	if runtime.GOOS == "windows" {
		javaPath += ".exe"
	}
	os.WriteFile(javaPath, []byte("dummy"), 0644)

	got := findJavaBin(tmp)
	if got != javaPath {
		t.Errorf("findJavaBin(%q) = %q, want %q", tmp, got, javaPath)
	}

	// Non-existent dir
	got2 := findJavaBin(filepath.Join(tmp, "nonexistent"))
	if got2 != "" {
		t.Errorf("findJavaBin(nonexistent) = %q, want ''", got2)
	}

	// Nested structure (Adoptium tarball style)
	nested := filepath.Join(tmp, "jdk-21.0.1")
	os.MkdirAll(filepath.Join(nested, "bin"), 0755)
	nestedJava := filepath.Join(nested, "bin", "java")
	if runtime.GOOS == "windows" {
		nestedJava += ".exe"
	}
	os.WriteFile(nestedJava, []byte("dummy"), 0644)

	got3 := findJavaBin(tmp)
	if got3 != javaPath {
		t.Errorf("findJavaBin(%q) should prefer direct bin/java, got %q", tmp, got3)
	}

	// Only nested
	tmp2 := t.TempDir()
	os.MkdirAll(filepath.Join(tmp2, "jdk-21.0.1", "bin"), 0755)
	nestedJava2 := filepath.Join(tmp2, "jdk-21.0.1", "bin", "java")
	if runtime.GOOS == "windows" {
		nestedJava2 += ".exe"
	}
	os.WriteFile(nestedJava2, []byte("dummy"), 0644)

	got4 := findJavaBin(tmp2)
	if got4 != nestedJava2 {
		t.Errorf("findJavaBin(%q) with nested = %q, want %q", tmp2, got4, nestedJava2)
	}
}

func TestInstalledJars_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	items, err := installedJars(tmp, "mods")
	if err != nil {
		t.Fatalf("installedJars: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("installedJars items = %d, want 0", len(items))
	}
}

func TestInstalledJars_WithFiles(t *testing.T) {
	tmp := t.TempDir()
	modsDir := filepath.Join(tmp, "mods")
	os.MkdirAll(modsDir, 0755)

	// Create a jar file
	os.WriteFile(filepath.Join(modsDir, "testmod.jar"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(modsDir, "notajar.txt"), []byte("test"), 0644)

	items, err := installedJars(tmp, "mods")
	if err != nil {
		t.Fatalf("installedJars: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("installedJars items = %d, want 1", len(items))
	}
	if items[0].FileName != "testmod.jar" {
		t.Errorf("file_name = %q, want %q", items[0].FileName, "testmod.jar")
	}
}

func TestErrorRespType(t *testing.T) {
	_ = errorResp
}

func TestInstalledJars_NestedDirs(t *testing.T) {
	tmp := t.TempDir()
	modsDir := filepath.Join(tmp, "mods")
	os.MkdirAll(filepath.Join(modsDir, "sub"), 0755)
	os.WriteFile(filepath.Join(modsDir, "a.jar"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(modsDir, "sub", "b.jar"), []byte("bb"), 0644)

	items, err := installedJars(tmp, "mods")
	if err != nil {
		t.Fatalf("installedJars: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("installedJars items = %d, want 1 (subdirs ignored)", len(items))
	}
}

func TestInstalledJars_SizesAndVersion(t *testing.T) {
	tmp := t.TempDir()
	modsDir := filepath.Join(tmp, "mods")
	os.MkdirAll(modsDir, 0755)
	os.WriteFile(filepath.Join(modsDir, "big.jar"), make([]byte, 2048), 0644)

	items, err := installedJars(tmp, "mods")
	if err != nil {
		t.Fatalf("installedJars: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Version != "unknown" {
		t.Errorf("version = %q, want %q", items[0].Version, "unknown")
	}
	if items[0].Size != "2.0K" {
		t.Errorf("size = %q, want %q", items[0].Size, "2.0K")
	}
	if items[0].Name != "big" {
		t.Errorf("name = %q, want %q", items[0].Name, "big")
	}
	if items[0].Source != "Local" {
		t.Errorf("source = %q, want %q", items[0].Source, "Local")
	}
}

func TestInstalledJars_NonexistentDir(t *testing.T) {
	items, err := installedJars("/nonexistent/path/xyz", "mods")
	if err != nil {
		t.Fatalf("expected no error for nonexistent dir, got: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestDirSize(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a.txt"), make([]byte, 100), 0644)
	os.MkdirAll(filepath.Join(tmp, "sub"), 0755)
	os.WriteFile(filepath.Join(tmp, "sub", "b.txt"), make([]byte, 200), 0644)

	size, err := dirSize(tmp)
	if err != nil {
		t.Fatalf("dirSize: %v", err)
	}
	if size != 300 {
		t.Errorf("dirSize = %d, want 300", size)
	}
}

func TestDirSize_Empty(t *testing.T) {
	tmp := t.TempDir()
	size, err := dirSize(tmp)
	if err != nil {
		t.Fatalf("dirSize: %v", err)
	}
	if size != 0 {
		t.Errorf("dirSize = %d, want 0", size)
	}
}

func TestModrinthSearchFacets(t *testing.T) {
	tests := []struct {
		name        string
		loaders     []string
		gameVersion string
		wantFacets  [][]string
	}{
		{"no filters", nil, "", nil},
		{"single loader", []string{"fabric"}, "", [][]string{{"categories:fabric"}}},
		{"multiple loaders", []string{"forge", "fabric"}, "", [][]string{{"categories:forge", "categories:fabric"}}},
		{"version only", nil, "1.20.1", [][]string{{"versions:1.20.1"}}},
		{"both", []string{"forge", "fabric"}, "1.20.1", [][]string{{"categories:forge", "categories:fabric"}, {"versions:1.20.1"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildModrinthFacets(tc.loaders, tc.gameVersion)
			if len(got) != len(tc.wantFacets) {
				t.Fatalf("got %d facet groups, want %d", len(got), len(tc.wantFacets))
			}
			for i := range got {
				if len(got[i]) != len(tc.wantFacets[i]) {
					t.Fatalf("facet group %d: got %d items, want %d", i, len(got[i]), len(tc.wantFacets[i]))
				}
				for j := range got[i] {
					if got[i][j] != tc.wantFacets[i][j] {
						t.Errorf("facet[%d][%d] = %q, want %q", i, j, got[i][j], tc.wantFacets[i][j])
					}
				}
			}
		})
	}
}
