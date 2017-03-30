// Copyright 2017 Bulat Gaifullin.  All rights reserved.
// Use of this source code is governed by a MIT license.

package main

import (
	"os"
	"strings"
)

// expand rewrites s to replace {k} with match[k] for each key k in match.
func expand(match map[string]string, s string) string {
	for k, v := range match {
		s = strings.Replace(s, "{"+k+"}", v, -1)
	}
	return s
}

// envForDir returns a copy of the environment
// suitable for running in the given directory.
// The environment is the current process's environment
// but with an updated $PWD, so that an os.Getwd in the
// child will be faster.
func envForDir(dir string) []string {
	env := os.Environ()
	// Internally we only use rooted paths, so dir is rooted.
	// Even if dir is not rooted, no harm done.
	return mergeEnvLists([]string{"PWD=" + dir}, env)
}

// mergeEnvLists merges the two environment lists such that
// variables with the same name in "in" replace those in "out".
func mergeEnvLists(in, out []string) []string {
NextVar:
	for _, inkv := range in {
		k := strings.SplitAfterN(inkv, "=", 2)[0]
		for i, outkv := range out {
			if strings.HasPrefix(outkv, k) {
				out[i] = inkv
				continue NextVar
			}
		}
		out = append(out, inkv)
	}
	return out
}
