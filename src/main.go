package main

import (
	"log"
	"os"
	"strings"
	"time"
)

var (
	VERSION = "v2"
)

func envString(name string, def string) string {
	val := os.Getenv(name)
	if len(val) == 0 {
		return def
	}
	return val
}

func envBool(name string, def bool) bool {
	val := strings.ToLower(envString(name, ""))
	if len(val) == 0 {
		return def
	}
	return val == "true" || val == "yes" || val == "on" || val == "1"
}

func initCommon(config *SyncerConfig) error {
	err := os.MkdirAll(config.Dest, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	log.Printf("syncer %s\n", VERSION)

	config, err := initSyncerConfig()
	if err != nil {
		log.Fatalf("Config error: %s\n", err)
	}

	var configFunc func(syncerConfig *SyncerConfig) error
	var initFunc func(syncerConfig *SyncerConfig) error
	var updateFunc func(syncerConfig *SyncerConfig) error

	switch config.Type {
	case "git":
		configFunc = gitConfig
		initFunc = gitInit
		updateFunc = gitUpdate
	default:
		log.Fatalf("Unhandled source type: %s\n", config.Source)
	}

	log.Printf("Configuring syncer...")
	err = configFunc(config)
	if err != nil {
		log.Fatalf("Error during config build: %s\n", err)
	}

	log.Printf("Initialising syncer...\n")
	log.Printf("              Type = %s\n", config.Type)
	log.Printf("            Source = %s\n", config.Source)
	log.Printf("              Dest = %s\n", config.Dest)
	log.Printf("   Update interval = %s\n", config.UpdateInterval.String())
	if config.Type == "git" {
		log.Printf("      Git upstream = %s\n", config.GitUpstream)
		log.Printf("        Git branch = %s\n", config.GitBranch)
		log.Printf("  SSH key filename = %s\n", config.GitSshKeyFilename)
	}

	err = initFunc(config)
	if err != nil {
		log.Fatalf("Error during initialisation: %s\n", err)
	}

	log.Printf("Iniitalisation complete.")

	updateTicker := time.NewTicker(config.UpdateInterval)
	for {
		select {
		case <-updateTicker.C:
			log.Printf("Updating...\n")
			err = updateFunc(config)
			if err != nil {
				log.Fatalf("Error during update: %s\n", err)
			}
			log.Printf("Update complete.")
		}
	}
}
