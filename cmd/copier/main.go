package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/cloudogu/dogu-data-seeder/internal/copy"
	"github.com/cloudogu/doguctl/registry"
	"log"
	"os"
	"strings"
)

const (
	defaultCesConfigBaseDir   = "/dogumount/etc/ces/config"
	defaultLocalConfigBaseDir = "/dogumount/var/ces/config"
)

var (
	copyCmd = flag.NewFlagSet("copy", flag.ExitOnError)
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("expected at least on of the following commands: \n"+
			"%s - copy files from specified volumes to destination paths", copyCmd.Name())
	}

	var err error
	switch os.Args[1] {
	case copyCmd.Name():
		err = handleCopyCommand(os.Args[2:], getCopier, getDoguConfig, getfileTracker)
	default:
		err = errors.New("unknown command")
	}

	if err != nil {
		log.Fatal(err.Error())
	}
}

func handleCopyCommand(args []string, volumeMountCopyGetter copierGetter, configGetter doguConfigGetter, fileTrackerGetter fileTrackerGetter) error {
	cesConfigBaseDir := copyCmd.String("cesConfigBaseDir", defaultCesConfigBaseDir, fmt.Sprintf("Defines the base dir for the dogu config - defaults to %s", defaultCesConfigBaseDir))
	localConfigBaseDir := copyCmd.String("localConfigBaseDir", defaultLocalConfigBaseDir, fmt.Sprintf("Defines the base dir for the local dogu config - defaults to %s", defaultLocalConfigBaseDir))

	var sourcePaths stringSliceFlag
	var targetPaths stringSliceFlag
	copyCmd.Var(&sourcePaths, "source", "")
	copyCmd.Var(&targetPaths, "target", "")
	err := copyCmd.Parse(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	doguConfigRegistry, err := configGetter(*cesConfigBaseDir, *localConfigBaseDir)
	if err != nil {
		return fmt.Errorf("failed to generate dogu file config with config dir %s and local config dir %s: %w", *cesConfigBaseDir, *localConfigBaseDir, err)
	}

	fileSystem := &copy.FileSystem{}
	fileTracker := fileTrackerGetter(doguConfigRegistry, fileSystem)
	log.Println("delete old tracked files")
	err = fileTracker.DeleteAllTrackedFiles()
	if err != nil {
		return err
	}

	volumeMountCopy := volumeMountCopyGetter(fileSystem, fileTracker)

	if len(sourcePaths) != len(targetPaths) {
		return fmt.Errorf("amount of source and target paths aren't equal")
	}

	if len(sourcePaths) == 0 && len(targetPaths) == 0 {
		log.Println("no source and target paths given")
		return nil
	}

	copyList := make([]copy.SrcAndDestination, 0, len(sourcePaths))
	for i := range sourcePaths {
		copyList = append(copyList, copy.SrcAndDestination{
			Src:  sourcePaths[i],
			Dest: targetPaths[i],
		})
	}

	err = volumeMountCopy.CopyVolumeMount(copyList)
	if err != nil {
		return err
	}

	return nil
}

type stringSliceFlag []string

func (i *stringSliceFlag) String() string {
	return strings.Join(*i, ",")
}

func (i *stringSliceFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type doguConfigGetter = func(cesConfigBaseDir, localConfigBaseDir string) (doguConfigReaderWriter, error)

func getDoguConfig(cesConfigBaseDir, localConfigBaseDir string) (doguConfigReaderWriter, error) {
	return registry.NewDoguFileConfigurationContext(cesConfigBaseDir, localConfigBaseDir)
}

type fileTrackerGetter = func(doguConfigRegistry doguConfigReaderWriter, filesystem filesystem) fileTracker

func getfileTracker(doguConfigRegistry doguConfigReaderWriter, filesystem filesystem) fileTracker {
	return copy.NewLocalConfigFileTracker(doguConfigRegistry, filesystem)
}

type copierGetter = func(filesystem filesystem, fileTracker fileTracker) volumeCopier

func getCopier(filesystem filesystem, fileTracker fileTracker) volumeCopier {
	return copy.NewVolumeMountCopier(filesystem, fileTracker)
}
