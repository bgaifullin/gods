// Copyright 2017 Bulat Gaifullin.  All rights reserved.
// Use of this source code is governed by a MIT license.

package main

import (
	"fmt"
	"log"
	"os"
	"path"
)

var cmdGet = &Command{
	UsageLine: "get <gover.yaml>",
	Short:     "download packages",
	Long:      `Get downloads packages specified in configuration file.`,
	Run:       runGet,
}

func runGet(cmd *Command, configs *configsHierarchy, args []string) {
	if len(args) == 0 {
		log.Fatal("missing config file")
	}
	localCfg := new(config)
	if err := localCfg.Load(args[0]); err != nil {
		log.Fatalf("cannot load configuration file '%s': %v", args[0], err)
	}

	missing := []*dependency{}
	for _, dep := range localCfg.Deps {
		if !configs.Contains(dep) {
			missing = append(missing, dep)
		}
	}
	if len(missing) == 0 {
		log.Println("everithing is up to date")
		return
	}
	top := configs.Top()
	if err := top.Update(localCfg.Name, missing); err != nil {
		log.Fatal(err)
	}

	log.Println(top.file)
	if err := download(path.Dir(top.file), missing); err != nil {
		log.Fatalf("cannot download packages: %v", err)
	}
	if err := top.Save(); err != nil {
		log.Fatalf("cannot save config: %v", err)
	}
}

func download(root string, deps []*dependency) error {
	for _, dep := range deps {
		dst := path.Join(root, "src", dep.Package)
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("package '%s' alredy exists and it is not in index. see gover help fix", dst)
		}
		if err := os.MkdirAll(path.Dir(dst), 0755); err != nil {
			return err
		}
		vcs := getVcsByUrl(dep.Url)
		if err := vcs.create(dst, dep.Url, dep.Version); err != nil {
			return err
		}
	}
	return nil
}
