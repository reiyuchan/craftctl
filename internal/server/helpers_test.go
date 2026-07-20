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
	// errorResp is a function that takes (c, code, err); we can't test the HTTP part
	// without a Fiber context, but we can verify it doesn't panic when called correctly.
	// This is a compile-time check - the function signature is correct.
	_ = errorResp
}
