package utils

import "os"

func CreateFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
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
