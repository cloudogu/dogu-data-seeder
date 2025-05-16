package main

import "github.com/cloudogu/dogu-data-seeder/internal/copy"

type volumeCopier interface {
	CopyVolumeMount(srcToDest []copy.SrcAndDestination) error
}

type filesystem interface {
	copy.Filesystem
}

type fileTracker interface {
	AddFile(path string) error
	DeleteAllTrackedFiles() error
}

type doguConfigReaderWriter interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Exists(key string) (bool, error)
}
