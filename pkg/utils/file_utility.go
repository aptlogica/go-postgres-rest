// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package utils

import (
	"fmt"
	"os"
)

func CreateFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer file.Close()
	return nil
}

func CreateDirRecursive(path string) error {
	return os.MkdirAll(path, 0777)
}

func DeleteFile(path string) error {
	return os.Remove(path)
}

func DeleteDirRecursive(path string) error {
	return os.RemoveAll(path)
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
