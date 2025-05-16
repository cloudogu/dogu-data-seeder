package main

import (
	"flag"
	"github.com/cloudogu/dogu-data-seeder/internal/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_handleCopyCommand(t *testing.T) {
	t.Run("should call copy subsequent src to destination volumes", func(t *testing.T) {
		// given
		copyCmd = flag.NewFlagSet("copy", flag.ExitOnError)
		args := []string{"--source=/src1", "--target=/target1", "--source=/src2", "--target=/target2"}
		expectedCopyList := []copy.SrcAndDestination{
			{Src: "/src1", Dest: "/target1"}, {Src: "/src2", Dest: "/target2"},
		}

		getter := func(filesystem filesystem, fileTracker fileTracker) volumeCopier {
			copier := newMockVolumeCopier(t)
			copier.EXPECT().CopyVolumeMount(expectedCopyList).Return(nil)
			return copier
		}

		configGetter := func(cesConfigBaseDir, localConfigBaseDir string) (doguConfigReaderWriter, error) {
			return newMockDoguConfigReaderWriter(t), nil
		}
		trackerGetter := func(doguConfigRegistry doguConfigReaderWriter, filesystem filesystem) fileTracker {
			tracker := newMockFileTracker(t)
			tracker.EXPECT().DeleteAllTrackedFiles().Return(nil)
			return tracker
		}

		// when
		err := handleCopyCommand(args, getter, configGetter, trackerGetter)

		// then
		require.NoError(t, err)
	})

	t.Run("should return error on copy error error", func(t *testing.T) {
		// given
		copyCmd = flag.NewFlagSet("copy", flag.ExitOnError)
		args := []string{"--source=/src1", "--target=/target1", "--source=/src2", "--target=/target2"}
		expectedCopyList := []copy.SrcAndDestination{
			{Src: "/src1", Dest: "/target1"}, {Src: "/src2", Dest: "/target2"},
		}
		getter := func(filesystem filesystem, fileTracker fileTracker) volumeCopier {
			copier := newMockVolumeCopier(t)
			copier.EXPECT().CopyVolumeMount(expectedCopyList).Return(assert.AnError)
			return copier
		}

		configGetter := func(cesConfigBaseDir, localConfigBaseDir string) (doguConfigReaderWriter, error) {
			return newMockDoguConfigReaderWriter(t), nil
		}
		trackerGetter := func(doguConfigRegistry doguConfigReaderWriter, filesystem filesystem) fileTracker {
			tracker := newMockFileTracker(t)
			tracker.EXPECT().DeleteAllTrackedFiles().Return(nil)
			return tracker
		}

		// when
		err := handleCopyCommand(args, getter, configGetter, trackerGetter)

		// then
		require.Error(t, err)
	})

	t.Run("should return nil and delete tracked files on empty parameter", func(t *testing.T) {
		// given
		copyCmd = flag.NewFlagSet("copy", flag.ExitOnError)
		var args []string

		getter := func(filesystem filesystem, fileTracker fileTracker) volumeCopier {
			copier := newMockVolumeCopier(t)
			return copier
		}

		configGetter := func(cesConfigBaseDir, localConfigBaseDir string) (doguConfigReaderWriter, error) {
			return newMockDoguConfigReaderWriter(t), nil
		}
		trackerGetter := func(doguConfigRegistry doguConfigReaderWriter, filesystem filesystem) fileTracker {
			tracker := newMockFileTracker(t)
			tracker.EXPECT().DeleteAllTrackedFiles().Return(nil)
			return tracker
		}

		// when
		err := handleCopyCommand(args, getter, configGetter, trackerGetter)

		// then
		require.NoError(t, err)
	})

	t.Run("should return error on odd parameters", func(t *testing.T) {
		// given
		copyCmd = flag.NewFlagSet("copy", flag.ExitOnError)
		args := []string{"--source=/src1", "--target=/target1", "--source=/src2"}

		getter := func(filesystem filesystem, fileTracker fileTracker) volumeCopier {
			copier := newMockVolumeCopier(t)
			return copier
		}

		configGetter := func(cesConfigBaseDir, localConfigBaseDir string) (doguConfigReaderWriter, error) {
			return newMockDoguConfigReaderWriter(t), nil
		}
		trackerGetter := func(doguConfigRegistry doguConfigReaderWriter, filesystem filesystem) fileTracker {
			tracker := newMockFileTracker(t)
			tracker.EXPECT().DeleteAllTrackedFiles().Return(nil)
			return tracker
		}

		// when
		err := handleCopyCommand(args, getter, configGetter, trackerGetter)

		// then
		require.Error(t, err)
		assert.ErrorContains(t, err, "amount of source and target paths aren't equal")
	})
}
