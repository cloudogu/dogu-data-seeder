package copy

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestCopier_copyFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// given
		src := "/mount/source"
		dest := "/dir/destination"
		srcFile := &os.File{}
		destFile := &os.File{}

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().Open(src).Return(srcFile, nil)
		filesystemMock.EXPECT().MkdirAll("/dir", os.FileMode(0770)).Return(nil)
		filesystemMock.EXPECT().Create(dest).Return(destFile, nil)
		filesystemMock.EXPECT().Copy(destFile, srcFile).Return(0, nil)
		filesystemMock.EXPECT().SyncFile(destFile).Return(nil)
		filesystemMock.EXPECT().CloseFile(srcFile).Return(nil)
		filesystemMock.EXPECT().CloseFile(destFile).Return(nil)

		// when
		err := copyFile(src, dest, filesystemMock)

		// then
		require.NoError(t, err)
	})

	t.Run("should return error on open source file error", func(t *testing.T) {
		// given
		src := "/mount/source"

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().Open(src).Return(nil, assert.AnError)

		// when
		err := copyFile(src, "", filesystemMock)

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to open file /mount/source")
	})

	t.Run("should return error on create subdir error", func(t *testing.T) {
		// given
		src := "/mount/source"
		dest := "/dir/destination"
		srcFile := &os.File{}

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().Open(src).Return(srcFile, nil)
		filesystemMock.EXPECT().CloseFile(srcFile).Return(nil)
		filesystemMock.EXPECT().MkdirAll("/dir", os.FileMode(0770)).Return(assert.AnError)

		// when
		err := copyFile(src, dest, filesystemMock)

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to create dirs for path /dir/destination")
	})

	t.Run("should return error on error creating destination file", func(t *testing.T) {
		// given
		src := "/mount/source"
		dest := "/dir/destination"
		srcFile := &os.File{}
		destFile := &os.File{}

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().Open(src).Return(srcFile, nil)
		filesystemMock.EXPECT().CloseFile(srcFile).Return(nil)
		filesystemMock.EXPECT().MkdirAll("/dir", os.FileMode(0770)).Return(nil)
		filesystemMock.EXPECT().Create(dest).Return(destFile, assert.AnError)

		// when
		err := copyFile(src, dest, filesystemMock)

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to open file /dir/destination")
	})

	t.Run("should return error on error copy file", func(t *testing.T) {
		// given
		src := "/mount/source"
		dest := "/dir/destination"
		srcFile := &os.File{}
		destFile := &os.File{}

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().Open(src).Return(srcFile, nil)
		filesystemMock.EXPECT().MkdirAll("/dir", os.FileMode(0770)).Return(nil)
		filesystemMock.EXPECT().Create(dest).Return(destFile, nil)
		filesystemMock.EXPECT().CloseFile(srcFile).Return(nil)
		filesystemMock.EXPECT().CloseFile(destFile).Return(nil)
		filesystemMock.EXPECT().Copy(destFile, srcFile).Return(0, assert.AnError)

		// when
		err := copyFile(src, dest, filesystemMock)

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to copy from /mount/source to /dir/destination")
	})

	t.Run("should return error on syncing file", func(t *testing.T) {
		// given
		src := "/mount/source"
		dest := "/dir/destination"
		srcFile := &os.File{}
		destFile := &os.File{}

		filesystemMock := NewMockFilesystem(t)
		filesystemMock.EXPECT().Open(src).Return(srcFile, nil)
		filesystemMock.EXPECT().MkdirAll("/dir", os.FileMode(0770)).Return(nil)
		filesystemMock.EXPECT().Create(dest).Return(destFile, nil)
		filesystemMock.EXPECT().CloseFile(srcFile).Return(nil)
		filesystemMock.EXPECT().CloseFile(destFile).Return(nil)
		filesystemMock.EXPECT().Copy(destFile, srcFile).Return(0, nil)
		filesystemMock.EXPECT().SyncFile(destFile).Return(assert.AnError)

		// when
		err := copyFile(src, dest, filesystemMock)

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to flush buffer to file /dir/destination")
	})
}
