package copy

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type SrcAndDestination struct {
	Src  string
	Dest string
}

type Copier func(src, dest string, filesystem Filesystem) error

type fileTracker interface {
	AddFile(path string) error
}

type VolumeMountCopier struct {
	fileSystem  Filesystem
	copier      Copier
	fileTracker fileTracker
}

func NewVolumeMountCopier(fileSystem Filesystem, fileTracker fileTracker) *VolumeMountCopier {
	return &VolumeMountCopier{fileSystem, copyFile, fileTracker}
}

// CopyVolumeMount copies all files from the given src path in srcToDest parameter to the associate destination path.
// It only handles regular files.
// Existing files will be overwritten.
// If the volume was mounted without the subPath attribute, it resolves the data symlink and copies the real files
// from the mount. In such cases, it is possible that there are also subPath volume mounts in the directory.
// Therefore, this method will walk through the dir behind the symlink and the root of the mount.
// In the second run the symlinks will be ignored.
// If only the subPath attribute was used, it just copies all regular files to the destination.
func (v *VolumeMountCopier) CopyVolumeMount(srcToDest []SrcAndDestination) error {
	var multiErr []error

	for _, obj := range srcToDest {
		src := obj.Src
		dest := obj.Dest
		log.Printf("Start copy files from dir %s to %s", src, dest)
		data := filepath.Join(src, "..data")
		log.Printf("Checking data symlink %s", data)
		dataFileInfo, err := v.fileSystem.Lstat(data)

		if err == nil && dataFileInfo.Mode()&os.ModeSymlink != 0 {
			log.Println("Detected data symlink")
			// this volume was mounted without a subPath and all regular files are actually behind symlinks
			// e.g. src/..2025_05_07_4643786234
			var symErr error
			realDir, symErr := v.resolveDataSymlink(data)
			if symErr != nil {
				return fmt.Errorf("failed to resolve data dir symlink %s: %w", data, err)
			}

			multiErr = append(multiErr, v.walkDir(realDir, dest, false))
		}

		// Copy all files mounted as subpaths
		multiErr = append(multiErr, v.walkDir(src, dest, true))
	}
	return errors.Join(multiErr...)
}

func (v *VolumeMountCopier) walkDir(src, dest string, copySubPathMounts bool) error {
	var multiErr []error

	err := v.fileSystem.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			multiErr = append(multiErr, fmt.Errorf("error during filepath walk for path %s: %w", path, err))
			return nil
		}

		// If We want to copy real files mounted from subpaths, ignore potential mount with symlink structure.
		if copySubPathMounts && d.IsDir() && (strings.HasPrefix(d.Name(), "..") || d.Name() == "..data") {
			return fs.SkipDir
		}

		multiErr = append(multiErr, v.walk(src, dest, path, copySubPathMounts, d))
		return nil
	})

	if err != nil {
		multiErr = append(multiErr, err)
	}

	return errors.Join(multiErr...)
}

// walk will be executed on every path in src by [CopyVolumeMount].
// It only copies regular files and if symLinkChain is not empty it will remove this path from the base path.
// This is needed in volumeMounts from configmaps and secrets without the subPath attributes. In this case
// the files are behind symlinks and the resolved folder is used as source. This path from src to the resolved folder
// should not be copied to the destination.
func (v *VolumeMountCopier) walk(srcVolume, destVolume, filePath string, isSubPathMount bool, d fs.DirEntry) error {
	log.Printf("Processing file %s", filePath)
	if d.IsDir() {
		log.Printf("Skip dir %s", filePath)
		return nil
	}

	sourceFileInfo, err := d.Info()
	if err != nil {
		return err
	}

	if !sourceFileInfo.Mode().IsRegular() {
		log.Printf("skip source file %s because it is not a regular file", filePath)
		return nil
	}

	var rel string
	if isSubPathMount {
		rel, err = filepath.Rel(srcVolume, filePath)
		if err != nil {
			return fmt.Errorf("can't get the relative path of the source file %s and the source volume %s: %w", filePath, srcVolume, err)
		}
	} else {
		// There can't be nested folders in the mount. Just use the file name from example /mount/..20250504/filename
		// to determine destination path.
		_, rel = path.Split(filePath)
	}

	destinationFilePath := path.Join(destVolume, rel)
	destFileInfo, err := v.fileSystem.Stat(destinationFilePath)
	if err == nil {
		if !destFileInfo.Mode().IsRegular() {
			return fmt.Errorf("destination file %s exists and is not a regular file", destinationFilePath)
		}

		if v.fileSystem.SameFile(sourceFileInfo, destFileInfo) {
			log.Printf("source file %s and destination file %s are equal", filePath, destinationFilePath)
			return nil
		}
	}

	err = v.copier(filePath, destinationFilePath, v.fileSystem)
	if err != nil {
		return err
	}

	err = v.fileTracker.AddFile(destinationFilePath)
	if err != nil {
		return err
	}

	return nil
}

// resolveDataSymlink follows the symlink and returns the path from the real file and the relative to the dir of the symlink
func (v *VolumeMountCopier) resolveDataSymlink(symlink string) (string, error) {
	resolvedDataLink, err := v.fileSystem.EvalSymlinks(symlink)
	if err != nil {
		return "", err
	}

	dirInfo, err := v.fileSystem.Stat(resolvedDataLink)
	if err != nil {
		return "", err
	}

	if !dirInfo.IsDir() {
		return "", fmt.Errorf("data symlink %s should point to a dir", symlink)
	}

	return resolvedDataLink, nil
}
