package fetcher

import (
	"fmt"
	"os"
	"github.com/go-git/go-git/v6"
	"path/filepath"
	"strings"
	"net/url"
	"path"
)

func nameRepo(repoURL string) string {
	repoURL = strings.TrimSpace(repoURL)
	repoURL = strings.TrimSuffix(repoURL, "/")

	var lastPart string

	if strings.HasPrefix(repoURL, "git@") || (strings.Contains(repoURL, ":") && !strings.Contains(repoURL, "://")) {
		if i := strings.LastIndex(repoURL, ":"); i != -1 {
			lastPart = repoURL[i+1:]
		} else {
			lastPart = repoURL
		}
	} else {
		if u, err := url.Parse(repoURL); err == nil && u.Path != "" {
			lastPart = path.Base(u.Path)
		} else {
			lastPart = path.Base(repoURL)
		}
	}

	lastPart = strings.TrimSuffix(lastPart, ".git")
	return lastPart
}


// CloneRepo clones a repository from the given URL to the specified directory.			
func CloneRepo(repoURL, dir string) error {
	
	dir = filepath.Join(dir, nameRepo(repoURL))
	path := filepath.Join(dir, ".git")

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		return fmt.Errorf("directory %s already exists", dir)
	}
	_, err := git.PlainClone(dir, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})

	
	if err := os.RemoveAll(path); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File not found:", path)
		} else {
			fmt.Println("Error removing:", err)
		}
		return err
	}

	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return err
}