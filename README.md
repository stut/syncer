# Syncer

This project is designed to sit alongside a main task in a Nomad job group (for my purposes).

It will clone a git repo into the destination and then update it according to the configured interval.

It copes with the clone already existing, and will bail out if the directory is not empty and it's not an existing clone
of the configured source. When pulling it will check for changes and will perform a hard reset if changes exist. That
can be disabled through the configuration (see below) in which case the process will exit if changes are present.

## Configuration

All configuration is done using environment variables.

### `SYNCER_SOURCE`

The `git@` or `https://` URL of the repo to clone.

Default: _none, must be specified_

### `SYNCER_DEST`

The destination directory. This will be created if it doesn't exist.

Default: _none, must be specified_

### `SYNCER_UPDATE_INTERVAL`

The update interval specified in [Go's time.Duration format](https://pkg.go.dev/time#ParseDuration), e.g. "1h", "30s", "12h", etc.

Default: `"1h"`

### `SYNCER_GIT_BRANCH` or `SYNCER_GIT_TAG`

The name of the branch or tag to clone. If both are specified the tag will be cloned.

Default: `main`

### `SYNCER_GIT_UPSTREAM`

The name of the upstream from which to pull.

Default: `"origin"`

### `SYNCER_GIT_RESET_ON_CHANGES`

Whether to perform a hard reset if uncommitted changes exist in the repo when performing an update. The value should be
`true`, `yes`, `on`, or `1`; all other values will be interpreted as `false`.

Default: `true`

### `SYNCER_SSH_KEY_FILENAME`

Optional private SSH key to use when accessing the remote git repository.

Default: _not used_

### `SYNCER_SSH_KEY_PASSWORD`

Optional SSH key password.

Default: _not used_

## Changelog

* v7: Changed the health endpoint to `/health` instead of `/`.
* v8: Added support for cloning a tag and https git source URLs.
* v9: Modified Dockerfile to copy root certificates from the build container.

