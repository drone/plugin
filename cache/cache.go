package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/lockedfile"
)

func Add(key string, addItem func() error) error {
	if err := os.MkdirAll(key, 0700); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create directory %s", key))
	}

	lockFpath := filepath.Join(key, ".lock")
	lock, err := lockedfile.Create(lockFpath)
	if err != nil {
		return errors.Wrap(err, "failed to take file lock")
	}
	defer lock.Close()

	// If data is already present, return
	if _, err := os.Stat(filepath.Join(key, ".success")); err == nil {
		return nil
	}

	if err := addItem(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to add item: %s to cache", key))
	}

	integrityFpath := filepath.Join(key, ".success")
	f, err := os.Create(integrityFpath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create integrity file: %s", integrityFpath))
	}
	f.Close()

	return nil
}

func GetKeyName(name string) string {
	return filepath.Join(getCacheDir(), sha(name))
}

func getCacheDir() string {
	dir, _ := os.UserHomeDir()
	return filepath.Join(dir, ".cache")
}

func sha(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
