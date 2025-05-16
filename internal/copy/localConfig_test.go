package copy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocalConfigFileTracker_DeleteAllTrackedFiles(t1 *testing.T) {
	keyAdditionalMounts := "additionalMounts"
	yamlFiles := `
  - /path/database
  - /path/config
`

	type fields struct {
		doguConfig func(t *testing.T) doguConfigReaderWriter
		fileSystem func(t *testing.T) Filesystem
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should delete all files from local config and reset those after that",
			fields: fields{
				doguConfig: func(t *testing.T) doguConfigReaderWriter {
					doguConfigMock := newMockDoguConfigReaderWriter(t)
					doguConfigMock.EXPECT().Exists(keyAdditionalMounts).Return(true, nil)
					doguConfigMock.EXPECT().Get(keyAdditionalMounts).Return(yamlFiles, nil)
					doguConfigMock.EXPECT().Set(keyAdditionalMounts, "").Return(nil)

					return doguConfigMock
				},
				fileSystem: func(t *testing.T) Filesystem {
					filesystemMock := NewMockFilesystem(t)
					filesystemMock.EXPECT().DeleteFile("/path/database").Return(nil)
					filesystemMock.EXPECT().DeleteFile("/path/config").Return(nil)
					return filesystemMock
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error on error resetting the config",
			fields: fields{
				doguConfig: func(t *testing.T) doguConfigReaderWriter {
					doguConfigMock := newMockDoguConfigReaderWriter(t)
					doguConfigMock.EXPECT().Exists(keyAdditionalMounts).Return(true, nil)
					doguConfigMock.EXPECT().Get(keyAdditionalMounts).Return(yamlFiles, nil)
					doguConfigMock.EXPECT().Set(keyAdditionalMounts, "").Return(assert.AnError)

					return doguConfigMock
				},
				fileSystem: func(t *testing.T) Filesystem {
					filesystemMock := NewMockFilesystem(t)
					filesystemMock.EXPECT().DeleteFile("/path/database").Return(nil)
					filesystemMock.EXPECT().DeleteFile("/path/config").Return(nil)
					return filesystemMock
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "failed to reset local config key")
				return true
			},
		},
		{
			name: "should return error on error deleting a file",
			fields: fields{
				doguConfig: func(t *testing.T) doguConfigReaderWriter {
					doguConfigMock := newMockDoguConfigReaderWriter(t)
					doguConfigMock.EXPECT().Exists(keyAdditionalMounts).Return(true, nil)
					doguConfigMock.EXPECT().Get(keyAdditionalMounts).Return(yamlFiles, nil)
					doguConfigMock.EXPECT().Set(keyAdditionalMounts, "").Return(nil)

					return doguConfigMock
				},
				fileSystem: func(t *testing.T) Filesystem {
					filesystemMock := NewMockFilesystem(t)
					filesystemMock.EXPECT().DeleteFile("/path/database").Return(nil)
					filesystemMock.EXPECT().DeleteFile("/path/config").Return(assert.AnError)
					return filesystemMock
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.ErrorIs(t, err, assert.AnError)
				return true
			},
		},
		{
			name: "should return error on error getting config",
			fields: fields{
				doguConfig: func(t *testing.T) doguConfigReaderWriter {
					doguConfigMock := newMockDoguConfigReaderWriter(t)
					doguConfigMock.EXPECT().Exists(keyAdditionalMounts).Return(true, assert.AnError)

					return doguConfigMock
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.ErrorIs(t, err, assert.AnError)
				return true
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t *testing.T) {
			var doguConfig doguConfigReaderWriter
			if tt.fields.doguConfig != nil {
				doguConfig = tt.fields.doguConfig(t)
			}
			var filesystem Filesystem
			if tt.fields.fileSystem != nil {
				filesystem = tt.fields.fileSystem(t)
			}

			sut := &LocalConfigFileTracker{
				doguConfig: doguConfig,
				fileSystem: filesystem,
			}
			tt.wantErr(t, sut.DeleteAllTrackedFiles(), fmt.Sprintf("DeleteAllTrackedFiles()"))
		})
	}
}

func TestLocalConfigFileTracker_AddFile(t1 *testing.T) {
	keyAdditionalMounts := "additionalMounts"
	actualYamlFiles := `
  - /path/database
  - /path/config
`
	expectedYamlFiles := `- /path/database
- /path/config
- /path/new
`

	type fields struct {
		doguConfig func(t *testing.T) doguConfigReaderWriter
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should add file in config",
			fields: fields{
				doguConfig: func(t *testing.T) doguConfigReaderWriter {
					doguConfigMock := newMockDoguConfigReaderWriter(t)
					doguConfigMock.EXPECT().Exists(keyAdditionalMounts).Return(true, nil)
					doguConfigMock.EXPECT().Get(keyAdditionalMounts).Return(actualYamlFiles, nil)
					doguConfigMock.EXPECT().Set(keyAdditionalMounts, expectedYamlFiles).Return(nil)

					return doguConfigMock
				},
			},
			args: args{
				path: "/path/new",
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error on error setting config",
			fields: fields{
				doguConfig: func(t *testing.T) doguConfigReaderWriter {
					doguConfigMock := newMockDoguConfigReaderWriter(t)
					doguConfigMock.EXPECT().Exists(keyAdditionalMounts).Return(true, nil)
					doguConfigMock.EXPECT().Get(keyAdditionalMounts).Return(actualYamlFiles, nil)
					doguConfigMock.EXPECT().Set(keyAdditionalMounts, expectedYamlFiles).Return(assert.AnError)

					return doguConfigMock
				},
			},
			args: args{
				path: "/path/new",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.ErrorIs(t, err, assert.AnError)
				assert.ErrorContains(t, err, "failed to set value - /path/database\n- /path/config\n- /path/new\n to key [/path/database /path/config /path/new]")
				return true
			},
		},
		{
			name: "should return error on error getting additional mounts",
			fields: fields{
				doguConfig: func(t *testing.T) doguConfigReaderWriter {
					doguConfigMock := newMockDoguConfigReaderWriter(t)
					doguConfigMock.EXPECT().Exists(keyAdditionalMounts).Return(false, assert.AnError)

					return doguConfigMock
				},
			},
			args: args{
				path: "/path/new",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.ErrorIs(t, err, assert.AnError)
				return true
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t *testing.T) {
			var doguConfig doguConfigReaderWriter
			if tt.fields.doguConfig != nil {
				doguConfig = tt.fields.doguConfig(t)
			}
			sut := &LocalConfigFileTracker{
				doguConfig: doguConfig,
			}

			tt.wantErr(t, sut.AddFile(tt.args.path), fmt.Sprintf("AddFile(%v)", tt.args.path))
		})
	}
}
