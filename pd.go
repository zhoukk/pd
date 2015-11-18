package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var usageTemplate = `pd is a static website generate tool.

Usage:
	pd command [arguments]

The commands are:
{{range .}}{{if .Runnable}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "pd help [command]" for more information about a command.

Additional help topics:
{{range .}}{{if not .Runnable}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "pd help [topic]" for more information about that topic.

`

var helpTemplate = `{{if .Runnable}}usage: pd {{.UsageLine}}

{{end}}{{.Long | trim}}`

type Command struct {
	Run         func(cmd *Command, args []string)
	UsageLine   string
	Short       string
	Long        string
	Flag        flag.FlagSet
	CustomFlags bool
}

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

func (c *Command) Runnable() bool {
	return c.Run != nil
}

var commands = []*Command{}

func AddCommand(cmd *Command) {
	commands = append(commands, cmd)
}

func zip_pd() {
	var b bytes.Buffer
	myzip := zip.NewWriter(&b)
	err := filepath.Walk(".pd", func(path string, info os.FileInfo, err error) error {
		var file []byte
		if err != nil {
			return filepath.SkipDir
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return filepath.SkipDir
		}
		p, _ := filepath.Rel(filepath.Dir(".pd"), path)
		header.Name = strings.Replace(p, "\\", "/", -1)
		if !info.IsDir() {
			header.Method = 8
			file, err = ioutil.ReadFile(path)
			if err != nil {
				return filepath.SkipDir
			}
		} else {
			file = nil
		}
		w, err := myzip.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = w.Write(file)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	myzip.Close()
	var bb bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &bb)
	encoder.Write(b.Bytes())
	encoder.Close()
	content := fmt.Sprintf("package main\nvar ZipInitStr string = \"%s\"", bb.String())
	err = ioutil.WriteFile("zip.go", []byte(content), os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
}

func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

func usage() {
	tmpl(os.Stdout, usageTemplate, commands)
	os.Exit(2)
}

func help(args []string) {
	if len(args) == 0 {
		usage()
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stdout, "usage: pd help command\n\nToo many arguments given.\n")
		os.Exit(2)
	}

	arg := args[0]

	for _, cmd := range commands {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			return
		}
	}

	fmt.Fprintf(os.Stdout, "Unknown help topic %#q.  Run 'pd help'.\n", arg)
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}
	if args[0] == "help" {
		help(args[1:])
		return
	}
	if args[0] == "zip" {
		zip_pd()
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}
			cmd.Run(cmd, args)
			os.Exit(2)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "pd: unknown subcommand %q\nRun 'pd help' for usage.\n", args[0])
	os.Exit(2)
}
