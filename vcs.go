// Copyright 2017 Bulat Gaifullin.  All rights reserved.
// Use of this source code is governed by a MIT license.

package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
)

// A vcsCmd describes how to use a version control system
// like Mercurial, Git, or Subversion.
type vcsCmd struct {
	name string
	cmd  string // name of binary to invoke command

	createCmd   string // command to download a fresh copy of a repository
	downloadCmd string // command to download updates into an existing repository
	checkoutCmd string
}

// vcsList lists the known version control systems
var vcsList = []*vcsCmd{
	vcsGit,
}

// vcsByCmd returns the version control system for the given
// command name (hg, git, svn, bzr).
func vcsByCmd(cmd string) *vcsCmd {
	for _, vcs := range vcsList {
		if vcs.cmd == cmd {
			return vcs
		}
	}
	return nil
}

// vcsGit describes how to use Git.
var vcsGit = &vcsCmd{
	name: "Git",
	cmd:  "git",

	createCmd:   "clone {repo} {dir} -b {branch}",
	downloadCmd: "fetch",
	checkoutCmd: "checkout {commit}",
}

func getVcsByUrl(url string) *vcsCmd {
	// there is no other vcs except git
	return vcsGit
}

// create creates a new copy of repo in dir.
// The parent of dir must exist; dir must not.
func (v *vcsCmd) create(dir, repo string, version string) error {
	tag := "master"
	commit := ""
	if strings.HasPrefix(version, "sha:") {
		commit = version[4:]
	} else {
		tag = version
	}
	if err := v.run(".", v.createCmd, "dir", dir, "repo", repo, "branch", tag); err != nil {
		return err
	}
	if commit != "" {
		return v.run(dir, v.checkoutCmd, "commit", commit)
	}
	return nil
}

// download downloads any new changes for the repo in dir.
func (v *vcsCmd) download(dir string) error {
	return v.run(dir, v.downloadCmd)
}

// run runs the command line cmd in the given directory.
// keyval is a list of key, value pairs.  run expands
// instances of {key} in cmd into value, but only after
// splitting cmd into individual arguments.
// If an error occurs, run prints the command line and the
// command's combined stdout+stderr to standard error.
// Otherwise run discards the command's output.
func (v *vcsCmd) run(dir string, cmd string, keyval ...string) error {
	_, err := v.run1(dir, cmd, keyval, true)
	return err
}

// runVerboseOnly is like run but only generates error output to standard error in verbose mode.
func (v *vcsCmd) runVerboseOnly(dir string, cmd string, keyval ...string) error {
	_, err := v.run1(dir, cmd, keyval, false)
	return err
}

// runOutput is like run but returns the output of the command.
func (v *vcsCmd) runOutput(dir string, cmd string, keyval ...string) ([]byte, error) {
	return v.run1(dir, cmd, keyval, true)
}

// run1 is the generalized implementation of run and runOutput.
func (v *vcsCmd) run1(dir string, cmdline string, keyval []string, verbose bool) ([]byte, error) {
	m := make(map[string]string)
	for i := 0; i < len(keyval); i += 2 {
		m[keyval[i]] = keyval[i+1]
	}
	args := strings.Fields(cmdline)
	for i, arg := range args {
		args[i] = expand(m, arg)
	}

	_, err := exec.LookPath(v.cmd)
	if err != nil {
		log.Printf("gods: missing %s command. See http://golang.org/s/gogetcmd\n", v.name)
		return nil, err
	}

	cmd := exec.Command(v.cmd, args...)
	cmd.Dir = dir
	cmd.Env = envForDir(cmd.Dir)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	out := buf.Bytes()
	if err != nil {
		if verbose {
			log.Printf("# cd %s; %s %s\n", dir, v.cmd, strings.Join(args, " "))
			os.Stderr.Write(out)
		}
		return nil, err
	}
	return out, nil
}
