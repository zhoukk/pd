package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

var newCmd = &Command{
	UsageLine: "new [postname, path1, path2, ...]",
	Short:     "new a post",
	Long: `
new a markdown format post.

after then, edit the markdown file and compile it.
`,
}

const newTpl = `
	title : %s,
	data : '%s',
	categories : ,
	tags : 
`

func init() {
	newCmd.Run = newApp
	AddCommand(newCmd)
}

func newApp(cmd *Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "[ERRO] Argument [postname] is missing")
		os.Exit(2)
	}

	filename := "post/" + args[0] + ".md"
	_, err := os.Stat(filename)
	if err == nil || !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "post file %s already exist? \n", filename)
		return
	}
	err = ioutil.WriteFile(filename, []byte(fmt.Sprintf(newTpl, args[0], time.Now().Format("2006-01-02 15:04:05"))), os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	fmt.Printf("new post at :%s\n", filename)
}
