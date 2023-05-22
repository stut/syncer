package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"log"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/mitchellh/go-homedir"
)

// gitConfig reads the git-specific configuration from the environment.
func gitConfig(config *SyncerConfig) error {
	var err error

	config.GitBranch = envString("SYNCER_GIT_BRANCH", "main")
	config.GitTag = envString("SYNCER_GIT_TAG", "")
	config.GitUpstream = envString("SYNCER_GIT_UPSTREAM", "origin")
	config.GitResetOnChange = envBool("SYNCER_GIT_RESET_ON_CHANGES", true)
	config.GitSshKeyFilename = envString("SYNCER_SSH_KEY_FILENAME", "")
	config.GitSshKeyPassword = envString("SYNCER_SSH_KEY_PASSWORD", "")

	if len(config.GitSshKeyFilename) > 0 {
		config.GitSshKeyFilename, err = homedir.Expand(config.GitSshKeyFilename)
		if err != nil {
			return fmt.Errorf("cannot perform homedir expansion on the SSH key filename: %s", config.GitSshKeyFilename)
		}
		if _, err := os.Stat(config.GitSshKeyFilename); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("SSH key filename does not exist: %s", config.GitSshKeyFilename)
		}
	}

	return nil
}

// getReferenceName creates a reference name for the tag or branch in the configuration.
func getReferenceName(config *SyncerConfig) plumbing.ReferenceName {
	if len(config.GitTag) > 0 {
		return plumbing.NewTagReferenceName(config.GitTag)
	}
	return plumbing.NewBranchReferenceName(config.GitBranch)
}

// gitInit performs the initial clone of a git source.
func gitInit(config *SyncerConfig) error {
	var err error

	err = initCommon(config)
	if err != nil {
		return err
	}

	isEmpty := false
	isEmpty, err = dirIsEmpty(config.Dest)
	if err != nil {
		return fmt.Errorf("cannot read from dest directory: %s", err)
	}
	if !isEmpty {
		err = checkGitConfigFile(config)
		if err != nil {
			return err
		}
		log.Printf("Clone already exists in dest dir, performing update instead...")
		return gitUpdate(config)
	}

	var publicKeys *ssh.PublicKeys
	publicKeys, err = getPublicKeys(config)
	if err != nil {
		return nil
	}

	log.Printf("Performing initial clone...")
	cloneOptions := &git.CloneOptions{
		URL:           config.Source,
		Progress:      log.Writer(),
		SingleBranch:  true,
		ReferenceName: getReferenceName(config),
	}
	if publicKeys != nil {
		cloneOptions.Auth = publicKeys
	}
	_, err = git.PlainClone(config.Dest, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to perform initial clone: %s", err)
	}

	return nil
}

// gitUpdate performs an update (pull) for a git source.
func gitUpdate(config *SyncerConfig) error {
	// We instantiate a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(config.Dest)
	if err != nil {
		if err.Error() == "repository does not exist" {
			log.Printf("  Repository does not exist, attempting reinitialisation...")
			return gitInit(config)
		}
		return fmt.Errorf("failed to open repo: %s", err)
	}

	var wt *git.Worktree
	wt, err = r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to open repo: %s", err)
	}

	var status git.Status
	status, err = wt.Status()
	if err != nil {
		return err
	}
	if !status.IsClean() {
		if config.GitResetOnChange {
			return gitReset(wt)
		}
		return fmt.Errorf("there are uncommitted changes, cannot pull")
	}

	var publicKeys *ssh.PublicKeys
	publicKeys, err = getPublicKeys(config)
	if err != nil {
		return err
	}

	pullOptions := &git.PullOptions{
		RemoteName:    config.GitUpstream,
		SingleBranch:  true,
		ReferenceName: getReferenceName(config),
		Progress:      log.Writer(),
	}
	if publicKeys != nil {
		pullOptions.Auth = publicKeys
	}
	err = wt.Pull(pullOptions)
	if err != nil && err.Error() != "already up-to-date" {
		return fmt.Errorf("pull failed: %s", err)
	}

	return nil
}

// gitReset resets a git clone.
func gitReset(wt *git.Worktree) error {
	log.Printf("  Performing hard reset...")
	return wt.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
}

// checkGitConfigFile checks an initial git clone to make sure it matches the configuration.
func checkGitConfigFile(config *SyncerConfig) error {
	// TODO: Use go-git functions to do this check
	gitConfigPath := path.Join(config.Dest, ".git", "config")
	if _, err := os.Stat(gitConfigPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("dest directory is not empty but does not contain a git clone")
	}

	file, err := os.Open(gitConfigPath)
	if err != nil {
		return fmt.Errorf("cannot open existing .git/config file in the dest dir: %s", err)
	}
	defer func() { _ = file.Close() }()

	inOriginSection := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), " \t")
		if !inOriginSection {
			if strings.HasPrefix(line, "[remote \"origin\"]") {
				inOriginSection = true
			}
		} else {
			if strings.HasPrefix(line, "url = ") {
				line = strings.TrimRight(line[len("url = "):], "\r\n \t")
				if line == config.Source {
					return nil
				}
				return fmt.Errorf("dest dir contains a git clone but the source doesn't match")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error while reading .git/config in the dest dir: %s", err)
	}

	return fmt.Errorf("dest dir contains a git clone but the source doesn't match")
}

// getPublicKeys reads the public keys from the private key file in the configuration.
func getPublicKeys(config *SyncerConfig) (*ssh.PublicKeys, error) {
	if len(config.GitSshKeyFilename) == 0 {
		return nil, nil
	}

	publicKeys, err := ssh.NewPublicKeysFromFile("git", config.GitSshKeyFilename, config.GitSshKeyPassword)
	if err != nil {
		return nil, fmt.Errorf("generate publickeys failed: %s", err)
	}
	return publicKeys, nil
}
