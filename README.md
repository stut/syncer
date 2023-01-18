# Syncer

This project is designed to sit alongside a main task in a Nomad job group (for my purposes).

It will clone a git repo into the destination and then update it according to the configured interval.

It copes with the clone already existing, and will bail out if the directory is not empty but it's not an existing clone of the source.

## Configuration

All configuration is done using environment variables.

### `SYNCER_SOURCE`

The `git@` URL of the repo to clone.

Default: _none, must be specified_

### `SYNCER_DEST`

The destination directory. This will be created if it doesn't exist.

Default: _none, must be specified_

### `SYNCER_UPDATE_INTERVAL`

The update interval specified in [Go's time.Duration format](https://pkg.go.dev/time#ParseDuration), e.g. "1h", "30s", "12h", etc.

Default: `"1h"`

### `SYNCER_GIT_UPSTREAM`

The name of the upstream from which to pull.

Default: `"origin"`

### `SYNCER_SSH_KEY_FILENAME`

Private SSH key to use when accessing the remote git repository.

### `SYNCER_SSH_KEY_PASSWORD`

The SSH key password.

