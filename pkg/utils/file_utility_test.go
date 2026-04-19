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

// Additional coverage for error-free paths (existing core tests live in file_utility_internal_test.go).
func TestFileUtilitiesAdditional(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "demo.txt")

	if err := utils.CreateFile(file); err != nil {
		t.Fatalf("CreateFile err: %v", err)
	}
	if info, err := os.Stat(file); err != nil || info.IsDir() {
		t.Fatalf("expected file to exist, err=%v", err)
	}

	nestedDir := filepath.Join(dir, "nested", "child")
	if err := utils.CreateDirRecursive(nestedDir); err != nil {
		t.Fatalf("CreateDirRecursive err: %v", err)
	}

	// Deleting parent should remove nested paths without error
	if err := utils.DeleteDirRecursive(filepath.Join(dir, "nested")); err != nil {
		t.Fatalf("DeleteDirRecursive err: %v", err)
	}
	if utils.Exists(nestedDir) {
		t.Fatalf("nested dir should be removed")
	}
}
