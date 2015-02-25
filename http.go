package main

import (
	"html/template"
	"log"
	"net/http"
)

var httpCmd = &Command{
	UsageLine: "http [config]",
	Short:     "launch a web server for test",
	Long: `
	launch a web server for test.

	first args is config file path.
	`,
}

func init() {
	httpCmd.Run = httpApp
	AddCommand(httpCmd)
}

func httpApp(cmd *Command, args []string) {
	config_file := "config.json"
	if len(args) > 0 {
		config_file = args[0]
	}
	err := LoadConf(config_file)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	err = LoadTheme()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	err = LoadPosts()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	for i, v := range Posts {
		var err error
		Posts[i], err = CreatePost(v, i)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}
	port := Config["port"].(string)
	root := Config["root"].(string)
	log.Printf("http listen at %s, root:%s\n", port, root)
	http.HandleFunc("/archive.html", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-PJAX") == "true" {
			t, err := Tpl.Clone()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t = template.Must(t.ParseFiles(Theme + "/archive.html"))
			err = t.ExecuteTemplate(w, "body", Mapper{"title": "存档", "tags": Tags, "config": Config})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		http.FileServer(http.Dir(root)).ServeHTTP(w, r)
	})
	http.HandleFunc("/about.html", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-PJAX") == "true" {
			t, err := Tpl.Clone()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t = template.Must(t.ParseFiles(Theme + "/about.html"))
			err = t.ExecuteTemplate(w, "body", Mapper{"title": "关于", "config": Config})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		http.FileServer(http.Dir(root)).ServeHTTP(w, r)
	})
	http.HandleFunc("/msgboard.html", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-PJAX") == "true" {
			t, err := Tpl.Clone()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t = template.Must(t.ParseFiles(Theme + "/msgboard.html"))
			err = t.ExecuteTemplate(w, "body", Mapper{"title": "留言板", "config": Config})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		http.FileServer(http.Dir(root)).ServeHTTP(w, r)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-PJAX") == "true" {
			log.Println(r.URL.Path)
			if r.URL.Path == "/" {
				t, err := Tpl.Clone()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				t = template.Must(t.ParseFiles(Theme + "/index.html"))
				err = t.ExecuteTemplate(w, "body", Mapper{"title": Config["title"], "posts": Posts, "config": Config})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			} else {
				for _, v := range Posts {
					if r.URL.Path == v["permalink"] {
						t, _ := Tpl.Clone()
						t = template.Must(t.ParseFiles(Theme + "/post.html"))
						err := t.ExecuteTemplate(w, "body", Mapper{"title": v["title"], "post": v, "config": Config})
						if err != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)
						}
						break
					}
				}
			}
			return
		}
		http.FileServer(http.Dir(root)).ServeHTTP(w, r)
	})
	http.ListenAndServe(port, nil)
}
