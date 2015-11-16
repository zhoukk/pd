package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"
)

var httpCmd = &Command{
	UsageLine: "http [config]",
	Short:     "launch a web server for test",
	Long: `
	launch a web server for test.

	first args is config file path.
	`,
}

type Comment struct {
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
	Url      string `json:"url"`
	Content  string `json:"content"`
	Time     string `json:"time"`
}

var comments map[string][]Comment

func init() {
	comments = make(map[string][]Comment, 0)
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
			err = t.ExecuteTemplate(w, "body", Mapper{"title": "存档", "categories": Categories, "config": Config})
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
	http.HandleFunc("/comment.list", func(w http.ResponseWriter, r *http.Request) {
		var data []Comment
		id := r.FormValue("id")
		data = comments[id]
		w.Header().Add("Content-Type", "application/json;charset=utf-8")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	http.HandleFunc("/comment.new", func(w http.ResponseWriter, r *http.Request) {
		var data Comment
		data.Id = r.FormValue("id")
		data.Nickname = r.FormValue("nickname")
		data.Url = r.FormValue("url")
		data.Content = r.FormValue("content")
		if len(data.Nickname) == 0 || len(data.Content) == 0 {
			return
		}
		data.Time = time.Now().Format("2006-01-02 15:04:05")
		data.Content = MarkdownToHtml(data.Content)
		comments[data.Id] = append(comments[data.Id], data)
		w.Header().Add("Access-Control-Allow-Origin", "*")
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
