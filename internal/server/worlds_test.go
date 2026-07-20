package server

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestZipDir_SingleFile(t *testing.T) {
	parent := t.TempDir()
	src := filepath.Join(parent, "world")
	os.MkdirAll(src, 0755)
	os.WriteFile(filepath.Join(src, "level.dat"), []byte("data"), 0644)

	dest := filepath.Join(t.TempDir(), "out.zip")
	if err := zipDir(src, dest); err != nil {
		t.Fatalf("zipDir: %v", err)
	}

	r, err := zip.OpenReader(dest)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 1 {
		t.Errorf("zip entries = %d, want 1", len(r.File))
	}
	if r.File[0].Name != "world/level.dat" {
		t.Errorf("zip entry = %q, want %q", r.File[0].Name, "world/level.dat")
	}
}

func TestZipDir_NestedFiles(t *testing.T) {
	src := t.TempDir()
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("bbb"), 0644)

	dest := filepath.Join(t.TempDir(), "out.zip")
	if err := zipDir(src, dest); err != nil {
		t.Fatalf("zipDir: %v", err)
	}

	r, err := zip.OpenReader(dest)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 2 {
		t.Errorf("zip entries = %d, want 2", len(r.File))
	}
}

func TestZipDir_NonexistentSource(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.zip")
	err := zipDir("/nonexistent/path/xyz", dest)
	if err == nil {
		t.Error("expected error for nonexistent source, got nil")
	}
}

func TestDirSize_World(t *testing.T) {
	world := t.TempDir()
	os.WriteFile(filepath.Join(world, "level.dat"), make([]byte, 500), 0644)
	os.MkdirAll(filepath.Join(world, "region"), 0755)
	os.WriteFile(filepath.Join(world, "region", "r.0.0.mca"), make([]byte, 1000), 0644)

	size, err := dirSize(world)
	if err != nil {
		t.Fatalf("dirSize: %v", err)
	}
	if size != 1500 {
		t.Errorf("dirSize = %d, want 1500", size)
	}
}
