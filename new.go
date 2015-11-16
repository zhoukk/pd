package main

import (
	"fmt"
	"os"
)

var newCmd = &Command{
	UsageLine: "new [sitename]",
	Short:     "new a site dir",
	Long: `
new a dir for site.

after then, you can compile it.
`,
}

func init() {
	newCmd.Run = newApp
	AddCommand(newCmd)
}

func newApp(cmd *Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "[ERRO] Argument [sitename] is missing")
		os.Exit(2)
	}
	err := os.MkdirAll(args[0], os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s error:%s\n", args[0], err.Error())
		return
	}
	fmt.Printf("new site at :%s\n", args[0])
}
