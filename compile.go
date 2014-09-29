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

type Posts []Mapper

func (p Posts) Len() int {
	return len(p)
}

func (p Posts) Less(i, j int) bool {
	p1 := p[i]["date"].(string)
	p2 := p[j]["date"].(string)
	pt1, err := time.Parse("2006-01-02 15:04:05", p1)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
	pt2, err := time.Parse("2006-01-02 15:04:05", p2)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
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
	posts, err := loadPosts()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}

	for i, v := range posts {
		p, _ := createPost(v)
		posts[i] = p
	}
	createIndex(posts)
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
			fmt.Fprintf(os.Stderr, err.Error())
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

func createPost(post Mapper) (Mapper, error) {
	var buf bytes.Buffer
	link := "." + post["permalink"].(string)
	err := os.MkdirAll(filepath.Dir(link), os.ModePerm)
	if err != nil {
		return post, err
	}
	content := post["content"].(string)
	post["content"] = template.HTML(MarkdownToHtml(content))
	t, err := template.ParseFiles("./tpl/post.html", "./tpl/comment.html", "./tpl/navbar.html", "./tpl/footer.html")
	if err != nil {
		return post, err
	}
	err = t.Execute(&buf, post)
	if err != nil {
		return post, err
	}
	err = ioutil.WriteFile(link, buf.Bytes(), os.ModePerm)
	return post, err
}

func createIndex(posts Posts) error {
	var buf bytes.Buffer
	t, err := template.ParseFiles("./tpl/index.html", "./tpl/navbar.html", "./tpl/footer.html")
	if err != nil {
		return err
	}
	err = t.Execute(&buf, posts)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./index.html", buf.Bytes(), os.ModePerm)
	return err
}
