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
	GitUpstream       string
	GitResetOnChange  bool
	GitSshKeyFilename string
	GitSshKeyPassword string
}

func initSyncerConfig() (*SyncerConfig, error) {
	var err error

	source := envString("SYNCER_SOURCE", "")
	if len(source) == 0 {
		log.Fatalln("A SYNCER_SOURCE is required!")
	}

	res := &SyncerConfig{
		Type:              determineSourceType(source),
		Source:            source,
		Dest:              envString("SYNCER_DEST", ""),
		UpdateInterval:    0,
		GitSshKeyFilename: envString("SYNCER_SSH_KEY_FILENAME", ""),
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

func determineSourceType(source string) string {
	if strings.HasPrefix(source, "git@") {
		return "git"
	}
	return "unsupported"
}
