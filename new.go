package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

var newCmd = &Command{
	UsageLine: "new [postname]",
	Short:     "new a post",
	Long: `
new a markdown format post.

after then, edit the markdown file and compile it.
`,
}

type Mapper map[string]interface{}

func init() {
	newCmd.Run = newApp
	AddCommand(newCmd)
}

func newApp(cmd *Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "[ERRO] Argument [postname] is missing")
		os.Exit(2)
	}
	err := LoadConf("config.json")
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	filename := Root + "/posts/" + args[0] + ".md"
	_, err = os.Stat(filename)
	if err == nil || !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "post file %s already exist? \n", filename)
		return
	}
	err = os.MkdirAll(filepath.Dir(filename), os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s error:%s\n", filename, err.Error())
		return
	}
	title := args[0]
	t := time.Now()
	y, m, _ := t.Date()
	mdata := Mapper{}
	mdata["title"] = title
	mdata["date"] = t.Format("2006-01-02 15:04:05")
	mdata["permalink"] = fmt.Sprintf("/%d/%d/%s.html", y, m, title)
	mdata["description"] = ""
	mdata["category"] = "默认"
	b, err := json.MarshalIndent(mdata, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	err = ioutil.WriteFile(filename, b, os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	fmt.Printf("new post at :%s\n", filename)
}
