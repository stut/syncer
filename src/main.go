package main

import (
	"log"
	"os"
	"time"
)

type SyncerConfig struct {
	Type           string
	Dest           string
	Source         string
	UpdateInterval time.Duration

	// Git
	GitUpstream    string
	SshKeyFilename string
	SshKeyPassword string
}

func envString(name string, def string) string {
	val := os.Getenv(name)
	if len(val) == 0 {
		return def
	}
	return val
}

func initCommon(config *SyncerConfig) error {
	err := os.MkdirAll(config.Dest, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	config, err := initSyncerConfig()
	if err != nil {
		log.Fatalf("Config error: %s\n", err)
	}

	var initFunc func(syncerConfig *SyncerConfig) error
	var updateFunc func(syncerConfig *SyncerConfig) error

	switch config.Type {
	case "git":
		initFunc = gitInit
		updateFunc = gitUpdate
	default:
		log.Fatalf("Unhandled source type: %s\n", config.Source)
	}

	log.Printf("Initialising syncer...\n")
	log.Printf("              Type = %s\n", config.Type)
	log.Printf("            Source = %s\n", config.Source)
	log.Printf("              Dest = %s\n", config.Dest)
	log.Printf("   Update interval = %s\n", config.UpdateInterval.String())
	if config.Type == "git" {
		log.Printf("      Git upstream = %s\n", config.GitUpstream)
		log.Printf("  SSH key filename = %s\n", config.SshKeyFilename)
	}

	err = initCommon(config)
	if err != nil {
		log.Fatalf("Error during common initialisation: %s\n", err)
	}
	err = initFunc(config)
	if err != nil {
		log.Fatalf("Error during initialisation: %s\n", err)
	}
	log.Printf("Initialisation complete.")

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
