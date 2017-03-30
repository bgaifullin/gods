// Copyright 2017 Bulat Gaifullin.  All rights reserved.
// Use of this source code is governed by a MIT license.

package main

import (
	"fmt"
	"path"
)

var cmdList = &Command{
	UsageLine: "list",
	Short:     "list packages in index",
	Long:      `Get downloads packages specified in configuration file.`,
	Run:       runList,
}

func runList(cmd *Command, configs *configsHierarchy, args []string) {
	for _, cfg := range configs.configs {
		fmt.Println(path.Base(cfg.file))
		for _, dep := range cfg.Deps {
			fmt.Printf("\t %s  %s  %s\n", dep.Package, dep.Url, dep.Version)
		}
	}
}
