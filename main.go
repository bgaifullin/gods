// Copyright 2017 Bulat Gaifullin.  All rights reserved.
// Use of this source code is governed by a MIT license.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

const configFileName = ".gover.yaml"

// A Command is an implementation of a go command
// like go build or go fix.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	// return 0 on success or other code on fail
	Run func(cmd *Command, configs *configsHierarchy, args []string)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'go help' output.
	Short string

	// Long is the long message shown in the 'go help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet

	// CustomFlags indicates that the command will do its own
	// flag parsing.
	CustomFlags bool

	// Existing Configs
	Configs []*config
}

// Name returns the command's name: the first word in the usage line.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
}

// Commands lists the available commands and help topics.
// The order here is the order in which they are printed by 'go help'.
var commands = []*Command{
	cmdGet,
	cmdList,
}

func main() {
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	if args[0] == "help" {
		help(args[1:])
		return
	}

	gopath := os.Getenv("GOPATH")
	// Diagnose common mistake: GOPATH==GOROOT.
	// This setting is equivalent to not setting GOPATH at all,
	// which is not what most people want when they do it.
	if gopath == "" || gopath == runtime.GOROOT() {
		log.Fatal("GOPATH is empty or set to GOROOT. Please set GOPATH.")
	}
	configs := new(configsHierarchy)

	for _, p := range filepath.SplitList(gopath) {
		if strings.HasPrefix(p, "~") {
			log.Fatalf("gover: GOPATH entry cannot start with shell metacharacter '~': %q\n", p)
		}
		if strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") {
			log.Fatalf("gover: GOPATH entry is relative; must be absolute path: %q.\nRun 'go help gopath' for usage.\n", p)
		}
		cfgfile := path.Join(p, configFileName)
		if err := configs.Append(cfgfile); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: cannot read file '%s': %v\n", cfgfile, err)
		}
	}
	if configs.Top() == nil {
		log.Fatalln("gover: cannot load/create config file. please check permissions")
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] {
			cmd.Flag.Usage = func() { cmd.Usage() }
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}
			cmd.Run(cmd, configs, args)
			return
		}
	}

	log.Printf("unknown subcommand %q\nRun 'gover help' for usage.\n", args[0])
}

var usageTemplate = `gover is a tool for managing go dependencies snapshots.
Usage:
	gover command [arguments]
The commands are:
{{range .}}
    {{.Name | printf "%-11s"}} {{.Short}}
{{end}}
Use "gover help [command]" for more information about a command.
`

var helpTemplate = `usage: gover {{.UsageLine}}
{{.Long | trim}}
`

var documentationTemplate = `// Copyright 2017 Bulat Gaifullin. All rights reserved.
// Use of this source code is governed by MIT license.
/*
{{range .}}{{if .Short}}{{.Short | title}}{{end}}
Usage:
	gover {{.UsageLine}}
{{.Long | trim}}
{{end}}*/
package main
// NOTE: cmdDoc is in fmt.go.
`

// tmpl executes the given template text on data, writing the result to w.
func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace, "title": strings.Title})
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

func printUsage(w io.Writer) {
	tmpl(w, usageTemplate, commands)
}

func usage() {
	printUsage(os.Stderr)
	os.Exit(2)
}

// help implements the 'help' command.
func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		// not exit 2: succeeded at 'go help'.
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: gover help command\n\nToo many arguments given.\n")
		os.Exit(2) // failed at 'go help'
	}

	arg := args[0]

	// 'go help documentation' generates doc.go.
	if arg == "documentation" {
		buf := new(bytes.Buffer)
		printUsage(buf)
		usage := &Command{Long: buf.String()}
		tmpl(os.Stdout, documentationTemplate, append([]*Command{usage}, commands...))
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			// not exit 2: succeeded at 'go help cmd'.
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run 'go help'.\n", arg)
	os.Exit(2) // failed at 'go help cmd'
}
