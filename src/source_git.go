package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"log"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/mitchellh/go-homedir"
)

func gitInit(config *SyncerConfig) error {
	var err error
	config.SshKeyFilename, err = homedir.Expand(config.SshKeyFilename)
	if err != nil {
		return fmt.Errorf("cannot perform homedir expansion on the SSH key filename: %s", config.SshKeyFilename)
	}
	if _, err := os.Stat(config.SshKeyFilename); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("SSH key filename does not exist: %s", config.SshKeyFilename)
	}

	config.SshKeyPassword = envString("SYNCER_SSH_KEY_PASSWORD", "")
	config.GitUpstream = envString("SYNCER_GIT_UPSTREAM", "origin")

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
		URL:          config.Source,
		Auth:         publicKeys,
		Progress:     os.Stdout,
		SingleBranch: true,
		Depth:        1,
	}
	_, err = git.PlainClone(config.Dest, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to perform initial clone: %s", err)
	}
	return nil
}

func gitUpdate(config *SyncerConfig) error {
	// We instantiate a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(config.Dest)
	if err != nil {
		return fmt.Errorf("failed to open repo: %s", err)
	}

	var wt *git.Worktree
	wt, err = r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to open repo: %s", err)
	}

	var publicKeys *ssh.PublicKeys
	publicKeys, err = getPublicKeys(config)
	if err != nil {
		return err
	}

	pullOptions := &git.PullOptions{
		RemoteName:   config.GitUpstream,
		SingleBranch: true,
		Depth:        1,
		Auth:         publicKeys,
		Progress:     os.Stdout,
	}
	err = wt.Pull(pullOptions)
	if err != nil && err.Error() != "already up-to-date" {
		return fmt.Errorf("pull failed: %s", err)
	}

	return nil
}

func checkGitConfigFile(config *SyncerConfig) error {
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

func getPublicKeys(config *SyncerConfig) (*ssh.PublicKeys, error) {
	if len(config.SshKeyPassword) > 0 {
		publicKeys, err := ssh.NewPublicKeysFromFile("git", config.SshKeyFilename, config.SshKeyPassword)
		if err != nil {
			return nil, fmt.Errorf("generate publickeys failed: %s", err)
		}
		return publicKeys, nil
	}
	return nil, nil
}

