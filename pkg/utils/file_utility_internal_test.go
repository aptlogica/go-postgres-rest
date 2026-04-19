// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	utils "github.com/aptlogica/go-postgres-rest/pkg/utils"
)

func TestFileUtilities(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "nested", "file.txt")

	// Create nested directory and file
	if err := utils.CreateDirRecursive(filepath.Dir(filePath)); err != nil {
		t.Fatalf("CreateDirRecursive error: %v", err)
	}
	if err := utils.CreateFile(filePath); err != nil {
		t.Fatalf("CreateFile error: %v", err)
	}
	if !utils.Exists(filePath) {
		t.Fatalf("file should exist after CreateFile")
	}

	// Delete file
	if err := utils.DeleteFile(filePath); err != nil {
		t.Fatalf("DeleteFile error: %v", err)
	}
	if utils.Exists(filePath) {
		t.Fatalf("file should be deleted")
	}

	// Delete directory recursively
	if err := utils.DeleteDirRecursive(filepath.Dir(filePath)); err != nil {
		t.Fatalf("DeleteDirRecursive error: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(filePath)); !os.IsNotExist(err) {
		t.Fatalf("expected dir to be deleted, got err=%v", err)
	}
}
