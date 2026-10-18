package store

import (
	"os"
	"time"
)

// tempFile is the part of *os.File that writeAtomic uses, so tests can fail writes.
type tempFile interface {
	Name() string
	Write(p []byte) (int, error)
	Sync() error
	Close() error
}

// fileOps holds the filesystem calls the store makes. New fills it with os
// functions; tests in this package replace fields to inject failures.
type fileOps struct {
	readFile   func(name string) ([]byte, error)
	createTemp func(dir, pattern string) (tempFile, error)
	rename     func(oldpath, newpath string) error
	remove     func(name string) error
	stat       func(name string) (os.FileInfo, error)
	mkdirAll   func(path string, perm os.FileMode) error
	readDir    func(name string) ([]os.DirEntry, error)
	sleep      func(d time.Duration)
}

func osFileOps() fileOps {
	return fileOps{
		readFile: os.ReadFile,
		createTemp: func(dir, pattern string) (tempFile, error) {
			return os.CreateTemp(dir, pattern)
		},
		rename:   os.Rename,
		remove:   os.Remove,
		stat:     os.Stat,
		mkdirAll: os.MkdirAll,
		readDir:  os.ReadDir,
		sleep:    time.Sleep,
	}
}
