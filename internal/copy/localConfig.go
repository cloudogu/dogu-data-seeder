package copy

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"slices"
)

const (
	additionalMountsConfigKey = "additionalMounts"
)

type doguConfigReaderWriter interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Exists(key string) (bool, error)
}

type LocalConfigFileTracker struct {
	doguConfig doguConfigReaderWriter
	fileSystem Filesystem
}

type PathSlice []string

func NewLocalConfigFileTracker(doguConfig doguConfigReaderWriter, system Filesystem) *LocalConfigFileTracker {
	return &LocalConfigFileTracker{doguConfig: doguConfig, fileSystem: system}
}

func (t *LocalConfigFileTracker) DeleteAllTrackedFiles() error {
	additionalMounts, err := t.getAdditionalMounts()
	if err != nil {
		return err
	}

	var multiErr []error
	for _, path := range additionalMounts {
		multiErr = append(multiErr, t.fileSystem.DeleteFile(path))
	}

	// TODO Should we reset the list only if multiErr is nil?
	err = t.doguConfig.Set(additionalMountsConfigKey, "")
	if err != nil {
		return fmt.Errorf("failed to reset local config key %s: %w", additionalMounts, err)
	}

	return errors.Join(multiErr...)
}

func (t *LocalConfigFileTracker) getAdditionalMounts() ([]string, error) {
	exists, err := t.doguConfig.Exists(additionalMountsConfigKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check if local config key %s exists: %w", additionalMountsConfigKey, err)
	}

	if !exists {
		return []string{}, nil
	}

	get, err := t.doguConfig.Get(additionalMountsConfigKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get local config key %s: %w", additionalMountsConfigKey, err)
	}

	paths := []string{}
	err = yaml.Unmarshal([]byte(get), &paths)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal local config key value %s from key %s: %w", get, additionalMountsConfigKey, err)
	}

	return paths, nil
}

func (t *LocalConfigFileTracker) AddFile(path string) error {
	additionalMounts, err := t.getAdditionalMounts()
	if err != nil {
		return err
	}

	if !slices.Contains(additionalMounts, path) {
		additionalMounts = append(additionalMounts, path)
	}

	out, err := yaml.Marshal(additionalMounts)
	if err != nil {
		return fmt.Errorf("failed to marshal additionalMounts %s to yaml: %w", additionalMounts, err)
	}

	value := string(out)
	err = t.doguConfig.Set(additionalMountsConfigKey, value)
	if err != nil {
		return fmt.Errorf("failed to set value %s to key %s: %w", value, additionalMounts, err)
	}

	return nil
}
