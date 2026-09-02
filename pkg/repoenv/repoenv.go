package repoenv

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const marker = "start-all.sh"

// FindRepoRoot walks up from start until it finds start-all.sh.
func FindRepoRoot(start string) (string, bool) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// Load reads the repository-root .env (next to start-all.sh). Nested .env files are ignored.
// Existing process environment variables are not overwritten.
func Load() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return LoadFrom(wd)
}

// LoadFrom is Load starting from an explicit directory.
func LoadFrom(start string) error {
	root, ok := FindRepoRoot(start)
	if !ok {
		return nil
	}
	path := filepath.Join(root, ".env")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return godotenv.Load(path)
}
