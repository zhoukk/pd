package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
)

var httpCmd = &Command{
	UsageLine: "http [port]",
	Short:     "launch a web server for test",
	Long: `
	launch a web server for test.

	default port is :80.
	`,
}

type Comment struct {
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
	Url      string `json:"url"`
	Content  string `json:"content"`
	Time     string `json:"time"`
}

type AllComment []Comment

func (p AllComment) Len() int {
	return len(p)
}

func (p AllComment) Less(i, j int) bool {
	p1 := p[i].Time
	p2 := p[j].Time
	pt1, err := time.Parse("2006-01-02 15:04:05", p1)
	if err != nil {
		log.Fatal(err)
	}
	pt2, err := time.Parse("2006-01-02 15:04:05", p2)
	if err != nil {
		log.Fatal(err)
	}
	return pt1.After(pt2)
}

func (p AllComment) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

var comments map[string]AllComment

func init() {
	comments = make(map[string]AllComment, 0)
	httpCmd.Run = httpApp
	AddCommand(httpCmd)
}

func httpApp(cmd *Command, args []string) {
	port := ":80"
	if len(args) > 0 {
		port = args[0]
	}
	http.HandleFunc("/comment.list", func(w http.ResponseWriter, r *http.Request) {
		var data AllComment
		id := r.FormValue("id")
		data = comments[id]
		sort.Sort(data)
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
	http.HandleFunc("/__a.gif", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.RawQuery)
		w.Header().Set("Content-Type", "image/gif")
		b, _ := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAP///wAAACwAAAAAAQABAAACAkQBADs=")
		w.Write(b)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(".")).ServeHTTP(w, r)
	})
	log.Fatal(http.ListenAndServe(port, nil))
}
