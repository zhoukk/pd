package main

import (
	"fmt"
	"net/http"
)

var httpCmd = &Command{
	UsageLine: "http [port]",
	Short:     "launch a web server for test",
	Long: `
launch a web server for test.

first args is port, default is :80.
`,
}

func init() {
	httpCmd.Run = httpApp
	AddCommand(httpCmd)
}

func httpApp(cmd *Command, args []string) {
	port := ":80"
	if len(args) > 0 {
		port = args[0]
	}
	fmt.Printf("http listen at %s\n", port)
	http.ListenAndServe(port, http.FileServer(http.Dir("./")))
}
