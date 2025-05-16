# Usage

This application currently supports these following commands:

- copy

## Copy

Copy copies all files from given source paths to destination paths.
The files have to be regular files and already existing files in the target will be overwritten.
An error during execution does not stop the whole process and does not remove previous copied files!

After every copy the destination file path will be tracked in a configuration file.
In real environments the local dogu config will be used.
At every start, the application deletes all files defined in the config to ensure data consistency.


### Example (local)

> You have to create the config files `normal/config.yaml` and `sensitive/config.yaml` in cesConfigBaseDir.

`target/dogu-data-seeder copy --cesConfigBaseDir=. --localConfigBaseDir=. --source=./cmd --target=./cmdCopy --source=./build --target=./buildCopy`

Where each source will be copied to the immediately following target.
