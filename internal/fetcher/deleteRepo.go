package fetcher

import (
	"os"
	"path/filepath"
)

func DeleteRepo(repoURL, dir string) error {
	dir = filepath.Join(dir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Директория не существует, нечего удалять
	}
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return nil
}