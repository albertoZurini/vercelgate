package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/khanakia/vercelgate/pkg/utils"
)

func TestOpenFile_Existing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.json")
	os.WriteFile(path, []byte(`{"key":"val"}`), 0600)

	got, err := utils.OpenFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != `{"key":"val"}` {
		t.Errorf("got %s", got)
	}
}

func TestOpenFile_Missing(t *testing.T) {
	_, err := utils.OpenFile(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestOpenFile_Directory(t *testing.T) {
	_, err := utils.OpenFile(t.TempDir())
	if err == nil {
		t.Error("expected error when path is a directory")
	}
}

func TestIsFileExists_Existing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(path, []byte("hello"), 0600)

	if err := utils.IsFileExists(path); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIsFileExists_Missing(t *testing.T) {
	err := utils.IsFileExists(filepath.Join(t.TempDir(), "missing.txt"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestIsFileExists_Directory(t *testing.T) {
	err := utils.IsFileExists(t.TempDir())
	if err == nil {
		t.Error("expected error when path is a directory")
	}
}
