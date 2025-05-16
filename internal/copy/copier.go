package copy

import (
	"fmt"
	"log"
	"path"
)

func copyFile(srcfilePath, destFilePath string, fileSystem Filesystem) error {
	from, err := fileSystem.Open(srcfilePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", srcfilePath, err)
	}

	defer func() {
		closeErr := fileSystem.CloseFile(from)
		if closeErr != nil {
			log.Println("failed to close fd: %w", closeErr)
		}
	}()

	err = fileSystem.MkdirAll(path.Dir(destFilePath), 0770)
	if err != nil {
		return fmt.Errorf("failed to create dirs for path %s: %w", destFilePath, err)
	}

	to, err := fileSystem.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", destFilePath, err)
	}

	defer func() {
		closeErr := fileSystem.CloseFile(to)
		if closeErr != nil {
			log.Println(fmt.Errorf("failed to close fd: %w", closeErr))
		}
	}()

	_, err = fileSystem.Copy(to, from)
	if err != nil {
		return fmt.Errorf("failed to copy from %s to %s: %w", srcfilePath, destFilePath, err)
	}

	err = fileSystem.SyncFile(to)
	if err != nil {
		return fmt.Errorf("failed to flush buffer to file %s: %w", destFilePath, err)
	}

	log.Printf("Copied file %s to %s", srcfilePath, destFilePath)

	return nil
}
