package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListCmdRegex(t *testing.T) {
	tests := []struct {
		input   string
		match   bool
		count   int
		players string
	}{
		{"There are 0/20 players online:", true, 0, ""},
		{"There are 3/20 players online: Alice, Bob, Charlie", true, 3, "Alice, Bob, Charlie"},
		{"There are 1/20 players online: Steve", true, 1, "Steve"},
		{"[12:00:00] [Server thread/INFO]: There are 2/10 players online: A, B", false, 0, ""},
		{"Something else entirely", false, 0, ""},
	}

	for _, tc := range tests {
		m := listCmdRe.FindStringSubmatch(tc.input)
		if tc.match && m == nil {
			t.Errorf("input %q: expected match, got nil", tc.input)
			continue
		}
		if !tc.match && m != nil {
			t.Errorf("input %q: expected no match, got %v", tc.input, m)
			continue
		}
		if tc.match {
			n := 0
			if m[1] != "" {
				for _, c := range m[1] {
					n = n*10 + int(c-'0')
				}
			}
			if n != tc.count {
				t.Errorf("count = %d, want %d", n, tc.count)
			}
		}
	}
}

func TestReadJSONList_OpsJson(t *testing.T) {
	dir := t.TempDir()
	data := `[{"name":"Alice","level":4,"uuid":"abc"},{"name":"Bob","level":1,"uuid":"def"}]`
	os.WriteFile(filepath.Join(dir, "ops.json"), []byte(data), 0644)

	list, err := readJSONList(dir, "ops.json")
	if err != nil {
		t.Fatalf("readJSONList: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
	if list[0]["name"] != "Alice" {
		t.Errorf("list[0].name = %v, want Alice", list[0]["name"])
	}
	if list[1]["name"] != "Bob" {
		t.Errorf("list[1].name = %v, want Bob", list[1]["name"])
	}
}

func TestReadJSONList_WhitelistJson(t *testing.T) {
	dir := t.TempDir()
	data := `[{"name":"Steve","uuid":"123"}]`
	os.WriteFile(filepath.Join(dir, "whitelist.json"), []byte(data), 0644)

	list, err := readJSONList(dir, "whitelist.json")
	if err != nil {
		t.Fatalf("readJSONList: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list))
	}
	if list[0]["name"] != "Steve" {
		t.Errorf("name = %v, want Steve", list[0]["name"])
	}
}

func TestReadJSONList_MissingFile(t *testing.T) {
	dir := t.TempDir()
	list, err := readJSONList(dir, "ops.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 entries, got %d", len(list))
	}
}

func TestReadJSONList_InvalidJson(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "ops.json"), []byte("{invalid"), 0644)
	_, err := readJSONList(dir, "ops.json")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestReadJSONList_EmptyArray(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "ops.json"), []byte("[]"), 0644)
	list, err := readJSONList(dir, "ops.json")
	if err != nil {
		t.Fatalf("readJSONList: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 entries, got %d", len(list))
	}
}
