// Package vfs is an abstraction of various filesystem operations.
package vfs

import (
	"github.com/pkg/errors"
	"io"
	"os"
)

var (
	ErrLocked           = errors.New("already locked")
	ErrAlreadyCommitted = errors.New("already committed")
)

type InputFile interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

type OutputFile interface {
	io.Writer
	io.Closer
	Sync() error
}

type AtomicOutputFile interface {
	OutputFile
	Commit() error
}

// FileSystem
type FileSystem interface {

	// Lock acquires an exclusive lock on a file and returns a Closer
	// that should be used for releasing the lock.
	Lock(name string) (io.Closer, error)

	// Close releases any resources associated with the filesystem.
	Close() error

	// Path returns the absolute path to the root of the filesystem.
	Path() string

	// ReadDir reads the root directory and returns a array of directory entries sorted by name.
	ReadDir() ([]os.FileInfo, error)

	// OpenFile opens the named file for reading. Returns an error if the file does not exist.
	OpenFile(name string) (InputFile, error)

	// CreateFile opens the named file for writing. If the files does not exist, it will be created.
	// If it does exist and overwrite is set to true, it will be truncated. Otherwise an error is returned.
	CreateFile(name string, overwrite bool) (OutputFile, error)

	// CreateAtomicFile opens the named file for writing. However, the file is not visible under
	// the target name until explicitly committed.
	CreateAtomicFile(name string) (AtomicOutputFile, error)

	// Rename moves oldname to newname. If newname already exists, it is replaced.
	Rename(newname, oldname string) error

	// Remove removes the named file.
	Remove(name string) error
}

func IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func IsExist(err error) bool {
	return os.IsExist(err)
}

func IsLocked(err error) bool {
	return errors.Cause(err) == ErrLocked
}
