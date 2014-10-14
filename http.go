package main

import (
	"fmt"
	"html/template"
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
	root := "./"
	if len(args) > 0 {
		port = args[0]
	}
	if len(args) > 1 {
		root = args[1]
	}
	LoadConf("config.json")
	LoadTheme()
	fmt.Printf("http listen at %s\n", port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-PJAX") == "true" {
			fmt.Println(r.URL.Path)
			t, _ := Tpl.Clone()
			t = template.Must(t.ParseFiles(Theme + "/post.html"))
			err := t.ExecuteTemplate(w, "body", Mapper{"config": Config})
			if err != nil {
				fmt.Println(err)
			}
			return
		}
		http.FileServer(http.Dir(root)).ServeHTTP(w, r)
	})
	http.ListenAndServe(port, nil)
}
