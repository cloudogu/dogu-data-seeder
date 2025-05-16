package copy

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"testing"
	"time"
)

func TestVolumeMountCopier_CopyVolumeMount(t *testing.T) {
	t.Run("should return nil on empty parameter map", func(t *testing.T) {
		// given
		sut := VolumeMountCopier{}

		// when
		err := sut.CopyVolumeMount([]SrcAndDestination{})

		// then
		require.NoError(t, err)
	})

	t.Run("should return error on error resolving symlink", func(t *testing.T) {
		// given
		copies := []SrcAndDestination{
			{
				Src:  "/mount",
				Dest: "/custom/config",
			},
		}
		dataFileInfo := myFileInfo{
			mode:  os.ModeSymlink,
			isDir: true,
		}

		sut := VolumeMountCopier{}
		fileSystemMock := NewMockFilesystem(t)
		fileSystemMock.EXPECT().Lstat("/mount/..data").Return(dataFileInfo, nil)
		fileSystemMock.EXPECT().EvalSymlinks("/mount/..data").Return("", assert.AnError)

		sut.fileSystem = fileSystemMock

		// when
		err := sut.CopyVolumeMount(copies)

		// then
		require.Error(t, err)
	})

	t.Run("should handle file with subPath volume mount", func(t *testing.T) {
		// given
		copies := []SrcAndDestination{
			{
				Src:  "/mount",
				Dest: "/custom/config",
			},
		}

		sut := VolumeMountCopier{}
		fileSystemMock := NewMockFilesystem(t)
		fileSystemMock.EXPECT().Lstat("/mount/..data").Return(nil, assert.AnError)
		fileSystemMock.EXPECT().WalkDir("/mount", mock.AnythingOfType("fs.WalkDirFunc")).Return(nil)

		sut.fileSystem = fileSystemMock

		// when
		err := sut.CopyVolumeMount(copies)

		// then
		require.NoError(t, err)
	})

	t.Run("should handle symlinked folder separate", func(t *testing.T) {
		// given
		copies := []SrcAndDestination{
			{
				Src:  "/mount",
				Dest: "/custom/config",
			},
		}
		symLinkPath := "/mount/..data"
		// realFilePath := "/mount/..20250504/file"
		realDirPath := "/mount/..20250504"
		dataFileInfo := myFileInfo{
			mode:  os.ModeSymlink,
			isDir: true,
		}
		realDataDirInfo := myFileInfo{
			mode:  os.ModeDir,
			isDir: true,
		}

		sut := VolumeMountCopier{}
		fileSystemMock := NewMockFilesystem(t)
		fileSystemMock.EXPECT().Lstat(symLinkPath).Return(dataFileInfo, nil)
		fileSystemMock.EXPECT().EvalSymlinks(symLinkPath).Return(realDirPath, nil)
		fileSystemMock.EXPECT().Stat(realDirPath).Return(realDataDirInfo, nil)
		fileSystemMock.EXPECT().WalkDir(realDirPath, mock.AnythingOfType("fs.WalkDirFunc")).Return(nil)
		fileSystemMock.EXPECT().WalkDir("/mount", mock.AnythingOfType("fs.WalkDirFunc")).Return(nil)

		sut.fileSystem = fileSystemMock

		// when
		err := sut.CopyVolumeMount(copies)

		// then
		require.NoError(t, err)
	})
}

func TestCopier_resolveSymLinkChain(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// given
		link := "/tmp/mount/config"
		realPath := "/tmp/mount/..data/..20250705"
		fileInfo := &myFileInfo{isDir: true}

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().EvalSymlinks(link).Return(realPath, nil)
		filesystemMock.EXPECT().Stat(realPath).Return(fileInfo, nil)

		sut := &VolumeMountCopier{
			fileSystem: filesystemMock,
		}

		// when
		filePath, err := sut.resolveDataSymlink(link)

		// then
		require.NoError(t, err)
		assert.Equal(t, realPath, filePath)
	})

	t.Run("should return error on stat error", func(t *testing.T) {
		// given
		link := "/tmp/mount/config"
		realPath := "/tmp/mount/..data/..20250705"
		fileInfo := &myFileInfo{isDir: true}

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().EvalSymlinks(link).Return(realPath, nil)
		filesystemMock.EXPECT().Stat(realPath).Return(fileInfo, assert.AnError)

		sut := &VolumeMountCopier{
			fileSystem: filesystemMock,
		}

		// when
		_, err := sut.resolveDataSymlink(link)

		// then
		require.Error(t, err)
	})

	t.Run("should return error if data link does not point to a dir", func(t *testing.T) {
		// given
		link := "/tmp/mount/config"
		realPath := "/tmp/mount/..data/..20250705"
		fileInfo := &myFileInfo{isDir: false}

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().EvalSymlinks(link).Return(realPath, nil)
		filesystemMock.EXPECT().Stat(realPath).Return(fileInfo, nil)

		sut := &VolumeMountCopier{
			fileSystem: filesystemMock,
		}

		// when
		_, err := sut.resolveDataSymlink(link)

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "data symlink /tmp/mount/config should point to a dir")
	})
}

func TestCopier_walk(t *testing.T) {
	t.Run("should do nothing if source file is a dir", func(t *testing.T) {
		// given
		srcFile := "/tmp/mount/dir"
		srcFileInfo := &myFileInfo{mode: os.ModeDir, isDir: true}
		dirEntry := &myDirEntry{fileInfo: srcFileInfo}

		sut := &VolumeMountCopier{}

		// when
		err := sut.walk("", "", srcFile, false, dirEntry)

		// then
		require.NoError(t, err)
	})

	t.Run("should return error on getting info error", func(t *testing.T) {
		// given
		srcFile := "/tmp/mount/file"
		srcFileInfo := &myFileInfo{mode: os.ModePerm, isDir: false}
		dirEntry := &myDirEntry{fileInfo: srcFileInfo, infoErr: assert.AnError}

		sut := &VolumeMountCopier{}

		// when
		err := sut.walk("", "", srcFile, false, dirEntry)

		// then
		require.Error(t, err)
		assert.ErrorIs(t, assert.AnError, err)
	})

	t.Run("should return nil if the source file is not a regular file", func(t *testing.T) {
		// given
		srcFile := "/tmp/mount/file"
		srcFileInfo := &myFileInfo{mode: os.ModeSymlink, isDir: false}
		dirEntry := &myDirEntry{fileInfo: srcFileInfo}

		sut := &VolumeMountCopier{}

		// when
		err := sut.walk("", "", srcFile, false, dirEntry)

		// then
		require.NoError(t, err)
	})

	t.Run("should return error if the destination file exists but is not a regular file", func(t *testing.T) {
		// given
		srcVolume := "/tmp/mount"
		srcFile := "/tmp/mount/dir/file"
		destVolume := "/var/lib/custom"
		destFile := "/var/lib/custom/dir/file"
		srcFileInfo := &myFileInfo{mode: os.ModePerm, isDir: false}
		destFileInfo := &myFileInfo{mode: os.ModeSymlink}
		dirEntry := &myDirEntry{fileInfo: srcFileInfo}

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().Stat(destFile).Return(destFileInfo, nil)

		sut := &VolumeMountCopier{}
		sut.fileSystem = filesystemMock

		// when
		err := sut.walk(srcVolume, destVolume, srcFile, true, dirEntry)

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "destination file /var/lib/custom/dir/file exists and is not a regular file")
	})

	t.Run("should copy with regular src file in non existent destination", func(t *testing.T) {
		// given
		src := "/tmp/mount"
		dest := "/var/lib/custom"
		srcFile := "/tmp/mount/config"
		destFile := "/var/lib/custom/config"
		srcFileInfo := &myFileInfo{mode: os.ModePerm}
		destFileInfo := &myFileInfo{mode: os.ModePerm}
		dirEntry := &myDirEntry{fileInfo: srcFileInfo}

		filesystemMock := NewMockFilesystem(t)
		// return error to indicate that the srcFile is not existent in the destination
		filesystemMock.EXPECT().Stat("/var/lib/custom/config").Return(destFileInfo, assert.AnError)
		copyMock := NewMockCopier(t)
		copyMock.EXPECT().Execute(srcFile, destFile, filesystemMock).Return(nil)
		fileTrackerMock := newMockFileTracker(t)
		fileTrackerMock.EXPECT().AddFile("/var/lib/custom/config").Return(nil)

		sut := &VolumeMountCopier{}
		sut.fileSystem = filesystemMock
		sut.fileTracker = fileTrackerMock
		sut.copier = copyMock.Execute

		// when
		err := sut.walk(src, dest, srcFile, false, dirEntry)

		// then
		require.NoError(t, err)
	})

	t.Run("should overwrite existing file in destination", func(t *testing.T) {
		// given
		src := "/tmp/mount"
		dest := "/var/lib/custom"
		srcFile := "/tmp/mount/config"
		destFile := "/var/lib/custom/config"
		srcFileInfo := &myFileInfo{mode: os.ModePerm}
		destFileInfo := &myFileInfo{mode: os.ModePerm}
		dirEntry := &myDirEntry{fileInfo: srcFileInfo}

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().Stat("/var/lib/custom/config").Return(destFileInfo, nil)
		filesystemMock.EXPECT().SameFile(srcFileInfo, destFileInfo).Return(false)
		copyMock := NewMockCopier(t)
		copyMock.EXPECT().Execute(srcFile, destFile, filesystemMock).Return(nil)
		fileTrackerMock := newMockFileTracker(t)
		fileTrackerMock.EXPECT().AddFile("/var/lib/custom/config").Return(nil)

		sut := &VolumeMountCopier{}
		sut.fileSystem = filesystemMock
		sut.fileTracker = fileTrackerMock
		sut.copier = copyMock.Execute

		// when
		err := sut.walk(src, dest, srcFile, false, dirEntry)

		// then
		require.NoError(t, err)
	})

	t.Run("should do nothing if fileinfo is equal", func(t *testing.T) {
		// given
		src := "/tmp/mount"
		dest := "/var/lib/custom"
		srcFile := "/tmp/mount/config"
		srcFileInfo := &myFileInfo{mode: os.ModePerm}
		destFileInfo := srcFileInfo
		dirEntry := &myDirEntry{fileInfo: srcFileInfo}

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().Stat("/var/lib/custom/config").Return(destFileInfo, nil)
		filesystemMock.EXPECT().SameFile(srcFileInfo, destFileInfo).Return(true)
		copyMock := NewMockCopier(t)
		fileTrackerMock := newMockFileTracker(t)

		sut := &VolumeMountCopier{}
		sut.fileSystem = filesystemMock
		sut.fileTracker = fileTrackerMock
		sut.copier = copyMock.Execute

		// when
		err := sut.walk(src, dest, srcFile, false, dirEntry)

		// then
		require.NoError(t, err)
	})

	t.Run("should keep nested dirs from subpath mounts", func(t *testing.T) {
		// given
		src := "/tmp/mount"
		dest := "/var/lib/custom"
		srcFile := "/tmp/mount/dir1/dir2/config"
		destFile := "/var/lib/custom/dir1/dir2/config"
		srcFileInfo := &myFileInfo{mode: os.ModePerm}
		destFileInfo := &myFileInfo{mode: os.ModePerm}
		dirEntry := &myDirEntry{fileInfo: srcFileInfo}

		filesystemMock := NewMockFilesystem(t)
		// return error to indicate that the srcFile is not existent in the destination
		filesystemMock.EXPECT().Stat("/var/lib/custom/dir1/dir2/config").Return(destFileInfo, assert.AnError)
		copyMock := NewMockCopier(t)
		copyMock.EXPECT().Execute(srcFile, destFile, filesystemMock).Return(nil)
		fileTrackerMock := newMockFileTracker(t)
		fileTrackerMock.EXPECT().AddFile("/var/lib/custom/dir1/dir2/config").Return(nil)

		sut := &VolumeMountCopier{}
		sut.fileSystem = filesystemMock
		sut.fileTracker = fileTrackerMock
		sut.copier = copyMock.Execute

		// when
		err := sut.walk(src, dest, srcFile, true, dirEntry)

		// then
		require.NoError(t, err)
	})
}

type myDirEntry struct {
	fileInfo *myFileInfo
	infoErr  error
}

func (m myDirEntry) Name() string {
	// TODO implement me
	panic("implement me")
}

func (m myDirEntry) IsDir() bool {
	return m.fileInfo.IsDir()
}

func (m myDirEntry) Type() fs.FileMode {
	// TODO implement me
	panic("implement me")
}

func (m myDirEntry) Info() (fs.FileInfo, error) {
	return m.fileInfo, m.infoErr
}

type myFileInfo struct {
	isDir bool
	mode  os.FileMode
}

func (m myFileInfo) Name() string {
	// TODO implement me
	panic("implement me")
}

func (m myFileInfo) Size() int64 {
	// TODO implement me
	panic("implement me")
}

func (m myFileInfo) Mode() fs.FileMode {
	return m.mode
}

func (m myFileInfo) ModTime() time.Time {
	// TODO implement me
	panic("implement me")
}

func (m myFileInfo) IsDir() bool {
	return m.isDir
}

func (m myFileInfo) Sys() any {
	// TODO implement me
	panic("implement me")
}
