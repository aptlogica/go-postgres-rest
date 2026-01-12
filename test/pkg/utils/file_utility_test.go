package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"go-postgres-rest/pkg/utils"
)

func TestFileUtilityCreateAndDeleteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := utils.CreateFile(path); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if !utils.Exists(path) {
		t.Fatalf("file should exist after create")
	}
	if err := utils.DeleteFile(path); err != nil {
		t.Fatalf("failed to delete file: %v", err)
	}
	if utils.Exists(path) {
		t.Fatalf("file should not exist after delete")
	}
}

func TestFileUtilityCreateAndDeleteDirRecursive(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")
	if err := utils.CreateDirRecursive(nested); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if info, err := os.Stat(nested); err != nil || !info.IsDir() {
		t.Fatalf("expected dir to exist: %v", err)
	}
	if err := utils.DeleteDirRecursive(filepath.Join(dir, "a")); err != nil {
		t.Fatalf("failed to delete dir: %v", err)
	}
	if utils.Exists(nested) {
		t.Fatalf("nested dir should be removed")
	}
}
