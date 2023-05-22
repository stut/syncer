package main

import (
	"log"
	"strings"
	"time"
)

type SyncerConfig struct {
	Type           string
	Dest           string
	Source         string
	UpdateInterval time.Duration

	// Git
	GitBranch         string
	GitTag            string
	GitUpstream       string
	GitResetOnChange  bool
	GitSshKeyFilename string
	GitSshKeyPassword string
}

// initSyncerConfig initialise the configuration from the environment.
func initSyncerConfig() (*SyncerConfig, error) {
	var err error

	source := envString("SYNCER_SOURCE", "")
	if len(source) == 0 {
		log.Fatalln("A SYNCER_SOURCE is required!")
	}

	res := &SyncerConfig{
		Type:           determineSourceType(source),
		Source:         source,
		Dest:           envString("SYNCER_DEST", ""),
		UpdateInterval: 0,
	}

	if len(res.Dest) == 0 {
		log.Fatalln("A SYNCER_DEST is required!")
	}

	res.UpdateInterval, err = time.ParseDuration(envString("SYNCER_UPDATE_INTERVAL", "1h"))
	if err != nil {
		return nil, err
	}

	return res, nil
}

// determineSourceType works out the source type from the source URL.
func determineSourceType(source string) string {
	if strings.HasPrefix(source, "git@") {
		return "git"
	}

	if strings.HasPrefix(source, "https://") && strings.HasSuffix(source, ".git") {
		return "git"
	}

	return "unsupported"
}
