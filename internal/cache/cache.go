package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Cache struct {
	dir string
}

func New() (*Cache, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("cache: resolve user cache dir: %w", err)
	}
	dir := filepath.Join(base, "chorus")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cache: create: %w", err)
	}
	return &Cache{dir: dir}, nil
}

func Key(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(strings.ToLower(p))
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func (c *Cache) Get(key string) ([]byte, bool) {
	data, err := os.ReadFile(filepath.Join(c.dir, key))
	if err != nil {
		return nil, false
	}
	return data, true
}

func (c *Cache) Set(key string, data []byte) error {
	return os.WriteFile(filepath.Join(c.dir, key), data, 0o644)
}

func (c *Cache) Size() (int64, error) {
	var total int64
	err := filepath.Walk(c.dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

func (c *Cache) Clear() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		os.Remove(filepath.Join(c.dir, e.Name()))
	}
	return nil
}
