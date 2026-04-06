package repo

import (
	"os"
	"path/filepath"
)

func DetectRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if root, ok := findGitRoot(cwd); ok {
		return root, nil
	}

	return cwd, nil
}

func findGitRoot(start string) (string, bool) {
	dir := start
	for {
		if isDir(filepath.Join(dir, ".git")) {
			return dir, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
