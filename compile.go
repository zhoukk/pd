package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var compileCmd = &Command{
	UsageLine: "compile",
	Short:     "compile whole website",
	Long: `
compile all markdown file in post.

compile markdown file to html file.
`,
}

type Config struct {
	Author string `json:"author"`
	Theme  string `json:"theme"`
	Title  string `json:"title"`
}

var (
	config Config
	theme  Mapper
)

type Posts []Mapper

func (p Posts) Len() int {
	return len(p)
}

func (p Posts) Less(i, j int) bool {
	p1 := p[i]["date"].(string)
	p2 := p[j]["date"].(string)
	pt1, err := time.Parse("2006-01-02 15:04:05", p1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	pt2, err := time.Parse("2006-01-02 15:04:05", p2)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	return pt1.After(pt2)
}

func (p Posts) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func init() {
	compileCmd.Run = compileApp
	AddCommand(compileCmd)
}

func compileApp(cmd *Command, args []string) {
	err := loadConf()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	err = loadTheme()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	posts, err := loadPosts()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	var prev, next *Mapper
	for i, v := range posts {
		if i > 0 {
			prev = &posts[i-1]
		} else {
			prev = nil
		}
		if i < len(posts)-1 {
			next = &posts[i+1]
		} else {
			next = nil
		}
		p, err := createPost(v, prev, next)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}
		posts[i] = p
	}
	if err := createIndex(posts); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

func loadConf() error {
	f, err := os.Open("config.json")
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return err
	}
	return nil
}

func loadTheme() error {
	f, err := os.Open("./theme/" + config.Theme + "/theme.json")
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&theme)
	if err != nil {
		return err
	}
	return nil
}

func loadPosts() (Posts, error) {
	posts := Posts{}

	err := filepath.Walk("./post/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		post, err := loadPost(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return err
		}
		post["path"] = path
		posts = append(posts, post)
		return nil
	})

	sort.Sort(posts)
	return posts, err
}

func loadPost(path string) (post Mapper, err error) {
	post = make(Mapper, 0)
	content, err := ioutil.ReadFile(path)
	n := strings.IndexRune(string(content), '}')
	if n == -1 {
		err = errors.New("error format of post header")
		return
	}
	head := content[:n+1]
	err = json.Unmarshal(head, &post)
	if err != nil {
		return
	}
	post["content"] = string(content[n+1:])
	return
}

func createPost(post Mapper, prev, next *Mapper) (Mapper, error) {
	var buf bytes.Buffer
	link := "." + post["permalink"].(string)
	err := os.MkdirAll(filepath.Dir(link), os.ModePerm)
	if err != nil {
		return post, err
	}
	content := post["content"].(string)
	post["content"] = template.HTML(MarkdownToHtml(content))

	var files []string
	for _, v := range theme["post"].([]interface{}) {
		files = append(files, "./theme/"+config.Theme+"/"+v.(string))
	}
	t, err := template.ParseFiles(files...)
	if err != nil {
		return post, err
	}
	if prev != nil {
		post["previous_url"] = (*prev)["permalink"]
		post["previous_title"] = (*prev)["title"]
	}
	if next != nil {
		post["next_url"] = (*next)["permalink"]
		post["next_title"] = (*next)["title"]
	}
	ti, err := time.Parse("2006-01-02 15:04:05", post["date"].(string))
	if err != nil {
		return post, err
	}
	post["sdate"] = ti.Format("2006-01-02")
	err = t.Execute(&buf, Mapper{"post": post, "config": config, "title": post["title"]})
	if err != nil {
		return post, err
	}
	err = ioutil.WriteFile(link, buf.Bytes(), os.ModePerm)
	return post, err
}

func createIndex(posts Posts) error {
	var buf bytes.Buffer
	var files []string
	for _, v := range theme["index"].([]interface{}) {
		files = append(files, "./theme/"+config.Theme+"/"+v.(string))
	}
	t, err := template.ParseFiles(files...)
	if err != nil {
		return err
	}
	err = t.Execute(&buf, Mapper{"posts": posts, "config": config, "title": config.Title})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./index.html", buf.Bytes(), os.ModePerm)
	return err
}
