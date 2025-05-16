package copy

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Filesystem interface {
	Lstat(path string) (os.FileInfo, error)
	EvalSymlinks(path string) (string, error)
	Stat(name string) (os.FileInfo, error)
	Open(name string) (*os.File, error)
	MkdirAll(path string, perm os.FileMode) error
	Create(name string) (*os.File, error)
	Copy(dst io.Writer, src io.Reader) (written int64, err error)
	CloseFile(file *os.File) error
	SyncFile(file *os.File) error
	SameFile(fi1, fi2 os.FileInfo) bool
	WalkDir(root string, fn fs.WalkDirFunc) error
	DeleteFile(path string) error
}

type FileSystem struct{}

func (f FileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}

func (f FileSystem) SameFile(fi1, fi2 os.FileInfo) bool {
	return os.SameFile(fi1, fi2)
}

func (f FileSystem) SyncFile(file *os.File) error {
	return file.Sync()
}

func (f FileSystem) CloseFile(file *os.File) error {
	return file.Close()
}

func (f FileSystem) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

func (f FileSystem) EvalSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

func (f FileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (f FileSystem) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (f FileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f FileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (f FileSystem) Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	return io.Copy(dst, src)
}

func (f FileSystem) DeleteFile(path string) error {
	if _, err := f.Stat(path); err == nil {
		return os.Remove(path)
	}

	return nil
}
