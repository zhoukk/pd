package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var postCmd = &Command{
	UsageLine: "post [filename] [title]",
	Short:     "create a new post",
	Long: `
create a markdown format post.

after then, edit the markdown file and compile it.
`,
}

type Mapper map[string]interface{}

func init() {
	postCmd.Run = postApp
	AddCommand(postCmd)
}

func create_post(name, title string) {
	filename := filepath.Join(".pd", "posts", name+".md")
	_, err := os.Stat(filename)
	if err == nil || !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "post file %s already exist? \n", filename)
		return
	}
	err = os.MkdirAll(filepath.Dir(filename), os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s error:%s\n", filename, err.Error())
		return
	}
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
	fmt.Printf("create file %s.\n", filename)
}

func postApp(cmd *Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "[ERRO] Argument [filename] is missing")
		os.Exit(2)
	}
	name := args[0]
	title := name
	if len(args) > 1 {
		title = args[1]
	}
	create_post(name, title)
}
