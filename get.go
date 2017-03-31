// Copyright 2017 Bulat Gaifullin.  All rights reserved.
// Use of this source code is governed by a MIT license.

package main

import (
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
	var err error
	for _, dep := range deps {
		dst := path.Join(root, "src", dep.Package)
		vcs := getVcsByUrl(dep.Url)
		if vcs.exists(dst) {
			log.Printf("warning: unmanaged repository '%s'. reset version\n", dst)
			err = vcs.checkout(dst, dep.Version)
		} else {
			log.Printf("create new repository '%s'\n", dst)
			if err = os.MkdirAll(path.Dir(dst), 0755); err == nil {
				err = vcs.create(dst, dep.Url, dep.Version)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}
