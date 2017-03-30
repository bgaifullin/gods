// Copyright 2017 Bulat Gaifullin.  All rights reserved.
// Use of this source code is governed by a MIT license.

package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
	"os"
)

type configsHierarchy struct {
	configs []*config
}

type config struct {
	Name    string        `yaml:"name"`
	Version int           `yaml:"version"`
	Deps    []*dependency `yaml:"dependencies"`
	index   map[string]*dependency
	file    string
}

type dependency struct {
	Package string `yaml:"package"`
	Version string `yaml:"version"`
	Url     string `yaml:"url"`
}

// Top returns the top level element in hierarchy
// which is associated with first element in GOPATH
func (h *configsHierarchy) Top() *config {
	if len(h.configs) > 0 {
		return h.configs[0]
	}
	return nil
}

// Append adds new config to hierarchy
func (h *configsHierarchy) Append(file string) error {
	cfg := new(config)
	err := cfg.Load(file)
	if err != nil && (os.IsPermission(err) || len(h.configs) > 0) {
		return err
	}
	if cfg.file == "" {
		cfg.file = file
		cfg.index = make(map[string]*dependency)
	}
	h.configs = append(h.configs, cfg)
	return nil
}

// Contains check that dependency already known
func (h *configsHierarchy) Contains(dep *dependency) bool {
	for _, cfg := range h.configs {
		if cfg.Contains(dep) {
			return true
		}
	}
	return false
}

// Load loads config from path
func (cfg *config) Load(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return err
	}
	cfg.rebuildIndex()
	cfg.file = file
	return nil
}

// Save saves changes in config to the same file
func (cfg *config) Save() error {
	return cfg.SaveTo(cfg.file)
}

// SaveTo saves config to specified file
func (cfg *config) SaveTo(file string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(file, data, 0644); err != nil {
		return err
	}
	cfg.file = file
	return nil
}

// Contains check that config contains dep
func (cfg *config) Contains(dep *dependency) bool {
	edep, ok := cfg.index[dep.Package]
	return ok && edep.Version == dep.Version
}

// Update merges deps into config and updates name if it is empty and increments version
// conflict may happen if same package with different version presents in both configs
func (cfg *config) Update(name string, deps []*dependency) error {
	for _, dep := range deps {
		if edep, ok := cfg.index[dep.Package]; ok {
			if edep.Version == dep.Version {
				continue
			}
			return fmt.Errorf(
				"Conflict: %s, existing - %s, new - %s", dep.Package, edep.Version, dep.Version,
			)
		}
		cfg.index[dep.Package] = dep
		cfg.Deps = append(cfg.Deps, dep)
	}
	if cfg.Name == "" {
		cfg.Name = name
	}
	cfg.Version++
	return nil
}

func (cfg *config) rebuildIndex() {
	index := make(map[string]*dependency)
	for _, dep := range cfg.Deps {
		index[dep.Package] = dep
	}
	cfg.index = index
}
